# Go DDD Template Project Overview

## Purpose
This is a template project implementing Clean Architecture and Domain-Driven Design (DDD) principles in Go. It serves as a starting point for domain-driven microservices.

## Tech Stack
- **Language**: Go 1.24.6
- **Module**: github.com/blackhorseya/go-ddd
- **Key Dependencies**:
  - golangci-lint - Code linting
  - swag - API documentation generation
  - go.uber.org/mock - Mock generation for testing

## Architecture (Clean Architecture + DDD)

專案採用四層架構：

### 1. Domain Layer (領域層)
- 核心業務邏輯，無外部依賴
- 包含：entity、valueobject、aggregate、repository介面、service、event、specification

### 2. Application Layer (應用層)
- 協調領域層完成用例
- 包含：command（寫）、query（讀）、usecase、dto、mapper
- 實作 CQRS 模式

### 3. Adapter Layer (適配器層)
- 處理外部請求，轉換格式
- 包含：http（REST API）、grpc（服務）、consumer（訊息消費）

### 4. Infrastructure Layer (基礎設施層)
- 所有技術實作
- 包含：
  - persistence（資料庫實作：postgres、mongodb、redis）
  - messaging（訊息佇列：kafka、rabbitmq）
  - client（外部服務客戶端）
  - config、logger、monitoring

## Project Structure
```
├── cmd/            # 應用程式進入點
├── internal/       # 私有應用程式程式碼
│   ├── domain/     # 領域層
│   ├── application/# 應用層
│   ├── adapter/    # 適配器層
│   └── infrastructure/ # 基礎設施層
├── pkg/            # 公共可重用套件
├── configs/        # 配置檔案
├── deployments/    # IaC和部署配置
├── tests/          # 測試套件
└── docs/           # 專案文件
```

## Key Design Principles
1. 依賴方向：外層 → 內層
2. Domain Layer 不依賴任何外層
3. 使用介面進行依賴反轉
4. CQRS 分離命令與查詢
5. 領域事件驅動跨邊界通訊