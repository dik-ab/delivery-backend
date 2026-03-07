package model

import "time"

// Trip type constants
const (
	TripTypeOutbound = "outbound" // 往路
	TripTypeReturn   = "return"   // 復路（空車・帰り便）
)

// Trip represents a transport company's trip (delivery route)
type Trip struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	DriverID           uint       `json:"driver_id"`
	Driver             User       `json:"driver" gorm:"foreignKey:DriverID" binding:"-"`
	OriginAddress      string     `json:"origin_address" binding:"required"`
	OriginLat          float64    `json:"origin_lat"`
	OriginLng          float64    `json:"origin_lng"`
	DestinationAddress string     `json:"destination_address" binding:"required"`
	DestinationLat     float64    `json:"destination_lat"`
	DestinationLng     float64    `json:"destination_lng"`
	DepartureAt        time.Time  `json:"departure_at"`
	EstimatedArrival   *time.Time `json:"estimated_arrival"`
	VehicleType        string     `json:"vehicle_type"`
	AvailableWeight    float64    `json:"available_weight"`                         // kg
	Price              int        `json:"price"`                                    // 円
	Status             string     `json:"status" gorm:"default:open"`              // open, matched, in_transit, completed, cancelled
	TripType           string     `json:"trip_type" gorm:"default:outbound"`       // outbound（往路）, return（帰り便・空車）
	IsPublic           bool       `json:"is_public" gorm:"default:false"`          // 外部公開するか
	IsSoloMode         bool       `json:"is_solo_mode" gorm:"default:false"`       // ソロモード（自社便情報のみ）
	Note               string     `json:"note"`
	DelayMinutes       int        `json:"delay_minutes" gorm:"default:0"`          // 渋滞遅延（分）
	RoutePolyline      string     `json:"route_polyline" gorm:"type:text"`         // エンコード済みポリライン
	RouteDurationSec   int        `json:"route_duration_sec" gorm:"default:0"`     // ルート全体の所要時間（秒）
	RouteStepsJSON     string     `json:"route_steps_json" gorm:"type:mediumtext"` // ステップごとの情報（JSON）
	Matches            []Match    `json:"matches" gorm:"foreignKey:TripID" binding:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Trip) TableName() string {
	return "trips"
}
