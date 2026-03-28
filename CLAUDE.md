# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go DDD 範本專案，實作 Clean Architecture 與 Domain-Driven Design 原則。

## Architecture (BC-first + Clean Architecture + DDD)

### 組織方式：Bounded Context First

本專案以 **Bounded Context (BC)** 為第一層組織單位，每個 BC 內部再按 Clean Architecture 分層。
跨 BC 共用的元件放在 `shared/`（Shared Kernel）。

```text
internal/
├── shared/              # Shared Kernel（跨 BC 共用）
│   ├── domain/          # 共用領域概念（event, valueobject, pagination）
│   ├── adapter/         # 共用適配器（HTTP server, ACL 文件）
│   └── infrastructure/  # 共用基礎設施（config, messaging）
├── {bc_name}/           # 各 Bounded Context
│   ├── domain/          # 該 BC 的領域層
│   ├── application/     # 該 BC 的應用層
│   ├── adapter/         # 該 BC 的適配器層
│   └── infrastructure/  # 該 BC 的基礎設施層
```

### Clean Architecture 分層（每個 BC 內部）

```text
┌─────────────────────────────────────────────────────────────┐
│                        Adapter Layer                        │
│              (HTTP handler, gRPC, Consumer, ACL)            │
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
  - `valueobject/` - BC 內部的值物件
- **依賴**: 僅依賴 `shared/domain/`

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
  - `http/handler/` - REST API handler（註冊到 shared router）
  - `grpc/` - gRPC 服務
  - `consumer/` - 訊息佇列消費者
  - `acl/{external_bc}/` - Anti-Corruption Layer（翻譯外部 BC 模型）
- **依賴**: Domain Layer、Application Layer

#### 4. Infrastructure Layer - 技術實作

- **職責**: 提供技術基礎設施
- **元件**:
  - `persistence/` - Repository 實作（postgres, redis）
  - `external/` - 外部服務客戶端（實作 port 介面）
- **依賴**: 可依賴所有層（提供技術支援）

### Shared Kernel

`internal/shared/` 包含跨 BC 共用的元件：

- **`shared/domain/event/`** - Domain Event 介面 (`DomainEvent`) 和基礎結構 (`BaseEvent`)、事件匯流排介面 (`EventBus`)
- **`shared/domain/valueobject/`** - 跨 BC 共用的值物件
- **`shared/domain/pagination.go`** - 分頁值物件
- **`shared/adapter/http/`** - HTTP server、router、middleware、response 格式
- **`shared/adapter/acl/`** - ACL pattern 文件與慣例
- **`shared/infrastructure/config/`** - 配置管理
- **`shared/infrastructure/messaging/`** - EventBus 實作（in-memory 等）

### Domain Event 慣例

- 事件類型命名: `"{aggregate}.{past_tense_verb}"`（如 `"order.created"`）
- 具體事件嵌入 `event.BaseEvent`
- EventBus 介面在 `shared/domain/event/bus.go`，實作在 `shared/infrastructure/messaging/`
- BC 內部事件定義在 `{bc}/domain/{aggregate}/event.go`

### Anti-Corruption Layer (ACL) 慣例

ACL 用於翻譯外部 BC 的模型到本地領域模型，防止外部概念污染本地 BC：

```text
{bc}/application/port/{external}.go        ← 定義需要的介面
{bc}/adapter/acl/{external_bc}/
  translator.go                            ← 實作介面，翻譯模型
  client.go                                ← 定義 client 介面
{bc}/infrastructure/external/{service}/
  client.go                                ← 實作 raw client
```

## Project Structure

```text
.
├── cmd/                                # 應用程式進入點
│   └── service/
├── internal/                           # 私有應用程式碼
│   ├── shared/                         # Shared Kernel（跨 BC 共用）
│   │   ├── domain/
│   │   │   ├── event/                  # DomainEvent, EventBus 介面
│   │   │   ├── valueobject/            # 跨 BC 共用值物件
│   │   │   └── pagination.go           # 分頁值物件
│   │   ├── adapter/
│   │   │   ├── http/                   # HTTP server, router, middleware
│   │   │   └── acl/                    # ACL pattern 文件
│   │   └── infrastructure/
│   │       ├── config/                 # 配置管理
│   │       └── messaging/inmemory/     # In-memory EventBus 實作
│   └── {bc_name}/                      # 各 Bounded Context
│       ├── domain/
│       │   └── {aggregate}/
│       │       ├── entity.go           # 聚合根與實體
│       │       ├── repository.go       # Repository 介面
│       │       ├── service.go          # 領域服務（可選）
│       │       └── event.go            # 領域事件（可選）
│       ├── application/
│       │   ├── usecase/                # 用例實作
│       │   ├── port/                   # 外部服務介面
│       │   ├── dto/                    # 資料傳輸物件
│       │   └── mapper/                 # DTO ↔ Domain 映射
│       ├── adapter/
│       │   ├── http/handler/           # REST API handler
│       │   └── acl/{external_bc}/      # Anti-Corruption Layer
│       └── infrastructure/
│           ├── persistence/            # Repository 實作
│           └── external/               # 外部服務客戶端
├── pkg/                                # 公共可重用套件
├── configs/                            # 配置檔案
├── tests/
│   ├── integration/
│   └── e2e/
└── docs/
```

## Development Commands

本專案使用 [Task](https://taskfile.dev/) 管理開發工作流程。執行 `task` 可查看所有可用任務。

### Build & Run

```bash
task run                    # 執行服務
task build                  # 編譯二進位檔案
```

### Test

```bash
task test                   # 執行全部測試
task test:cover             # 測試覆蓋率
task test:race              # 競態檢測
task test:all               # 執行所有測試變體
```

### Code Quality

```bash
task lint                   # Lint 檢查（提交前必須通過）
task lint:fix               # Lint 自動修復
task fmt                    # 格式化程式碼
task imports                # 整理 imports
```

### Dependencies

```bash
task tidy                   # 整理依賴
task verify                 # 驗證依賴
```

### Code Generation

```bash
task generate               # 產生 Mocks
task generate:wire          # Wire 依賴注入
task swagger                # Swagger 文件
```

### Development Workflow

```bash
task dev                    # 開發檢查（lint + test）
task ci                     # CI 流程（fmt, lint, test, build）
task check                  # 提交前完整檢查
task clean                  # 清理編譯產物
```

## Task Completion Checklist

執行 `task check` 會自動完成以下檢查：

1. `task fmt` - 格式化程式碼
2. `task tidy` - 整理依賴
3. `task lint` - Lint 檢查通過
4. `task test` - 全部測試通過

## DDD 實踐指南

### Domain Entity 範例

```go
// internal/order/domain/order/order.go
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

Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask.
