# Code Style and Conventions

## Go Conventions
- Follow standard Go naming conventions (exported names start with capital letters)
- Use camelCase for variable and function names
- Use PascalCase for exported types and functions
- Keep package names lowercase and concise

## DDD-Specific Conventions

### Domain Layer
- **Entity**: 具有唯一標識的業務物件，包含業務行為
- **Value Object**: 不可變物件，透過值比較
- **Aggregate**: 聚合根確保業務一致性，聚合間透過 ID 引用
- **Repository**: 介面定義在 domain 層，實作在 infrastructure 層
- **Domain Service**: 跨聚合的無狀態業務邏輯
- **Domain Event**: 領域事件用於解耦和非同步處理
- **Specification**: 業務規則的封裝

### Application Layer
- **Command**: 寫入操作，改變系統狀態
- **Query**: 讀取操作，不改變系統狀態
- **Use Case**: 單一業務用例的協調邏輯
- **DTO**: 資料傳輸物件，用於層間傳遞
- **Mapper**: DTO 與領域物件的轉換

### Adapter Layer
- HTTP 控制器處理 REST API 請求
- gRPC 處理器實作 RPC 服務
- Consumer 處理訊息佇列消費

### Infrastructure Layer
- 所有技術相關實作
- Repository 實作
- 外部服務整合
- 技術框架配置

## Testing Conventions
- 單元測試放在同目錄的 `*_test.go`
- 整合測試在 `tests/integration/`
- E2E 測試在 `tests/e2e/`
- 使用 table-driven tests
- 遵循 Arrange-Act-Assert 模式

## Package Organization
- 按領域概念組織，而非技術層
- 每個聚合一個套件
- 共用值物件放在 `valueobject/`
- 使用 Ubiquitous Language 命名

## Error Handling
- 領域錯誤使用自定義類型
- 基礎設施錯誤需要轉換為領域錯誤
- 保持錯誤資訊的業務語意

## Dependency Injection
- 使用介面定義依賴
- 在 cmd/service/main.go 進行依賴組裝
- Repository 介面定義在 Domain 層，實作在 Infrastructure 層