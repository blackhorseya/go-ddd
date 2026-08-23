# Go DDD Template Project

Go 語言實作 **Bounded Context First + Clean Architecture + Domain-Driven Design (DDD)** 的專案範本。

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

# 5. 複製設定檔
cp configs/config.example.yaml configs/config.yaml

# 6. 安裝依賴
task tidy

# 7. 執行服務（HTTP :8080, gRPC :9090）
task run
```

## 架構概覽

採用 **Bounded Context (BC) First** 組織方式：每個 BC 為獨立頂層目錄，內部按 Clean Architecture 分層。

```text
internal/
├── shared/                         # Shared Kernel（跨 BC 共用）
│   ├── domain/                     # 共用領域概念（event, valueobject, pagination）
│   ├── adapter/                    # HTTP/gRPC server, router, middleware
│   └── infrastructure/             # config, messaging
├── identity/                       # Identity BC（認證）
│   ├── domain/credential/          # Credential 聚合（aggregate root）
│   ├── application/                # usecase, dto, port, mapper
│   ├── adapter/http/handler/       # REST API handler
│   └── infrastructure/             # persistence, idgen
pkg/                                # 公共可重用套件（contextx, logx, otelx）
scripts/
└── migrations/{bc}/                # 各 BC 的 SQL schema migration（embed 進 binary）
```

### Clean Architecture 分層

```text
依賴方向：外層 → 內層（Domain 不依賴任何外層）

Adapter Layer        ← HTTP handler, gRPC, Consumer, ACL
Application Layer    ← Use Cases, DTOs, Ports
Domain Layer         ← Aggregates, Entities, Value Objects
Infrastructure Layer ← Persistence, Messaging, External Services
```

| 層級               | 職責                   | 依賴限制               |
| ------------------ | ---------------------- | ---------------------- |
| **Domain**         | 核心業務邏輯與規則     | 僅依賴 `shared/domain` |
| **Application**    | 用例編排，協調領域層   | 僅依賴 Domain Layer    |
| **Adapter**        | 處理外部請求，轉換格式 | Domain + Application   |
| **Infrastructure** | 技術基礎設施實作       | 可依賴所有層           |

### Dependency Injection

使用 [Wire](https://github.com/google/wire) 做編譯期依賴注入。每個 BC 暴露 `ProviderSet`，在 `cmd/service/wire.go` 組裝：

```text
cmd/service/wire.go          ← wire.Build() 組裝所有 ProviderSet
internal/{bc}/wire.go        ← 該 BC 的 ProviderSet
internal/shared/wire.go      ← Shared Kernel 的 ProviderSet
```

## DDD 實踐指南

### Domain Entity（聚合根）

```go
// internal/identity/domain/credential/credential.go
package credential

type Credential struct {
    id        string         // private fields
    email     Email          // Value Object
    password  HashedPassword // Value Object
    status    Status
    createdAt time.Time
    updatedAt time.Time
}

// Constructor — 確保物件永遠有效
func NewCredential(params NewCredentialParams) (*Credential, error) {
    if params.ID == "" {
        return nil, ErrEmptyID
    }
    // ...validation...
    return &Credential{
        id:     params.ID,
        email:  params.Email,
        status: StatusInactive,
    }, nil
}

// Behavior — 方法名稱反映業務行為，內含狀態轉換守衛
func (x *Credential) Activate() error {
    if x.status != StatusInactive {
        return ErrCannotActivate
    }
    x.status = StatusActive
    return nil
}

// Reconstitute — 從持久層還原，不套用業務規則
func ReconstituteCredential(params ReconstituteCredentialParams) (*Credential, error) { ... }
```

### Repository Interface（定義在 Domain Layer）

```go
// internal/identity/domain/credential/repository.go
//go:generate go tool mockgen -destination=mock_${GOFILE} -package=${GOPACKAGE} -source=${GOFILE}

type Repository interface {
    Save(c context.Context, cred *Credential) error
    FindByID(c context.Context, id string) (*Credential, error)
    FindByEmail(c context.Context, email Email) (*Credential, error)
    Delete(c context.Context, id string) error
    List(c context.Context, req domain.PageRequest) (domain.PageResult[*Credential], error)
}
```

### Use Case（Application Layer）

```go
// internal/identity/application/usecase/register_credential.go
type RegisterUseCase struct {
    repo  credential.Repository
    idGen port.IDGenerator
}

func (uc *RegisterUseCase) Execute(c context.Context, input dto.RegisterInput) (dto.CredentialOutput, error) {
    // 1. 建立 Value Objects
    // 2. 建立 Aggregate Root（NewCredential）
    // 3. 透過 Repository 持久化
    // 4. 透過 Mapper 轉換為 DTO 回傳
}
```

### 設計原則

1. **Domain 物件欄位皆為 private** — 透過 Getter 存取
2. **Constructor 驗證** — `New*()` 確保物件永遠處於有效狀態
3. **行為導向命名** — `Activate()` 而非 `SetStatus()`
4. **Repository 介面在 Domain** — 依賴反轉（Dependency Inversion）
5. **Use Case 只做編排** — 業務邏輯放 Domain
6. **按聚合組織套件** — 一個聚合一個 package

## 開發指令

本專案使用 [Task](https://taskfile.dev/) 管理開發工作流程。

```bash
# 建置與執行
task run                    # 執行服務
task build                  # 編譯二進位檔案

# 測試
task test                   # 執行全部測試
task test:cover             # 測試覆蓋率
task test:race              # 競態檢測
task test:integration       # 整合測試（需 PostgreSQL，連不上則 skip）

