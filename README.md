# Go DDD Template Project

Go 語言實作 Clean Architecture 與 Domain-Driven Design (DDD) 的專案範本。

## 快速開始

```bash
# 1. 使用此範本建立新專案
# 點擊 GitHub 上的 "Use this template"

# 2. Clone 新專案
git clone https://github.com/your-org/your-project.git
cd your-project

# 3. 更新模組名稱
go mod edit -module github.com/your-org/your-project

# 4. 安裝 Task（如尚未安裝）
# macOS: brew install go-task
# 其他系統: https://taskfile.dev/installation/

# 5. 安裝依賴
task tidy

# 6. 執行服務
task run
```

## 專案結構

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

## 架構概覽

採用 **Clean Architecture + Hexagonal Architecture + DDD** 設計：

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

### 層級職責

| 層級               | 職責                   | 元件                                                  |
| ------------------ | ---------------------- | ----------------------------------------------------- |
| **Domain**         | 核心業務邏輯與規則     | Aggregate, Entity, Value Object, Repository Interface |
| **Application**    | 用例編排，協調領域層   | Use Case, DTO, Port, Mapper                           |
| **Adapter**        | 處理外部請求，轉換格式 | HTTP Handler, gRPC Service, Consumer                  |
| **Infrastructure** | 技術基礎設施實作       | Repository Impl, Config, Logger, External             |

## 開發指令

本專案使用 [Task](https://taskfile.dev/) 管理開發工作流程。執行 `task` 可查看所有可用任務。

### 建置與執行

```bash
task run                    # 執行服務
task build                  # 編譯二進位檔案
```

### 測試

```bash
task test                   # 執行全部測試
task test:cover             # 測試覆蓋率
task test:race              # 競態檢測
task test:all               # 執行所有測試變體
```

### 程式碼品質

```bash
task lint                   # Lint 檢查（提交前必須通過）
task lint:fix               # Lint 自動修復
task fmt                    # 格式化程式碼
task imports                # 整理 imports
```

### 依賴管理

```bash
task tidy                   # 整理依賴
task verify                 # 驗證依賴
```

### 程式碼產生

```bash
task generate               # 產生 Mocks
task generate:wire          # Wire 依賴注入
task swagger                # Swagger 文件
```

### 開發工作流程

```bash
task dev                    # 開發檢查（lint + test）
task ci                     # CI 流程（fmt, lint, test, build）
task check                  # 提交前完整檢查
task clean                  # 清理編譯產物
```

## DDD 實踐指南

### Domain Entity

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

### Repository Interface (Domain Layer)

```go
// internal/domain/order/repository.go
package order

type Repository interface {
    Get(ctx context.Context, id string) (*Order, error)
    Save(ctx context.Context, order *Order) error
    Update(ctx context.Context, id string,
        updateFn func(ctx context.Context, o *Order) (*Order, error)) error
}
```

### Use Case (Application Layer)

```go
// internal/application/usecase/confirm_order.go
package usecase

type ConfirmOrderUseCase struct {
    repo     order.Repository
    notifier port.NotificationService
}

func (uc *ConfirmOrderUseCase) Execute(ctx context.Context, input dto.ConfirmOrderInput) error {
    return uc.repo.Update(ctx, input.OrderID,
        func(ctx context.Context, o *order.Order) (*order.Order, error) {
            if err := o.Confirm(); err != nil {
                return nil, err
            }
            if err := uc.notifier.SendConfirmation(ctx, o.UserID()); err != nil {
                return nil, err
            }
            return o, nil
        },
    )
}
```

### 設計原則

1. **Domain 物件欄位皆為 private** - 透過 Getter 存取
2. **Constructor 驗證** - 物件永遠處於有效狀態
3. **行為導向命名** - `Confirm()` 而非 `SetStatus()`
4. **Repository 介面在 Domain** - 依賴反轉
5. **Use Case 只做編排** - 業務邏輯放 Domain
6. **按聚合組織套件** - 相關程式碼放一起

## 測試策略

| 類型     | 位置                     | 說明               |
| -------- | ------------------------ | ------------------ |
| 單元測試 | `*_test.go` 與程式碼並列 | 測試單一函數/方法  |
| 整合測試 | `tests/integration/`     | 測試元件間整合     |
| E2E 測試 | `tests/e2e/`             | 測試完整使用者流程 |

- 使用 **Table-Driven Tests** 處理多種情境
- 遵循 **Arrange-Act-Assert** 模式
- Domain Layer 必須 100% 覆蓋

## 提交前檢查清單

執行 `task check` 會自動完成以下檢查：

- [ ] `task fmt` - 格式化程式碼
- [ ] `task tidy` - 整理依賴
- [ ] `task lint` - Lint 檢查通過
- [ ] `task test` - 全部測試通過

## 授權條款

詳見 [LICENSE](LICENSE)
