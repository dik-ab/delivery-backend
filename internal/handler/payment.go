package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/delivery-app/delivery-api/internal/model"
	"github.com/delivery-app/delivery-api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"
)

// PaymentHandler handles payment-related requests
type PaymentHandler struct {
	paymentRepo *repository.PaymentRepository
	matchRepo   *repository.MatchRepository
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentRepo *repository.PaymentRepository, matchRepo *repository.MatchRepository) *PaymentHandler {
	// Set Stripe secret key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &PaymentHandler{
		paymentRepo: paymentRepo,
		matchRepo:   matchRepo,
	}
}

// CreatePaymentIntentRequest represents a payment intent creation request
type CreatePaymentIntentRequest struct {
	MatchID uint `json:"match_id" binding:"required"`
	Amount  int  `json:"amount" binding:"required,min=50"` // 最低50円（Stripe JPY最低額）
}

// CreatePaymentIntent godoc
// @Summary Create a Stripe Payment Intent
// @Description Create a payment intent for a matched trip
// @Tags payments
// @Accept json
// @Produce json
// @Param request body CreatePaymentIntentRequest true "Payment request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payments/create-intent [post]
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Verify match exists and is approved
	match, err := h.matchRepo.GetByID(req.MatchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	if match.Status != model.MatchStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "マッチングが承認されていません"})
		return
	}

	// Create Stripe Payment Intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(req.Amount)),
		Currency: stripe.String("jpy"),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Metadata: map[string]string{
			"match_id": strconv.Itoa(int(req.MatchID)),
			"payer_id": strconv.Itoa(int(userID.(uint))),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe Payment Intent の作成に失敗しました: " + err.Error()})
		return
	}

	// Save payment record
	payment := &model.Payment{
		MatchID:         req.MatchID,
		PayerID:         userID.(uint),
		Amount:          req.Amount,
		Currency:        "jpy",
		StripePaymentID: pi.ID,
		Status:          "pending",
	}

	if err := h.paymentRepo.Create(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "決済記録の保存に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment_id":    payment.ID,
		"client_secret": pi.ClientSecret, // フロントエンドに渡してStripe.jsで決済確定
		"amount":        req.Amount,
		"currency":      "jpy",
	})
}

// GetPaymentsByMatch godoc
// @Summary Get payments for a match
// @Description Get payment records for a specific match
// @Tags payments
// @Accept json
// @Produce json
// @Param match_id path int true "Match ID"
// @Success 200 {array} model.Payment
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payments/match/{match_id} [get]
func (h *PaymentHandler) GetPaymentsByMatch(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("match_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match ID"})
		return
	}

	payments, err := h.paymentRepo.GetByMatchID(uint(matchID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "決済情報の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// ConfirmPayment godoc
// @Summary Confirm a payment (webhook simulation for learning)
// @Description Manually confirm a payment status (in production, use Stripe Webhooks)
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} model.Payment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/payments/{id}/confirm [put]
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	payment, err := h.paymentRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	payment.Status = "succeeded"
	if err := h.paymentRepo.Update(uint(id), payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "決済ステータスの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// HandleWebhook godoc
// @Summary Handle Stripe webhook events
// @Description Receive and process Stripe webhook events (payment_intent.succeeded, payment_intent.payment_failed)
// @Tags payments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/webhook/stripe [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	// Stripe Webhook のリクエストボディサイズ制限（Stripe推奨: 最大64KB）
	const maxBodyBytes = int64(65536)
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストボディの読み取りに失敗しました"})
		return
	}

	// Webhook Signing Secret で署名を検証
	// Stripe Dashboard → Webhooks → Signing secret でコピーした値を環境変数に設定
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	var event stripe.Event

	if endpointSecret != "" {
		// 本番: 署名検証あり（推奨）
		event, err = webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), endpointSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook署名の検証に失敗しました: " + err.Error()})
			return
		}
	} else {
		// 開発環境: 署名検証なし（STRIPE_WEBHOOK_SECRET未設定時）
		if err := json.Unmarshal(body, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhookイベントのパースに失敗しました"})
			return
		}
	}

	// イベントタイプに応じて処理
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PaymentIntentのパースに失敗しました"})
			return
		}

		// DB上の決済レコードを更新
		payment, err := h.paymentRepo.GetByStripePaymentID(pi.ID)
		if err != nil {
			// 該当する決済レコードが見つからない場合はログだけ出して200返す
			c.JSON(http.StatusOK, gin.H{"message": "payment record not found, skipped"})
			return
		}

		payment.Status = "succeeded"
		if err := h.paymentRepo.Update(payment.ID, payment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "決済ステータスの更新に失敗しました"})
			return
		}

		// マッチングステータスも完了に更新
		if matchIDStr, ok := pi.Metadata["match_id"]; ok {
			matchID, err := strconv.ParseUint(matchIDStr, 10, 32)
			if err == nil {
				match, err := h.matchRepo.GetByID(uint(matchID))
				if err == nil {
					match.Status = model.MatchStatusCompleted
					h.matchRepo.Update(uint(matchID), match)
				}
			}
		}

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PaymentIntentのパースに失敗しました"})
			return
		}

		// DB上の決済レコードを失敗に更新
		payment, err := h.paymentRepo.GetByStripePaymentID(pi.ID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "payment record not found, skipped"})
			return
		}

		payment.Status = "failed"
		if err := h.paymentRepo.Update(payment.ID, payment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "決済ステータスの更新に失敗しました"})
			return
		}

	default:
		// 他のイベントタイプは無視して200を返す
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
