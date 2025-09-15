# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go DDD (Domain-Driven Design) template project implementing Clean Architecture principles. It serves as a starting point for building domain-driven microservices in Go.

## Architecture (Clean Architecture + DDD)

### 分層架構說明

專案遵循 Clean Architecture 和 Domain-Driven Design 原則，採用以下分層：

#### 1. Domain Layer (領域層) - 核心業務

- **職責**: 包含所有業務邏輯和規則
- **元件**:
  - `entity/`: 具有唯一標識的業務物件
  - `valueobject/`: 不可變的值物件
  - `aggregate/`: 聚合根，確保業務不變性
  - `repository/`: Repository 介面定義
  - `service/`: 領域服務（跨實體的業務邏輯）
  - `event/`: 領域事件
  - `specification/`: 業務規則規格
- **依賴**: 無外部依賴

#### 2. Application Layer (應用層) - 用例協調

- **職責**: 協調領域層完成用例
- **元件**:
  - `command/`: 命令處理器（寫入操作，CQRS）
  - `query/`: 查詢處理器（讀取操作，CQRS）
  - `usecase/`: 用例實作
  - `dto/`: 資料傳輸物件
  - `mapper/`: DTO 與領域物件轉換
- **依賴**: 僅依賴 Domain Layer

#### 3. Adapter Layer (適配器層) - 介面轉換

- **職責**: 處理外部請求，轉換為應用層可理解的格式
- **元件**:
  - `http/`: REST API 控制器與路由
  - `grpc/`: gRPC 服務定義與處理器
  - `consumer/`: 訊息佇列消費者
- **依賴**: Domain Layer、Application Layer

#### 4. Infrastructure Layer (基礎設施層) - 技術實作

- **職責**: 提供技術基礎設施
- **元件**:
  - `config/`: 配置管理
  - `database/`: 資料庫實作
  - `cache/`: 快取實作
  - `logger/`: 日誌實作
  - `monitoring/`: 監控實作
- **依賴**: 可依賴所有層（提供技術支援）

### 依賴規則

- 依賴方向：外層 → 內層
- Domain Layer 不依賴任何外層
- 使用介面進行依賴反轉

## Common Development Commands

### Building and Running

```bash
# Run the service
go run cmd/service/main.go

# Build the application
go build -o bin/service cmd/service/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestName ./path/to/package
```

### Code Quality

```bash
# Run linting (MUST pass before committing)
golangci-lint run

# Auto-fix linting issues
golangci-lint run --fix

# Format code
go fmt ./...
```

### Dependency Management

```bash
# Clean up dependencies (run after adding/removing imports)
go mod tidy

# Verify dependencies
go mod verify
```

### Code Generation

```bash
# Generate mocks (when interfaces change)
go generate ./...

# Generate Swagger docs (when APIs change)
swag init -g cmd/service/main.go
```

## Project Structure (Clean Architecture + DDD)

```text
.
├── cmd/                    # 應用程式進入點
│   └── service/           # 主要服務進入點
├── internal/              # 私有應用程式程式碼（遵循 DDD 分層）
│   ├── domain/            # 領域層（核心業務邏輯）
│   │   ├── entity/        # 實體
│   │   ├── valueobject/   # 值物件
│   │   ├── aggregate/     # 聚合根
│   │   ├── repository/    # Repository 介面
│   │   ├── service/       # 領域服務
│   │   ├── event/         # 領域事件
│   │   └── specification/ # 規格模式
│   ├── application/       # 應用層（用例協調）
│   │   ├── command/       # 命令處理器（寫入操作）
│   │   ├── query/         # 查詢處理器（讀取操作）
│   │   ├── usecase/       # 用例實作
│   │   ├── dto/           # 資料傳輸物件
│   │   └── mapper/        # DTO 與領域物件映射
│   ├── adapter/           # 適配器層（介面轉換）
│   │   ├── http/          # HTTP REST API 控制器
│   │   ├── grpc/          # gRPC 服務處理器
│   │   └── consumer/      # 訊息佇列消費者
│   └── infrastructure/    # 基礎設施層（技術實作）
│       ├── config/        # 配置管理
│       ├── persistence/   # 資料持久化實作
│       │   ├── postgres/  # PostgreSQL 實作
│       │   ├── mongodb/   # MongoDB 實作
│       │   └── redis/     # Redis 快取實作
│       ├── messaging/     # 訊息佇列實作
│       │   ├── kafka/     # Kafka 實作
│       │   └── rabbitmq/  # RabbitMQ 實作
│       ├── client/        # 外部服務客戶端
│       │   ├── http/      # HTTP 客戶端
│       │   └── grpc/      # gRPC 客戶端
│       ├── logger/        # 日誌記錄實作
│       └── monitoring/    # 監控與追蹤實作
├── pkg/                   # 公共可重用套件
├── configs/               # 配置檔案
├── deployments/           # 部署配置（Terraform, Helm）
├── tests/                 # 測試套件
│   ├── unit/             # 單元測試
│   ├── integration/      # 整合測試
│   └── e2e/              # 端對端測試
└── docs/                  # 專案文件
```

## Task Completion Requirements

Before considering any task complete:

1. Run `golangci-lint run` and fix all issues
2. Run `go test ./...` and ensure all tests pass
3. Run `go mod tidy` to clean up dependencies
4. Run `go fmt ./...` to format code

## Testing Strategy

- **Unit tests**: Place `*_test.go` files alongside the code being tested
- **Integration tests**: Place in `tests/integration/`
- **E2E tests**: Place in `tests/e2e/`
- Use table-driven tests for multiple test cases
- Follow Arrange-Act-Assert pattern

## DDD 實踐指南

### 領域建模原則

1. **聚合設計**
   - 每個聚合有唯一的聚合根
   - 聚合內部保持一致性
   - 聚合之間透過 ID 引用

2. **值物件設計**
   - 不可變性（Immutable）
   - 透過值比較而非引用比較
   - 沒有唯一標識

3. **實體設計**
   - 具有唯一標識
   - 生命週期管理
   - 包含業務行為

4. **領域服務**
   - 跨多個聚合的業務邏輯
   - 無狀態操作
   - 使用 Repository 介面

### 程式碼組織原則

1. **套件結構**
   - 按領域概念組織，而非技術層
   - 每個聚合一個套件
   - 共用的值物件放在 `valueobject/`

2. **依賴注入**
   - 使用介面定義依賴
   - 在 main.go 進行組裝
   - Repository 介面定義在 Domain 層

3. **錯誤處理**
   - 領域錯誤使用自定義類型
   - 基礎設施錯誤需要轉換
   - 保持錯誤資訊的業務語意

## Important Notes

- This is a template project - core domain implementation needs to be added
- Always maintain separation between layers (domain should not depend on infrastructure)
- Use interfaces for infrastructure concerns to maintain testability
- Follow CQRS pattern - separate commands (writes) from queries (reads)
- 遵循 DDD 戰術設計模式（Tactical Design Patterns）
- 使用 Ubiquitous Language（統一語言）命名
