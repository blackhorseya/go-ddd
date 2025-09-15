# Architecture Guidelines - Clean Architecture + DDD

## 架構層次職責

### Domain Layer (領域層)
**位置**: `internal/domain/`
**職責**: 核心業務邏輯
**原則**:
- 無任何外部依賴
- 純粹的業務規則
- 不知道技術細節存在

**包含**:
- entity/: 業務實體
- valueobject/: 值物件
- aggregate/: 聚合根
- repository/: Repository 介面
- service/: 領域服務
- event/: 領域事件
- specification/: 業務規格

### Application Layer (應用層)
**位置**: `internal/application/`
**職責**: 用例協調
**原則**:
- 協調領域物件完成用例
- 不包含業務規則
- 實作 CQRS 模式

**包含**:
- command/: 寫入操作處理器
- query/: 讀取操作處理器
- usecase/: 用例實作
- dto/: 資料傳輸物件
- mapper/: 物件映射

### Adapter Layer (適配器層)
**位置**: `internal/adapter/`
**職責**: 介面轉換
**原則**:
- 轉換外部格式為內部格式
- 處理 HTTP/gRPC/訊息請求
- 呼叫應用層服務

**包含**:
- http/: REST API 控制器
- grpc/: gRPC 服務處理器
- consumer/: 訊息佇列消費者

### Infrastructure Layer (基礎設施層)
**位置**: `internal/infrastructure/`
**職責**: 技術實作
**原則**:
- 實作領域層定義的介面
- 處理所有技術細節
- 可依賴所有層

**包含**:
- persistence/: 資料庫實作
  - postgres/: PostgreSQL 實作
  - mongodb/: MongoDB 實作
  - redis/: Redis 快取
- messaging/: 訊息佇列實作
  - kafka/: Kafka 實作
  - rabbitmq/: RabbitMQ 實作
- client/: 外部服務客戶端
  - http/: HTTP 客戶端
  - grpc/: gRPC 客戶端
- config/: 配置管理
- logger/: 日誌實作
- monitoring/: 監控追蹤

## 依賴規則

1. **依賴方向**: 外層 → 內層
   - Infrastructure → Adapter → Application → Domain
   - Domain 層不依賴任何外層

2. **依賴反轉**:
   - Domain 定義介面
   - Infrastructure 實作介面
   - 透過依賴注入連接

3. **套件引用規則**:
   ```
   Domain: 不引用其他層
   Application: 只引用 Domain
   Adapter: 引用 Domain、Application
   Infrastructure: 可引用所有層
   ```

## 常見模式應用

### Repository 模式
- 介面定義: `internal/domain/repository/`
- 實作: `internal/infrastructure/persistence/`

### CQRS 模式
- Command: `internal/application/command/`
- Query: `internal/application/query/`
- 分離讀寫操作

### 聚合模式
- 每個聚合有唯一的聚合根
- 聚合內保持一致性
- 聚合間透過 ID 引用

### 領域事件
- 定義: `internal/domain/event/`
- 發布: Domain Service 或 Aggregate
- 處理: Application Layer 或 Infrastructure

## 檔案組織建議

```
internal/
├── domain/
│   └── order/              # Order 聚合
│       ├── entity.go       # Order 實體
│       ├── value.go        # 值物件
│       ├── repository.go   # Repository 介面
│       └── service.go      # 領域服務
├── application/
│   └── order/
│       ├── command/        # 命令處理
│       ├── query/          # 查詢處理
│       └── dto/            # DTO 定義
├── adapter/
│   └── http/
│       └── order/          # Order REST API
└── infrastructure/
    └── persistence/
        └── postgres/
            └── order.go    # Order Repository 實作
```

## 注意事項

1. 保持領域層純粹性
2. 避免貧血模型（Anemic Domain Model）
3. 使用 Ubiquitous Language
4. 每個聚合維護自己的一致性
5. 跨聚合操作使用領域服務或領域事件