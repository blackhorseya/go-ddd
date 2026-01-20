# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go DDD 範本專案，實作 Clean Architecture 與 Domain-Driven Design 原則。

## Architecture (Clean Architecture + DDD)

### 分層架構

```text
┌─────────────────────────────────────────────────────────────┐
│                        Adapter Layer                        │
│                   (HTTP, gRPC, Consumer)                    │
├─────────────────────────────────────────────────────────────┤
│                      Application Layer                      │
│                  (Use Cases, DTOs, Ports)                   │
├─────────────────────────────────────────────────────────────┤
│                        Domain Layer                         │
│         (Aggregates, Entities, Value Objects)               │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                     │
│        (Persistence, Messaging, External Services)          │
└─────────────────────────────────────────────────────────────┘

依賴方向：外層 → 內層（Domain 不依賴任何外層）
```

### 層級說明

#### 1. Domain Layer - 核心業務

- **職責**: 包含所有業務邏輯和規則
- **組織方式**: 按聚合（Aggregate）劃分套件
- **元件**:
  - `{aggregate}/` - 每個聚合一個套件
    - `entity.go` - 聚合根與實體
    - `repository.go` - Repository 介面
    - `service.go` - 領域服務（可選）
    - `event.go` - 領域事件（可選）
  - `valueobject/` - 跨聚合共用的值物件
  - `event/` - 跨聚合共用的事件定義
- **依賴**: 無外部依賴

#### 2. Application Layer - 用例協調

- **職責**: 協調領域層完成用例
- **元件**:
  - `usecase/` - 用例實作
  - `port/` - 外部服務介面（Output Ports）
  - `dto/` - 資料傳輸物件
  - `mapper/` - DTO 與領域物件轉換
- **依賴**: 僅依賴 Domain Layer

#### 3. Adapter Layer - 對外介面

- **職責**: 處理外部請求，轉換為應用層可理解的格式
- **元件**:
  - `http/` - REST API（handler, middleware, router）
  - `grpc/` - gRPC 服務
  - `consumer/` - 訊息佇列消費者
- **依賴**: Domain Layer、Application Layer

#### 4. Infrastructure Layer - 技術實作

- **職責**: 提供技術基礎設施
- **元件**:
  - `config/` - 配置管理
  - `persistence/` - Repository 實作（postgres, redis）
  - `messaging/` - 訊息佇列實作
  - `external/` - 外部服務客戶端（實作 port 介面）
  - `logger/` - 日誌實作
- **依賴**: 可依賴所有層（提供技術支援）

## Project Structure

```text
.
├── cmd/                        # 應用程式進入點
│   └── service/
├── internal/                   # 私有應用程式碼
│   ├── domain/                 # 領域層（按聚合組織）
│   │   ├── order/              # Order 聚合
│   │   │   ├── order.go        # 聚合根
│   │   │   ├── item.go         # 聚合內實體
│   │   │   ├── repository.go   # Repository 介面
│   │   │   └── service.go      # 領域服務
│   │   ├── user/               # User 聚合
│   │   │   ├── user.go
│   │   │   └── repository.go
│   │   ├── valueobject/        # 共用值物件
│   │   │   ├── money.go
│   │   │   └── address.go
│   │   └── event/              # 共用領域事件
│   ├── application/            # 應用層
│   │   ├── usecase/            # 用例實作
│   │   │   ├── create_order.go
│   │   │   └── confirm_order.go
│   │   ├── port/               # 外部服務介面
│   │   ├── dto/                # 資料傳輸物件
│   │   └── mapper/             # DTO ↔ Domain 映射
│   ├── adapter/                # 適配器層
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── router/
│   │   ├── grpc/
│   │   └── consumer/
│   └── infrastructure/         # 基礎設施層
│       ├── config/
│       ├── persistence/
│       │   ├── postgres/
│       │   └── redis/
│       ├── messaging/
│       ├── external/           # 外部服務客戶端
│       └── logger/
├── pkg/                        # 公共可重用套件
├── configs/                    # 配置檔案
├── tests/
│   ├── integration/
│   └── e2e/
└── docs/
```

## Development Commands

### Build & Run

```bash
go run cmd/service/main.go
go build -o bin/service cmd/service/main.go
```

### Test

```bash
go test ./...                           # 全部測試
go test -cover ./...                    # 覆蓋率
go test -race ./...                     # 競態檢測
go test -run TestName ./path/to/pkg     # 特定測試
```

### Code Quality

```bash
golangci-lint run           # Lint（提交前必須通過）
golangci-lint run --fix     # 自動修復
gofmt -w .                  # 格式化
goimports -w .              # 整理 imports
```

### Dependencies

```bash
go mod tidy                 # 整理依賴
go mod verify               # 驗證依賴
```

### Code Generation

```bash
go generate ./...                       # 產生 Mocks
go generate ./cmd/service/...           # Wire 依賴注入
swag init -g cmd/service/main.go        # Swagger 文件
```

## Task Completion Checklist

1. `golangci-lint run` 通過
2. `go test ./...` 全部通過
3. `go mod tidy` 已執行
4. `gofmt -w .` 已執行

## DDD 實踐指南

### Domain Entity 範例

```go
// internal/domain/order/order.go
package order

type Order struct {
    id        string       // private fields
    userID    string
    items     []Item
    status    Status
    createdAt time.Time
}

// Constructor - 確保物件永遠有效
func NewOrder(id, userID string, items []Item) (*Order, error) {
    if id == "" {
        return nil, errors.New("empty order id")
    }
    if len(items) == 0 {
        return nil, errors.New("order must have at least one item")
    }
    return &Order{
        id:        id,
        userID:    userID,
        items:     items,
        status:    StatusPending,
        createdAt: time.Now(),
    }, nil
}

// Behavior - 方法名稱反映業務行為
func (o *Order) Confirm() error {
    if o.status != StatusPending {
        return ErrCannotConfirm
    }
    o.status = StatusConfirmed
    return nil
}

// Getter (唯讀)
func (o *Order) ID() string { return o.id }
```

### 設計原則

1. **Domain 物件欄位皆為 private** - 透過 Getter 存取
2. **Constructor 驗證** - 物件永遠處於有效狀態
3. **行為導向命名** - `Confirm()` 而非 `SetStatus()`
4. **Repository 介面在 Domain** - 依賴反轉
5. **Use Case 只做編排** - 業務邏輯放 Domain
6. **按聚合組織套件** - 相關程式碼放一起

### 測試策略

| 類型     | 位置                     | 說明               |
| -------- | ------------------------ | ------------------ |
| 單元測試 | `*_test.go` 與程式碼並列 | 測試單一函數/方法  |
| 整合測試 | `tests/integration/`     | 測試元件間整合     |
| E2E 測試 | `tests/e2e/`             | 測試完整使用者流程 |

- 使用 Table-Driven Tests
- 遵循 Arrange-Act-Assert 模式
- Domain Layer 必須 100% 覆蓋