# 程式碼品質
task lint                   # golangci-lint 檢查
task lint:fix               # Lint 自動修復
task fmt                    # 格式化程式碼

# 程式碼產生
task generate               # 產生 Mocks
task generate:wire          # Wire 依賴注入
task swagger                # Swagger 文件（輸出至 api/openapi/）

# 資料庫 Migration
task db:migrate:up                          # 套用所有待執行 migration
task db:migrate:version                     # 查看目前 schema 版本
task db:migrate:create -- add_user_table    # 產生新的 up/down 檔案對

# 開發工作流程
task dev                    # 開發檢查（lint + test）
task check                  # 提交前完整檢查（fmt → tidy → lint → test）
```

### 工具管理

開發工具透過 `go.mod` 的 `tool` directive 管理，以 `go tool <name>` 執行，無需額外安裝：

- `golangci-lint` — Lint
- `wire` — 依賴注入
- `swag` — Swagger 文件產生
- `mockgen` — Mock 產生

## 設定

設定檔：`configs/config.yaml`（參考 `configs/config.example.yaml`）

也可透過命令列旗標指定路徑：

```bash
go run ./cmd/service/ -config ./configs/config.yaml
```

支援環境變數覆蓋（Viper），如 `APP_NAME=myservice`。

## 資料庫 Migration

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 作為 **library**（非 CLI）。
SQL 檔放在 `scripts/migrations/{bc}/`，由該目錄的 `embed.go` 以 `embed.FS` 打包進 binary，
部署時不需額外掛載 SQL 檔。

```bash
task infra:up                               # 啟動本地 PostgreSQL
task db:migrate:up                          # 套用所有待執行 migration
task db:migrate:version                     # 查看目前 schema 版本
task db:migrate:steps -- -1                 # 回退一步
task db:migrate:down                        # 全部回滾（破壞性）
task db:migrate:force -- 1                  # 修復 dirty 狀態（不執行 SQL）
task db:migrate:create -- add_user_table    # 產生新的 up/down 檔案對

# 指定其他設定檔
task db:migrate:up CONFIG=configs/staging.yaml
```

連線參數取自設定檔的 `identity.database` 區段。Migration **不會在服務啟動時自動執行**，
須明確透過上述指令（或 `cmd/migrate/` binary）觸發，讓 schema 變更與服務部署解耦。

檔案命名遵循 golang-migrate 慣例：`{6 位序號}_{描述}.{up|down}.sql`。

發版時 `migrate` 有自己的 archive（`go-ddd-migrate_{version}_{os}_{arch}.tar.gz`），
只跑 migration 的環境不必下載整個 service。同時它也被打包進 service 的 Docker image，
可在 Kubernetes 以 initContainer 執行，確保 migration 與服務版本一致：

```yaml
initContainers:
  - name: migrate
    image: ghcr.io/blackhorseya/go-ddd:1.2.0
    command: ["/app/migrate", "-config", "/etc/go-ddd/config.yaml", "up"]
```

### 在 AWS Lambda 執行

同一支 binary 也能當 Lambda function：它偵測到 `AWS_LAMBDA_RUNTIME_API` 就切換成
Runtime API 模式，套用所有待執行的 migration 後回報結果。

```bash
serverless invoke -f migrate --stage dev
# {"version":1,"dirty":false,"message":"migrations applied"}
```

- **只支援 `up`**：`down` / `force` 這類破壞性操作在 CLI 有 `-yes` 把關，
  而 invocation 沒有對等的確認機制，因此不開放。函式不接受任何 event payload。
- 設定走 `APP_IDENTITY_DATABASE_*` 環境變數（部署包裡沒有設定檔），
  憑證建議由 SSM / Secrets Manager 注入。
- Lambda 逾時上限 15 分鐘，超大 migration 不適合這條路；
  函式需具備連到 RDS 的 VPC 權限。
- 併發呼叫是安全的 —— golang-migrate 會取得 PostgreSQL advisory lock。

## 持久層

Identity BC 的 `credential.Repository` 有兩個實作：

| 實作 | 位置 | 用途 |
| ---- | ---- | ---- |
| PostgreSQL | `internal/identity/infrastructure/persistence/postgres/` | 正式實作，由 Wire 注入 |
| In-memory | `internal/identity/infrastructure/persistence/memory/` | 單元測試與範例 |

PostgreSQL 實作以 `database/sql` + `pgx/v5` stdlib driver 撰寫，並將驅動錯誤轉譯回領域錯誤
（unique violation → `ErrEmailDuplicated`、`sql.ErrNoRows` → `ErrNotFound`），
讓 Application 層只需要處理 domain error。

服務啟動時會 ping 資料庫，**連不上即啟動失敗**（fail-fast），避免問題延遲到第一個請求才浮現。

## 測試策略

| 類型     | 位置                     | 說明               |
| -------- | ------------------------ | ------------------ |
| 單元測試 | `*_test.go` 與程式碼並列 | 測試單一函數/方法  |
| 整合測試 | `tests/integration/`     | 測試元件間整合     |
| E2E 測試 | `tests/e2e/`             | 測試完整使用者流程 |

- 使用 **Table-Driven Tests** + **Arrange-Act-Assert** 模式
- Domain Layer 必須 100% 覆蓋
- Mock 使用 `go.uber.org/mock`，mock 檔案與介面同目錄

## 部署

```bash
# GoReleaser（本機測試）
task release:snapshot

# Docker 映像
task image

# AWS Lambda
task deploy                 # 預設 dev stage
task deploy STAGE=prod      # 指定 stage
```

## 授權條款

GPL-3.0 — 詳見 [LICENSE](LICENSE)
