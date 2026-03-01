# Go プログラミング言語 - JavaScript/TypeScript 開発者向けガイド

このガイドは、JavaScript/TypeScript の経験がある開発者が Go を学ぶためのドキュメントです。このプロジェクトで使用する Go の概念を、TypeScript との比較を交えながら詳しく説明します。

---

## 目次

1. [Go vs JavaScript/TypeScript の違い](#1-go-vs-javascripttypescript-の違い)
2. [Go の基本構文](#2-go-の基本構文)
3. [このプロジェクトで使う Go の機能](#3-このプロジェクトで使う-go-の機能)
4. [Gin フレームワーク](#4-gin-フレームワーク)
5. [GORM](#5-gorm)

---

## 1. Go vs JavaScript/TypeScript の違い

### 1-1. 型システム（静的型付け）

JavaScript は動的型付けですが、Go は静的型付けです。これは大きな違いです。

**JavaScript（動的型付け）:**
```javascript
function add(a, b) {
  return a + b;
}

add(5, 3);           // OK: 8
add("5", "3");       // OK: "53"
add(5, "3");         // OK: "53"（型の自動変換）
add({}, []);         // OK: "[object Object]"（予期しない結果）
```

型の自動変換により、予期しないバグが発生する可能性があります。

**TypeScript（静的型付け）:**
```typescript
function add(a: number, b: number): number {
  return a + b;
}

add(5, 3);           // OK: 8
add("5", "3");       // エラー: Argument of type 'string' is not assignable to parameter of type 'number'
add({}, []);         // エラー
```

型が明示されているため、エラーを事前に検出できます。

**Go（静的型付け）:**
```go
func Add(a int, b int) int {
  return a + b
}

Add(5, 3)            // OK: 8
Add("5", "3")        // コンパイルエラー: cannot use "5" (type untyped string) as type int
Add(int64(5), 3)     // エラー: cannot use int64(5) (type int64) as type int
```

Go では型チェックがコンパイル時に行われます。TypeScript との大きな違いは、**型の自動変換がほとんど行われない**ということです。

**型の明示的な変換（型キャスト）:**
```go
var a int32 = 5
var b int64 = int64(a)   // 型キャストが必要

var str string = "hello"
var num int = 42
result := str + num      // エラー: invalid operation: str + num (type string, type int)
result := str + strconv.Itoa(num)  // 型変換後に連結
```

**比較:**

| 項目 | JavaScript | TypeScript | Go |
|------|-----------|-----------|-----|
| 型チェック | 実行時 | コンパイル時 | コンパイル時 |
| 型の自動変換 | あり（多い） | なし | なし |
| 型キャスト | 暗黙的 | 明示的 | 明示的 |

---

### 1-2. コンパイル言語 vs インタプリタ言語

**JavaScript:**
- インタプリタ言語
- コードをそのまま実行
- デプロイが簡単（`node app.js` で実行）
- 実行時にエラーが発見される

```javascript
// app.js
console.log("Hello");
console.log(someUndefinedVariable);  // 実行時にエラー
```

```bash
$ node app.js
Hello
ReferenceError: someUndefinedVariable is not defined
```

**Go:**
- コンパイル言語
- コンパイルして実行ファイル（バイナリ）を生成してから実行
- デプロイが簡単（単一のバイナリファイルをコピーするだけ）
- コンパイル時にエラーが発見される

```go
// main.go
package main

import "fmt"

func main() {
  fmt.Println("Hello")
  fmt.Println(someUndefinedVariable)  // コンパイルエラー
}
```

```bash
$ go build -o app main.go
./main.go:8:3: undefined: someUndefinedVariable
```

**比較:**

| 項目 | JavaScript | Go |
|------|-----------|-----|
| 処理 | インタプリタ | コンパイル |
| 実行方法 | `node app.js` | `./app`（バイナリ）または `go run main.go` |
| エラー検出 | 実行時 | コンパイル時 |
| パフォーマンス | 遅い | 高速 |
| デプロイ | 実行環境が必要（Node.js） | 単一バイナリ |

---

### 1-3. パッケージシステム

**JavaScript/TypeScript:**

```javascript
// utils.js（モジュールの定義）
function greet(name) {
  return `Hello, ${name}`;
}

module.exports = { greet };  // CommonJS
// または
export { greet };             // ES6 modules
```

```javascript
// app.js（モジュールの使用）
const { greet } = require('./utils');  // CommonJS
// または
import { greet } from './utils.js';    // ES6 modules

console.log(greet("World"));
```

**Go:**

```go
// utils.go
package utils

// 関数名が大文字で始まる = 外部パッケージからアクセス可能（public）
func Greet(name string) string {
  return "Hello, " + name
}

// 関数名が小文字で始まる = パッケージ内のみアクセス可能（private）
func privateFunc() string {
  return "private"
}
```

```go
// main.go
package main

import (
  "fmt"
  "myproject/utils"  // パッケージをインポート
)

func main() {
  result := utils.Greet("World")  // 大文字で始まる関数のみ呼び出し可能
  fmt.Println(result)
}
```

**重要なポイント:**

- **大文字 = Public**: 関数・変数・定数が大文字で始まる場合、他のパッケージからアクセス可能
- **小文字 = Private**: 小文字で始まる場合、同じパッケージ内のみアクセス可能

```go
// package.go
package mypackage

var PublicVar = "accessible"       // 外部からアクセス可能
var privateVar = "not accessible"  // パッケージ内のみ

func PublicFunc() {}               // 外部からアクセス可能
func privateFunc() {}              // パッケージ内のみ
```

---

### 1-4. エラーハンドリング（try-catch がない）

**JavaScript/TypeScript（try-catch）:**

```typescript
async function fetchUser(id: number) {
  try {
    const response = await fetch(`/api/users/${id}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const user = await response.json();
    return user;
  } catch (error) {
    console.error("Error fetching user:", error);
    return null;
  }
}
```

**Go（エラーを戻り値として返す）:**

```go
package main

import (
  "errors"
  "fmt"
  "io"
  "net/http"
)

func FetchUser(id int) (string, error) {
  resp, err := http.Get(fmt.Sprintf("http://api/users/%d", id))
  if err != nil {
    return "", err  // エラーを戻り値として返す
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return "", errors.New("HTTP error: status " + fmt.Sprint(resp.StatusCode))
  }

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }

  return string(body), nil  // 正常系は (データ, nil) を返す
}

func main() {
  user, err := FetchUser(1)
  if err != nil {
    fmt.Println("Error:", err)
    return
  }
  fmt.Println("User:", user)
}
```

**Go のエラーハンドリングの特徴:**

1. **複数戻り値**: 関数が複数の値を返す。通常は `(結果, エラー)` の形式
2. **エラーチェック**: `if err != nil` でエラーを確認
3. **例外がない**: `panic` と `recover` という機構はあるが、通常は使わない

**パターン:**

```go
// パターン1: エラー処理
result, err := someFunction()
if err != nil {
  // エラー処理
  return err
}
// 結果を使う
fmt.Println(result)

// パターン2: エラーを無視（非推奨）
result, _ := someFunction()
fmt.Println(result)

// パターン3: エラーのみ処理
err := someFunction2()
if err != nil {
  // エラー処理
  return err
}
```

**複数の戻り値の例:**

```go
// 関数の定義
func Divide(a, b float64) (float64, error) {
  if b == 0 {
    return 0, errors.New("division by zero")
  }
  return a / b, nil
}

// 使用
result, err := Divide(10, 2)
if err != nil {
  fmt.Println("Error:", err)
} else {
  fmt.Println("Result:", result)  // 5.0
}
```

---

## 2. Go の基本構文

### 2-1. 変数宣言

**Go には複数の宣言方法があります。**

#### 方法1: `var` キーワード

```go
// 型を明示
var name string = "Taro"
var age int = 30
var active bool = true

// 型を推論
var country = "Japan"  // 型が string と推論される
```

**複数の変数を同時に宣言:**
```go
var (
  name string = "Taro"
  age int = 30
  country = "Japan"
)
```

#### 方法2: `:=` 演算子（短い宣言形式、関数内のみ）

```go
// 最も短く書く方法（type は推論）
name := "Taro"
age := 30
country := "Japan"
```

⚠️ `:=` は関数内でのみ使用可能です。パッケージレベルの変数には `var` を使います。

#### 方法3: `const` 定数

```go
const pi = 3.14159
const maxRetries int = 5

// 複数の定数を同時に宣言
const (
  red = 0
  green = 1
  blue = 2
)
```

**比較:**

| 方法 | 場所 | 例 |
|------|------|-----|
| `var` | パッケージ・関数内 | `var name string = "Taro"` |
| `:=` | 関数内のみ | `name := "Taro"` |
| `const` | パッケージ・関数内 | `const pi = 3.14` |

**TypeScript との比較:**

```typescript
// TypeScript
const name: string = "Taro";
let age: number = 30;
var country: string = "Japan";  // 非推奨
```

```go
// Go
const name = "Taro"  // 型は推論
var age = 30         // または age := 30（関数内）
var country = "Japan"
```

---

### 2-2. 関数定義

**基本的な関数:**

```go
func Greet(name string) string {
  return "Hello, " + name
}

// 使用
result := Greet("World")  // "Hello, World"
```

**複数のパラメータ:**

```go
func Add(a int, b int) int {
  return a + b
}

// 同じ型が連続する場合は短縮形が使える
func Add(a, b int) int {
  return a + b
}
```

**複数の戻り値（Go の特徴）:**

```go
func DivideWithRemainder(a, b int) (int, int) {
  return a / b, a % b
}

// 使用
quotient, remainder := DivideWithRemainder(10, 3)
fmt.Println(quotient, remainder)  // 3, 1
```

**戻り値に名前をつける（Named Return Values）:**

```go
func Divide(a, b float64) (result float64, err error) {
  if b == 0 {
    err = errors.New("division by zero")
    return  // result = 0, err = エラー
  }
  result = a / b
  return
}
```

**可変長引数:**

```go
func Sum(numbers ...int) int {
  total := 0
  for _, num := range numbers {
    total += num
  }
  return total
}

// 使用
Sum(1, 2, 3, 4, 5)  // 15
```

**TypeScript との比較:**

```typescript
// TypeScript
function add(a: number, b: number): number {
  return a + b;
}

function divide(a: number, b: number): [number, Error | null] {
  if (b === 0) {
    return [0, new Error("division by zero")];
  }
  return [a / b, null];
}
```

```go
// Go
func Add(a, b int) int {
  return a + b
}

func Divide(a, b float64) (float64, error) {
  if b == 0 {
    return 0, errors.New("division by zero")
  }
  return a / b, nil
}
```

---

### 2-3. 構造体（struct）- JavaScript のクラスとの比較

**JavaScript のクラス:**

```javascript
class User {
  constructor(name, email, age) {
    this.name = name;
    this.email = email;
    this.age = age;
  }

  getInfo() {
    return `${this.name} (${this.email})`;
  }
}

const user = new User("Taro", "taro@example.com", 30);
console.log(user.getInfo());
```

**Go の構造体:**

```go
// 構造体の定義
type User struct {
  Name  string
  Email string
  Age   int
}

// 利用
user := User{
  Name:  "Taro",
  Email: "taro@example.com",
  Age:   30,
}

fmt.Println(user.Name)    // "Taro"
fmt.Println(user.Email)   // "taro@example.com"
```

**構造体の初期化方法:**

```go
// 方法1: フィールド名を明示
user := User{
  Name:  "Taro",
  Email: "taro@example.com",
  Age:   30,
}

// 方法2: 順序で指定（推奨しない）
user := User{"Taro", "taro@example.com", 30}

// 方法3: ゼロ値で初期化
var user User
user.Name = "Taro"
user.Email = "taro@example.com"
user.Age = 30

// 方法4: new を使って初期化（ポインタを返す）
user := new(User)
user.Name = "Taro"
```

**重要な違い:**

| 項目 | JavaScript | Go |
|------|-----------|-----|
| インスタンス化 | `new User()` | `User{...}` |
| メソッド定義 | クラス内に定義 | 構造体の外で定義（レシーバー） |
| 継承 | あり（`extends`) | なし（埋め込み） |
| プライベート | `#` で始まる | 小文字で始まる |

---

### 2-4. メソッド（レシーバー）

Go では、クラスのメソッドのようなものを「レシーバーを持つ関数」として定義します。

```go
type User struct {
  Name  string
  Email string
  Age   int
}

// ポインタレシーバー（推奨）
func (u *User) UpdateEmail(newEmail string) {
  u.Email = newEmail
}

// 値レシーバー
func (u User) GetInfo() string {
  return u.Name + " (" + u.Email + ")"
}

// 使用
user := User{Name: "Taro", Email: "taro@example.com", Age: 30}
user.UpdateEmail("newemail@example.com")
fmt.Println(user.GetInfo())
```

**ポインタレシーバー vs 値レシーバー:**

```go
// 値レシーバー: 構造体のコピーで操作（元の構造体は変わらない）
func (u User) SetName(name string) {
  u.Name = name
}

user := User{Name: "Taro"}
user.SetName("Hanako")
fmt.Println(user.Name)  // "Taro"（変わらない）

// ポインタレシーバー: ポインタで操作（元の構造体が変わる）
func (u *User) SetName(name string) {
  u.Name = name
}

user := User{Name: "Taro"}
user.SetName("Hanako")
fmt.Println(user.Name)  // "Hanako"（変わった）
```

**TypeScript のクラスとの比較:**

```typescript
// TypeScript
class User {
  constructor(
    public name: string,
    public email: string,
    public age: number
  ) {}

  updateEmail(newEmail: string): void {
    this.email = newEmail;
  }

  getInfo(): string {
    return `${this.name} (${this.email})`;
  }
}

const user = new User("Taro", "taro@example.com", 30);
user.updateEmail("newemail@example.com");
console.log(user.getInfo());
```

```go
// Go
type User struct {
  Name  string
  Email string
  Age   int
}

func (u *User) UpdateEmail(newEmail string) {
  u.Email = newEmail
}

func (u User) GetInfo() string {
  return u.Name + " (" + u.Email + ")"
}

user := &User{Name: "Taro", Email: "taro@example.com", Age: 30}
user.UpdateEmail("newemail@example.com")
fmt.Println(user.GetInfo())
```

---

### 2-5. インターフェース

インターフェースは、メソッドのセットを定義します。そのインターフェースを満たす型なら、どの型でも使えます。

**TypeScript のインターフェース:**

```typescript
interface Reader {
  read(): string;
}

class FileReader implements Reader {
  read(): string {
    return "file content";
  }
}

class NetworkReader implements Reader {
  read(): string {
    return "network content";
  }
}

function processData(reader: Reader): void {
  console.log(reader.read());
}

processData(new FileReader());    // "file content"
processData(new NetworkReader()); // "network content"
```

**Go のインターフェース（Implicit Implementation）:**

```go
// インターフェース定義
type Reader interface {
  Read() string
}

// FileReader 型
type FileReader struct {
  Path string
}

func (f FileReader) Read() string {
  return "file content"
}

// NetworkReader 型
type NetworkReader struct {
  URL string
}

func (n NetworkReader) Read() string {
  return "network content"
}

// 関数
func ProcessData(reader Reader) {
  fmt.Println(reader.Read())
}

// 使用
ProcessData(FileReader{Path: "/path/to/file"})      // "file content"
ProcessData(NetworkReader{URL: "http://example.com"})  // "network content"
```

**Go のインターフェースの特徴:**

- **Implicit Implementation**: 明示的に `implements` と書く必要がない
- **ダックタイピング**: インターフェースで定義されたメソッドを持っていれば、自動的にそのインターフェースを満たす
- **空のインターフェース**: `interface{}` は任意の型を受け入れる

```go
// 空のインターフェース（any と同じ）
var x interface{} = "hello"
var y interface{} = 42
var z interface{} = []int{1, 2, 3}

// どんな型でも入れられる
func PrintAnything(v interface{}) {
  fmt.Println(v)
}
```

---

### 2-6. ポインタ（&, *）

ポインタは JavaScript にない概念です。変数のメモリアドレスを指すものです。

```go
// ポインタの基礎
var x int = 10
var p *int = &x  // &x で x のアドレス、*int はそのアドレスを指す整数型ポインタ

fmt.Println(x)    // 10（値）
fmt.Println(&x)   // 0xc0000120a0（アドレス）
fmt.Println(p)    // 0xc0000120a0（ポインタが指すアドレス）
fmt.Println(*p)   // 10（ポインタの指す値）
```

**ポインタで値を変更:**

```go
var x int = 10
p := &x

*p = 20  // ポインタを通して値を変更

fmt.Println(x)   // 20
fmt.Println(*p)  // 20
```

**構造体のポインタ:**

```go
type User struct {
  Name string
  Age  int
}

user := User{Name: "Taro", Age: 30}
p := &user

// ポインタを通してフィールドにアクセス
p.Name = "Hanako"
p.Age = 25

fmt.Println(user.Name)  // "Hanako"
```

**ポインタレシーバーで構造体を変更:**

```go
type User struct {
  Name string
  Age  int
}

// ポインタレシーバー
func (u *User) HaveBirthday() {
  u.Age++
}

user := User{Name: "Taro", Age: 30}
user.HaveBirthday()  // Go が自動的に &user に変換

fmt.Println(user.Age)  // 31
```

**ポインタと nil:**

```go
var p *int = nil

if p == nil {
  fmt.Println("pointer is nil")
}

// nil ポインタへのアクセスはパニック
// fmt.Println(*p)  // panic: runtime error: invalid memory address or nil pointer dereference
```

**比較:**

| 項目 | JavaScript | Go |
|------|-----------|-----|
| ポインタ | 存在しない | &x で取得、*p で参照 |
| 参照型 | オブジェクト、配列など | ポインタ、スライス、マップなど |
| メモリ管理 | GC が自動管理 | GC が自動管理 |

---

### 2-7. スライスとマップ

**スライス（可変長の配列のような型）:**

```go
// 配列（固定長）
var arr [3]int = [3]int{1, 2, 3}

// スライス（可変長）
slice := []int{1, 2, 3, 4, 5}

// スライスの長さとキャパシティ
fmt.Println(len(slice))    // 5
fmt.Println(cap(slice))    // 5

// スライスへの要素追加
slice = append(slice, 6, 7)
fmt.Println(slice)         // [1 2 3 4 5 6 7]

// スライスの部分抽出
subSlice := slice[1:4]     // インデックス 1～3
fmt.Println(subSlice)      // [2 3 4]
```

**TypeScript との比較:**

```typescript
// TypeScript の配列
const arr: number[] = [1, 2, 3, 4, 5];
arr.push(6, 7);
const subArr = arr.slice(1, 4);
```

```go
// Go のスライス
slice := []int{1, 2, 3, 4, 5}
slice = append(slice, 6, 7)
subSlice := slice[1:4]
```

**マップ（辞書・オブジェクトのような型）:**

```go
// マップの作成
user := map[string]string{
  "name":  "Taro",
  "email": "taro@example.com",
}

// 要素へのアクセス
fmt.Println(user["name"])  // "Taro"

// 要素の追加・変更
user["age"] = "30"

// 要素の削除
delete(user, "age")

// キーの存在確認
value, ok := user["name"]
if ok {
  fmt.Println(value)       // "Taro"
} else {
  fmt.Println("key not found")
}
```

**TypeScript との比較:**

```typescript
// TypeScript のオブジェクト
const user: Record<string, string> = {
  name: "Taro",
  email: "taro@example.com",
};

user.age = "30";
delete user.age;

if ("name" in user) {
  console.log(user["name"]);
}
```

```go
// Go のマップ
user := map[string]string{
  "name":  "Taro",
  "email": "taro@example.com",
}

user["age"] = "30"
delete(user, "age")

if value, ok := user["name"]; ok {
  fmt.Println(value)
}
```

---

## 3. このプロジェクトで使う Go の機能

### 3-1. Struct Tags

構造体のタグ（`tag`）は、JSON のシリアライズやデータベースのマッピングに使われます。

```go
type Delivery struct {
  ID        int    `json:"id" gorm:"primaryKey"`
  Status    string `json:"status" gorm:"column:status"`
  Address   string `json:"address" gorm:"column:address"`
  Timestamp int64  `json:"timestamp" gorm:"column:created_at"`
}
```

**タグの意味:**

- `json:"id"`: JSON に変換する際のキー名
- `gorm:"primaryKey"`: GORM でプライマリキーとして認識
- `gorm:"column:status"`: データベースのカラム名

**例: JSON との変換**

```go
import "encoding/json"

delivery := Delivery{
  ID:      1,
  Status:  "pending",
  Address: "Tokyo",
}

// Go 構造体 → JSON
jsonData, _ := json.Marshal(delivery)
fmt.Println(string(jsonData))
// {"id":1,"status":"pending","address":"Tokyo","timestamp":0}

// JSON → Go 構造体
jsonStr := `{"id":1,"status":"completed"}`
var delivery2 Delivery
json.Unmarshal([]byte(jsonStr), &delivery2)
fmt.Println(delivery2.Status)  // "completed"
```

**タグのオプション:**

```go
type User struct {
  Name     string `json:"name" gorm:"column:name;not null"`
  Email    string `json:"email,omitempty" gorm:"column:email;unique"`
  Password string `json:"-" gorm:"column:password"`  // JSON に含めない
}
```

| タグ | 説明 |
|------|------|
| `json:"name"` | JSON キー名を指定 |
| `json:"name,omitempty"` | ゼロ値の場合は JSON に含めない |
| `json:"-"` | JSON に含めない |
| `gorm:"primaryKey"` | プライマリキー |
| `gorm:"column:name"` | データベースのカラム名 |
| `gorm:"not null"` | NOT NULL 制約 |
| `gorm:"unique"` | UNIQUE 制約 |

---

### 3-2. Import と Package

**パッケージの構成:**

```
myproject/
├── go.mod
├── go.sum
├── main.go
├── utils/
│   └── helpers.go
└── handlers/
    └── delivery.go
```

**go.mod:**

```go
module myproject

go 1.21
```

**utils/helpers.go:**

```go
package utils

func Greet(name string) string {
  return "Hello, " + name
}
```

**handlers/delivery.go:**

```go
package handlers

import (
  "fmt"
  "myproject/utils"  // 同じプロジェクト内のパッケージをインポート
)

func HandleDelivery() {
  greeting := utils.Greet("World")
  fmt.Println(greeting)
}
```

**main.go:**

```go
package main

import (
  "fmt"
  "myproject/handlers"  // パッケージをインポート
)

func main() {
  handlers.HandleDelivery()
}
```

**外部パッケージのインポート:**

```go
import (
  "fmt"                  // 標準ライブラリ
  "net/http"             // 標準ライブラリ
  "github.com/gin-gonic/gin"  // 外部パッケージ
  "gorm.io/gorm"              // 外部パッケージ
)
```

---

### 3-3. エラーハンドリングパターン

**基本的なパターン:**

```go
func ReadFile(path string) (string, error) {
  data, err := ioutil.ReadFile(path)
  if err != nil {
    return "", err  // エラーを呼び出し元に返す
  }
  return string(data), nil
}

// 使用
content, err := ReadFile("file.txt")
if err != nil {
  fmt.Println("Error:", err)
  return
}
fmt.Println(content)
```

**エラーをラップして情報を追加:**

```go
import "fmt"

func FetchUser(id int) (*User, error) {
  // ...
  if user == nil {
    return nil, fmt.Errorf("user with id %d not found", id)
  }
  return user, nil
}
```

**複数のエラーチェック:**

```go
func ProcessData() error {
  file, err := os.Open("data.txt")
  if err != nil {
    return fmt.Errorf("failed to open file: %w", err)
  }
  defer file.Close()

  data, err := ioutil.ReadAll(file)
  if err != nil {
    return fmt.Errorf("failed to read file: %w", err)
  }

  // 処理...
  return nil
}
```

**カスタムエラー型:**

```go
type ValidationError struct {
  Field   string
  Message string
}

func (e ValidationError) Error() string {
  return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// 使用
func ValidateEmail(email string) error {
  if email == "" {
    return ValidationError{
      Field:   "email",
      Message: "email cannot be empty",
    }
  }
  return nil
}
```

---

### 3-4. ゴルーチンとチャネル（概要）

Go は「ゴルーチン」という軽量なスレッド機構を持っています。このプロジェクトではあまり使いませんが、基本を理解しておくと良いです。

**ゴルーチン（並行処理）:**

```go
import (
  "fmt"
  "time"
)

func Worker(id int) {
  for i := 1; i <= 3; i++ {
    fmt.Printf("Worker %d: Task %d\n", id, i)
    time.Sleep(1 * time.Second)
  }
}

func main() {
  // 通常の呼び出し（順序に実行）
  Worker(1)
  Worker(2)

  // ゴルーチンで並行実行
  go Worker(1)
  go Worker(2)

  time.Sleep(4 * time.Second)  // ゴルーチンの完了を待つ
}
```

**チャネル（ゴルーチン間の通信）:**

```go
import "fmt"

func main() {
  // チャネルの作成
  ch := make(chan string)

  // ゴルーチンでデータを送信
  go func() {
    ch <- "Hello from goroutine"
  }()

  // チャネルからデータを受信
  message := <-ch
  fmt.Println(message)  // "Hello from goroutine"
}
```

このプロジェクトではシンプルなCRUD操作を扱うため、ゴルーチンとチャネルは複雑には使いません。より詳しく学びたい場合は、公式ドキュメントを参照してください。

---

## 4. Gin フレームワーク

Gin は Go で Web API を開発するフレームワークです。Express.js に似ています。

### 4-1. Express.js との比較

**Express.js:**

```javascript
const express = require('express');
const app = express();

app.use(express.json());

// ルート定義
app.get('/api/health', (req, res) => {
  res.json({ status: 'healthy' });
});

app.post('/api/deliveries', (req, res) => {
  const delivery = req.body;
  // 処理...
  res.status(201).json(delivery);
});

app.listen(8080, () => {
  console.log('Server running on port 8080');
});
```

**Gin:**

```go
package main

import (
  "github.com/gin-gonic/gin"
  "net/http"
)

func main() {
  r := gin.Default()

  // ルート定義
  r.GET("/api/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
  })

  r.POST("/api/deliveries", func(c *gin.Context) {
    var delivery Delivery
    c.ShouldBindJSON(&delivery)
    // 処理...
    c.JSON(http.StatusCreated, delivery)
  })

  r.Run(":8080")
}
```

**比較:**

| 項目 | Express.js | Gin |
|------|-----------|-----|
| インスタンス作成 | `express()` | `gin.Default()` |
| ルート定義 | `app.get()`, `app.post()` | `r.GET()`, `r.POST()` |
| Context オブジェクト | `req`, `res` | `c *gin.Context` |
| JSON レスポンス | `res.json()` | `c.JSON()` |
| ポート起動 | `app.listen()` | `r.Run()` |

---

### 4-2. gin.Context の役割

`gin.Context` はリクエスト・レスポンスの処理に必要なすべての情報を含むオブジェクトです。Express.js の `req`, `res` を合わせたようなものです。

```go
func GetDelivery(c *gin.Context) {
  // パラメータを取得
  id := c.Param("id")

  // クエリパラメータを取得
  filter := c.Query("status")

  // ボディから JSON をパース
  var delivery Delivery
  c.ShouldBindJSON(&delivery)

  // レスポンスを返す
  c.JSON(200, gin.H{"id": id, "data": delivery})

  // エラーレスポンス
  c.JSON(400, gin.H{"error": "Invalid request"})
}
```

---

### 4-3. よく使う Gin のメソッド

**c.JSON: JSON レスポンスを返す**

```go
r.GET("/users", func(c *gin.Context) {
  users := []User{{Name: "Taro"}, {Name: "Hanako"}}
  c.JSON(200, users)
})
```

**c.ShouldBindJSON: リクエストボディを構造体にパース**

```go
r.POST("/users", func(c *gin.Context) {
  var user User
  if err := c.ShouldBindJSON(&user); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
  }
  // user を使用
  c.JSON(201, user)
})
```

**c.Param: URL パラメータを取得**

```go
r.GET("/users/:id", func(c *gin.Context) {
  id := c.Param("id")
  // id を使用
  c.JSON(200, gin.H{"id": id})
})
```

**c.Query: クエリパラメータを取得**

```go
r.GET("/deliveries", func(c *gin.Context) {
  status := c.Query("status")        // ?status=pending
  limit := c.DefaultQuery("limit", "10")  // デフォルト値を指定
  c.JSON(200, gin.H{"status": status, "limit": limit})
})
```

**c.GetJSON: JSON形式でレスポンスを返す**

```go
c.JSON(http.StatusOK, gin.H{
  "message": "success",
  "data": user,
})
```

---

### 4-4. ルーティング

**基本的なルート:**

```go
r := gin.Default()

// GET
r.GET("/path", handler)

// POST
r.POST("/path", handler)

// PUT
r.PUT("/path", handler)

// DELETE
r.DELETE("/path", handler)

// PATCH
r.PATCH("/path", handler)
```

**URL パラメータ:**

```go
r.GET("/users/:id", func(c *gin.Context) {
  id := c.Param("id")
  c.JSON(200, gin.H{"id": id})
})

r.GET("/users/:id/posts/:post_id", func(c *gin.Context) {
  id := c.Param("id")
  postID := c.Param("post_id")
  c.JSON(200, gin.H{"user_id": id, "post_id": postID})
})
```

**ルートグループ:**

```go
api := r.Group("/api")
{
  api.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
  })

  v1 := api.Group("/v1")
  {
    v1.GET("/deliveries", getDeliveries)
    v1.POST("/deliveries", createDelivery)
    v1.GET("/deliveries/:id", getDelivery)
  }
}

// /api/health
// /api/v1/deliveries
// /api/v1/deliveries/:id
```

---

### 4-5. ミドルウェア

ミドルウェアはすべてのリクエストで実行される処理です（Express.js と同じ概念）。

**ロギングミドルウェア:**

```go
func LoggingMiddleware(c *gin.Context) {
  log.Printf("Request: %s %s", c.Request.Method, c.Request.URL)
  c.Next()  // 次のハンドラーに処理を進める
  log.Printf("Response: %d", c.Writer.Status())
}

r := gin.Default()
r.Use(LoggingMiddleware)  // ミドルウェアを登録

r.GET("/users", func(c *gin.Context) {
  c.JSON(200, gin.H{"message": "ok"})
})
```

**認証ミドルウェア:**

```go
func AuthMiddleware(c *gin.Context) {
  token := c.GetHeader("Authorization")
  if token == "" {
    c.JSON(401, gin.H{"error": "missing authorization header"})
    c.Abort()  // 処理を中止
    return
  }

  // トークンを検証
  if !validateToken(token) {
    c.JSON(401, gin.H{"error": "invalid token"})
    c.Abort()
    return
  }

  c.Next()  // 次のハンドラーへ
}

r.Use(AuthMiddleware)
```

**特定のルートにのみミドルウェアを適用:**

```go
r.GET("/public", func(c *gin.Context) {
  c.JSON(200, gin.H{"message": "public"})
})

protected := r.Group("/protected")
protected.Use(AuthMiddleware)
{
  protected.GET("/secret", func(c *gin.Context) {
    c.JSON(200, gin.H{"message": "secret"})
  })
}
```

---

## 5. GORM

GORM は Go の ORM ライブラリです。Prisma や TypeORM と同様の役割です。

### 5-1. ORM とは

ORM（Object-Relational Mapping）は、データベーステーブルをプログラムの構造体にマッピングするライブラリです。

**SQL を直接書く場合:**

```sql
SELECT id, name, email FROM users WHERE id = 1;
INSERT INTO users (name, email) VALUES ('Taro', 'taro@example.com');
UPDATE users SET name = 'Hanako' WHERE id = 1;
DELETE FROM users WHERE id = 1;
```

**GORM を使う場合:**

```go
var user User
db.First(&user, 1)  // SELECT

db.Create(&User{Name: "Taro", Email: "taro@example.com"})  // INSERT

db.Model(&user).Update("name", "Hanako")  // UPDATE

db.Delete(&user)  // DELETE
```

---

### 5-2. Prisma/TypeORM との比較

**Prisma (TypeScript):**

```typescript
const user = await prisma.user.findUnique({
  where: { id: 1 },
});

const newUser = await prisma.user.create({
  data: {
    name: "Taro",
    email: "taro@example.com",
  },
});

await prisma.user.update({
  where: { id: 1 },
  data: { name: "Hanako" },
});

await prisma.user.delete({
  where: { id: 1 },
});
```

**TypeORM (TypeScript):**

```typescript
const user = await userRepository.findOne({ where: { id: 1 } });

const newUser = userRepository.create({
  name: "Taro",
  email: "taro@example.com",
});
await userRepository.save(newUser);

await userRepository.update({ id: 1 }, { name: "Hanako" });

await userRepository.delete({ id: 1 });
```

**GORM (Go):**

```go
var user User
db.First(&user, 1)

db.Create(&User{Name: "Taro", Email: "taro@example.com"})

db.Model(&user).Update("name", "Hanako")

db.Delete(&user)
```

---

### 5-3. 基本的なセットアップ

**go.mod:**

```go
require gorm.io/gorm v1.25.0
require gorm.io/driver/mysql v1.5.0
```

**main.go:**

```go
package main

import (
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
  "log"
)

func main() {
  // データベースに接続
  dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True"
  db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
  if err != nil {
    log.Fatal("Failed to connect to database:", err)
  }

  // テーブルを自動作成
  db.AutoMigrate(&User{}, &Delivery{})

  // これで db を使用できる
}
```

---

### 5-4. AutoMigrate（自動マイグレーション）

テーブルを自動的に作成・更新します。

```go
type User struct {
  ID        uint      `gorm:"primaryKey"`
  Name      string
  Email     string
  Age       int
  CreatedAt time.Time
}

type Delivery struct {
  ID        uint       `gorm:"primaryKey"`
  UserID    uint       `gorm:"foreignKey"`
  Status    string
  Address   string
  CreatedAt time.Time
}

// テーブルを自動作成
db.AutoMigrate(&User{}, &Delivery{})
```

**詳細なカラム定義:**

```go
type Product struct {
  ID          uint      `gorm:"primaryKey"`
  Name        string    `gorm:"column:name;not null;unique"`
  Price       float64   `gorm:"column:price;type:decimal(10,2)"`
  Description string    `gorm:"column:description;type:text"`
  CreatedAt   time.Time `gorm:"column:created_at"`
}
```

---

### 5-5. CRUD 操作

**Create（作成）:**

```go
// 1つのレコードを作成
user := User{Name: "Taro", Email: "taro@example.com"}
result := db.Create(&user)
if result.Error != nil {
  log.Fatal(result.Error)
}
fmt.Println(user.ID)  // ID が自動割り当てされる

// 複数のレコードを一括作成
users := []User{
  {Name: "Taro", Email: "taro@example.com"},
  {Name: "Hanako", Email: "hanako@example.com"},
}
db.Create(&users)
```

**Read（読み込み）:**

```go
// ID で取得
var user User
db.First(&user, 1)  // ID = 1

// 条件で取得
var users []User
db.Where("age > ?", 20).Find(&users)

// 単一レコードを取得
var user User
db.Where("email = ?", "taro@example.com").First(&user)

// 存在確認
var count int64
db.Model(&User{}).Where("id = ?", 1).Count(&count)
if count > 0 {
  fmt.Println("User exists")
}
```

**Update（更新）:**

```go
// 全フィールドを更新
var user User
db.Model(&user).Updates(User{Name: "Hanako", Age: 25})

// 特定フィールドを更新
db.Model(&user).Update("name", "Hanako")

// 複数フィールドを更新
db.Model(&user).Updates(map[string]interface{}{
  "name": "Hanako",
  "age":  25,
})

// 条件付き更新
db.Where("age > ?", 20).Update("status", "senior")
```

**Delete（削除）:**

```go
// ID で削除
var user User
db.Delete(&user, 1)

// 条件で削除
db.Where("age < ?", 18).Delete(&User{})

// すべてを削除（危険！）
// db.DeleteAll(&User{})
```

---

### 5-6. 構造体タグとカラムマッピング

構造体タグでカラム名やデータベース制約を指定します。

```go
type User struct {
  ID        uint   `gorm:"primaryKey"`              // プライマリキー
  Name      string `gorm:"column:name;not null"`    // カラム名、NOT NULL
  Email     string `gorm:"column:email;unique"`     // UNIQUE 制約
  Password  string `gorm:"column:password"`
  Age       int    `gorm:"column:age;default:0"`    // デフォルト値
  Active    bool   `gorm:"column:active;default:true"`
  CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
  UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
```

**タグの組み合わせ:**

```go
type Product struct {
  ID          uint    `gorm:"primaryKey;autoIncrement"`
  SKU         string  `gorm:"column:sku;unique;not null"`
  Name        string  `gorm:"column:name;not null;index"`
  Price       float64 `gorm:"column:price;type:decimal(10,2);not null"`
  Stock       int     `gorm:"column:stock;default:0"`
  Description string  `gorm:"column:description;type:text"`
}
```

**よく使うタグオプション:**

| オプション | 説明 |
|-----------|------|
| `primaryKey` | プライマリキー |
| `column:name` | カラム名を指定 |
| `not null` | NOT NULL 制約 |
| `unique` | UNIQUE 制約 |
| `default:value` | デフォルト値 |
| `index` | インデックスを作成 |
| `type:TEXT` | カラムの型を指定 |
| `autoCreateTime` | レコード作成時にタイムスタンプを自動設定 |
| `autoUpdateTime` | レコード更新時にタイムスタンプを自動設定 |

---

### 5-7. リレーション（関連）

**One-to-Many（1対多）:**

```go
type User struct {
  ID        uint
  Name      string
  Deliveries []Delivery  // 1人のユーザーが複数の配送を持つ
}

type Delivery struct {
  ID     uint
  UserID uint
  User   User    // 逆参照
  Status string
}

// 取得時にリレーションをロード
var user User
db.Preload("Deliveries").First(&user, 1)
```

**Many-to-Many（多対多）:**

```go
type Student struct {
  ID    uint
  Name  string
  Courses []Course `gorm:"many2many:student_courses;"`
}

type Course struct {
  ID       uint
  Name     string
  Students []Student `gorm:"many2many:student_courses;"`
}

// 自動的に中間テーブル student_courses が作成される
```

---

## 実践例：APIエンドポイントの実装

ここまでで学んだことを組み合わせて、簡単な API エンドポイントを実装してみましょう。

```go
package main

import (
  "github.com/gin-gonic/gin"
  "gorm.io/driver/mysql"
  "gorm.io/gorm"
  "net/http"
)

type Delivery struct {
  ID      uint   `json:"id" gorm:"primaryKey"`
  Status  string `json:"status" gorm:"column:status"`
  Address string `json:"address" gorm:"column:address"`
}

var db *gorm.DB

func main() {
  // データベースに接続
  dsn := "user:password@tcp(localhost:3306)/deliveries_db"
  var err error
  db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
  if err != nil {
    panic(err)
  }

  // テーブルを自動作成
  db.AutoMigrate(&Delivery{})

  // Gin エンジンを作成
  r := gin.Default()

  // ルートを定義
  api := r.Group("/api")
  {
    v1 := api.Group("/v1")
    {
      // ヘルスチェック
      v1.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
      })

      // 配送一覧を取得
      v1.GET("/deliveries", func(c *gin.Context) {
        var deliveries []Delivery
        db.Find(&deliveries)
        c.JSON(http.StatusOK, deliveries)
      })

      // 配送を作成
      v1.POST("/deliveries", func(c *gin.Context) {
        var delivery Delivery
        if err := c.ShouldBindJSON(&delivery); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
          return
        }

        if result := db.Create(&delivery); result.Error != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
          return
        }

        c.JSON(http.StatusCreated, delivery)
      })

      // 配送を取得
      v1.GET("/deliveries/:id", func(c *gin.Context) {
        id := c.Param("id")
        var delivery Delivery
        if result := db.First(&delivery, id); result.Error != nil {
          c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
          return
        }

        c.JSON(http.StatusOK, delivery)
      })

      // 配送を更新
      v1.PUT("/deliveries/:id", func(c *gin.Context) {
        id := c.Param("id")
        var delivery Delivery
        if result := db.First(&delivery, id); result.Error != nil {
          c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
          return
        }

        if err := c.ShouldBindJSON(&delivery); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
          return
        }

        db.Save(&delivery)
        c.JSON(http.StatusOK, delivery)
      })

      // 配送を削除
      v1.DELETE("/deliveries/:id", func(c *gin.Context) {
        id := c.Param("id")
        if result := db.Delete(&Delivery{}, id); result.Error != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
          return
        }

        c.JSON(http.StatusNoContent, nil)
      })
    }
  }

  // サーバーを起動
  r.Run(":8080")
}
```

このコードで以下が実現されます：

- `GET /api/v1/health` - ヘルスチェック
- `GET /api/v1/deliveries` - 全配送を取得
- `POST /api/v1/deliveries` - 配送を作成
- `GET /api/v1/deliveries/:id` - 特定の配送を取得
- `PUT /api/v1/deliveries/:id` - 配送を更新
- `DELETE /api/v1/deliveries/:id` - 配送を削除

---

## まとめ

このガイドで、JavaScript/TypeScript 開発者が Go を学ぶために必要な基本概念をカバーしました：

1. ✅ 型システム、コンパイル言語、パッケージシステムの違いを理解した
2. ✅ 変数宣言、関数、構造体、メソッド、ポインタなど Go の基本構文を学んだ
3. ✅ Struct タグ、エラーハンドリング、ゴルーチンを学んだ
4. ✅ Gin フレームワークで HTTP API を構築する方法を学んだ
5. ✅ GORM で データベース操作を行う方法を学んだ

これでこのプロジェクトの開発を始める準備ができました。わからないことがあれば、公式ドキュメントを参照するか、質問してください。

Happy coding! 🎉
