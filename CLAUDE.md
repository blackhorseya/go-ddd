# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go DDD 範本專案，實作 Clean Architecture 與 Domain-Driven Design 原則。
Module: `github.com/blackhorseya/go-ddd`，Go 1.25+，使用 Gin (HTTP) + gRPC 雙協定。

## Architecture (BC-first + Clean Architecture + DDD)

### 組織方式：Bounded Context First

本專案以 **Bounded Context (BC)** 為第一層組織單位，每個 BC 內部再按 Clean Architecture 分層。
跨 BC 共用的元件放在 `shared/`（Shared Kernel）。

```text
internal/
├── shared/              # Shared Kernel（跨 BC 共用）
│   ├── domain/          # 共用領域概念（event, valueobject, pagination）
│   ├── adapter/         # 共用適配器（HTTP server, gRPC server, ACL）
│   └── infrastructure/  # 共用基礎設施（config, messaging）
├── {bc_name}/           # 各 Bounded Context（目前：identity）
│   ├── domain/          # 該 BC 的領域層
│   ├── application/     # 該 BC 的應用層
│   ├── adapter/         # 該 BC 的適配器層
│   └── infrastructure/  # 該 BC 的基礎設施層
pkg/                     # 公共可重用套件（contextx, logx, otelx）
scripts/
└── migrations/{bc}/     # 各 BC 的 SQL schema migration（golang-migrate，embed 進 binary）
```

### Clean Architecture 分層（每個 BC 內部）

```text
依賴方向：外層 → 內層（Domain 不依賴任何外層）

Adapter Layer        HTTP handler, gRPC, Consumer, ACL
Application Layer    Use Cases, DTOs, Ports
Domain Layer         Aggregates, Entities, Value Objects
Infrastructure Layer Persistence, Messaging, External Services
```

### 層級職責

| 層級 | 職責 | 元件 | 依賴限制 |
|------|------|------|----------|
| Domain | 業務邏輯與規則 | `{aggregate}/entity.go`, `repository.go`, `service.go`, `event.go`, `valueobject/` | 僅依賴 `shared/domain/` |
| Application | 協調領域層完成用例 | `usecase/`, `port/`, `dto/`, `mapper/` | 僅依賴 Domain Layer |
| Adapter | 處理外部請求 | `http/handler/`, `grpc/`, `consumer/`, `acl/{external_bc}/` | Domain + Application |
| Infrastructure | 技術實作 | `persistence/`, `external/` | 可依賴所有層 |

### Dependency Injection (Wire)

每個 BC 有 `wire.go` 暴露 `ProviderSet`，在 `cmd/service/wire.go` 的 `InitializeApp()` 中組裝：

```text
cmd/service/wire.go          ← wire.Build() 組裝所有 ProviderSet
internal/{bc}/wire.go        ← 該 BC 的 ProviderSet（handler, usecase, repo, port 實作）
internal/shared/wire.go      ← Shared Kernel 的 ProviderSet
```

修改 DI 後須執行 `task generate:wire` 重新產生 `wire_gen.go`。

### Domain Event 慣例

- 事件類型命名: `"{aggregate}.{past_tense_verb}"`（如 `"credential.registered"`）
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

## Development Commands

本專案使用 [Task](https://taskfile.dev/) 管理開發工作流程。執行 `task` 可查看所有可用任務。

### 常用指令

```bash
task run                    # 執行服務（go run ./cmd/service/）
task build                  # 編譯二進位檔案
task check                  # 提交前完整檢查（fmt → tidy → lint → test）
task dev                    # 開發檢查（lint + test）
```

### 測試

```bash
task test                   # 執行全部測試
task test:cover             # 測試覆蓋率
task test:race              # 競態檢測

# 執行單一測試（直接用 go test）
go test -run TestCredential_Validate ./internal/identity/domain/credential/...
```

### Code Quality

```bash
task lint                   # golangci-lint（提交前必須通過）
task lint:fix               # Lint 自動修復
task fmt                    # gofmt 格式化
task imports                # goimports 整理
```

### Code Generation

```bash
task generate               # 產生 Mocks（go generate ./...）
task generate:wire          # Wire 依賴注入（修改 wire.go 後必須執行）
task swagger                # Swagger 文件（輸出至 api/openapi/）
```

### Database Migration

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 作為 library（非 CLI），
SQL 放在 `scripts/migrations/{bc}/`，透過 `embed.FS` 打包進 binary。

```bash
task db:migrate:up                          # 套用所有待執行 migration
task db:migrate:version                     # 查看目前 schema 版本
task db:migrate:down                        # 全部回滾（破壞性，task 已帶 -yes）
task db:migrate:steps -- -1                 # 前進/回退 n 步
task db:migrate:force -- 1                  # 修復 dirty 狀態（不執行 SQL）
task db:migrate:create -- add_user_table    # 產生新的 up/down 檔案對

# 指定其他設定檔
task db:migrate:up CONFIG=configs/staging.yaml
```

- 連線參數取自設定檔的 `identity.database`，不在 Taskfile 重複硬寫 DSN
- Migration **不在服務啟動時自動執行**，須明確以上述指令觸發
- golang-migrate 的 CLI 將各 DB driver 藏在 build tag 後，`go.mod` 的 `tool` directive
  無法傳 tags，故不納入 tool 管理，改由 `cmd/migrate/` 這支自製 CLI 承擔

### 工具管理

開發工具（golangci-lint, wire, swag, mockgen）透過 `go.mod` 的 `tool` directive 管理，
以 `go tool <name>` 執行（如 `go tool golangci-lint run`），無需額外安裝。

## Code Conventions

### Import 排序

goimports local-prefixes: `github.com/blackhorseya/go-ddd`（三段式：stdlib / third-party / local）

### DDD 設計原則

1. **Domain 物件欄位皆為 private** - 透過 Getter 存取
2. **Constructor 驗證** - `New*()` 確保物件永遠處於有效狀態
3. **行為導向命名** - `Register()` 而非 `SetStatus()`
4. **Repository 介面在 Domain** - 依賴反轉
5. **Use Case 只做編排** - 業務邏輯放 Domain
6. **按聚合組織套件** - 一個聚合一個 package（如 `domain/credential/`）
7. **Mock 生成** - 使用 `go.uber.org/mock/mockgen`，mock 檔案與介面同目錄
8. **Repository 實作按技術分 package** - `persistence/postgres/`（正式）、`persistence/memory/`（測試與範例）
9. **驅動錯誤在 Repository 轉譯成 Domain error** - 如 unique violation → `credential.ErrEmailDuplicated`，
   判斷依據是 constraint 名稱而非只看 SQLSTATE，避免日後新增 unique 欄位時誤判

### 測試策略

| 類型     | 位置                     | 說明               |
| -------- | ------------------------ | ------------------ |
| 單元測試 | `*_test.go` 與程式碼並列 | 測試單一函數/方法  |
| 整合測試 | `tests/integration/`     | 測試元件間整合     |
| E2E 測試 | `tests/e2e/`             | 測試完整使用者流程 |

整合測試需要真實 PostgreSQL（`task test:integration`），連不上時會 skip 而非 fail，
所以 `task test` 在沒有本機 infra 時仍會全綠。連線設定走 `TEST_DATABASE_*` 環境變數，
預設連到獨立的 `identity_test` 資料庫（不存在時自動建立），不會動到開發用的 `identity`。

- 使用 Table-Driven Tests + Arrange-Act-Assert 模式
- Domain Layer 必須 100% 覆蓋
- 斷言使用 `github.com/stretchr/testify`

Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask.
