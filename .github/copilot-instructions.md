# GitHub Copilot Custom Instructions

本專案採用 Golang、Clean Architecture 與 Domain-Driven Design（DDD）實作，並以 [Effective Go](https://go.dev/doc/effective_go) 為撰寫風格指引。以下列出團隊開發準則，請 Copilot 回應符合本專案的程式碼與建議。

---

## 語言與程式風格

- 使用 Go 1.21+
- 所有程式碼需經 `gofmt` 處理，統一使用 tab 做為縮排
- 命名遵循 MixedCaps，**不使用底線**
- Getter 方法命名不加 `Get` 前綴：`User()` 而非 `GetUser()`
- Interface 命名以 `-er` 結尾：例如 `Writer`, `Fetcher`, `Repository`
- 多個回傳值中的錯誤應命名為 `err`，早期返回處理錯誤
- 善用 blank identifier (`_`) 忽略不需要的返回值
- 控制結構（if, for, switch）不使用小括號

---

## 架構原則

- 採用 Clean Architecture：
  - **Domain**：純粹邏輯、entity、value object、interface（不含任何外部依賴）
  - **Application/Usecase**：執行商業流程，調用 domain 與 infrastructure
  - **Infrastructure**：處理外部依賴，如資料庫、API、GORM
  - **Interfaces/Delivery**：HTTP、gRPC 等 API 入口
- 採用 DDD：
  - 聚合邊界清晰，entity 自我封裝
  - 避免貧血模型，邏輯不可散落於 handler 或 repo 中

---

## 套件與檔案命名慣例

- 具體實作 struct 命名以 `Impl` 結尾（如 `UserRepositoryImpl`）
- 實作檔案命名為 `_impl.go`（如 `user_repository_impl.go`）
- 測試檔案為 `_test.go`
- 所有 public symbol 需有完整 doc comment

---

## Context 命名慣例

- 原生 `context.Context` 變數命名為 `c`
- 延伸 context（如 `contextx.Contextx`）命名為 `ctx`
- 起手式請使用 `contextx.WithContext(c)` 取得包含 logger 的擴展 context

---

## 相依注入與工具

- 使用 [`google/wire`](https://github.com/google/wire) 作為 DI 工具
  - 建議以模組分組 provider，避免手動組裝依賴
- MCP（Model Context Protocol）工具已整合：
  - 若功能涉及 metadata lookup、轉換或協作工具，請調用 MCP tools
  - Copilot 回覆時，請主動建議適用的工具並說明用途

---

## ORM 使用原則

- 可使用 GORM，但僅限於 infrastructure 層
- domain 層只能依賴定義於自身的 repository interface
- struct 上 GORM tag 僅限於 infra 層 struct，不可滲透至 domain

---

## 測試規範

- 所有商業邏輯需覆蓋單元測試
- infrastructure 可進行整合測試
- Mock 工具統一使用 [`uber-go/mock`](https://github.com/uber-go/mock)
  - **嚴禁使用已棄用的 `golang/mock`**

---

## Copilot 回應風格

請產出：
- idiomatic Go code，符合 [Effective Go](https://go.dev/doc/effective_go)
- 優先提供 interface 設計與 struct 實作起手式
- 每個程式碼建議應可測試、清晰、乾淨
- 如涉及外部資料或工具調用，請主動建議 MCP 使用方式
- 除非上下文明確，請勿生成過度樣板或框架型 code