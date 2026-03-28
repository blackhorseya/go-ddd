// Package acl provides the Anti-Corruption Layer (ACL) pattern for translating
// external bounded context models into the local domain model.
//
// The ACL prevents external domain concepts from leaking into the local bounded context.
// Each external BC gets its own sub-package under the consuming BC's adapter/acl/ directory.
//
// Architecture:
//
//	┌──────────────────────────────────────────────────────────┐
//	│ Application Layer                                        │
//	│   port/{external}.go         ← defines what we need      │
//	├──────────────────────────────────────────────────────────┤
//	│ Adapter Layer                                            │
//	│   acl/{external_bc}/                                     │
//	│     translator.go            ← translates models         │
//	│     client.go                ← defines client interface  │
//	├──────────────────────────────────────────────────────────┤
//	│ Infrastructure Layer                                     │
//	│   external/{service}/                                    │
//	│     client.go                ← implements raw client     │
//	└──────────────────────────────────────────────────────────┘
//
// The ACL adapter implements the application port interface and depends on
// a client interface it defines locally (dependency inversion).
//
// Example structure for an Order BC consuming a Payment BC:
//
//	internal/order/
//	├── application/port/payment.go        # PaymentService interface
//	├── adapter/acl/payment/
//	│   ├── translator.go                  # Implements port.PaymentService
//	│   └── client.go                      # Defines PaymentClient interface
//	└── infrastructure/external/payment/
//	    └── client.go                      # Implements PaymentClient (HTTP/gRPC)
//
// This shared/adapter/acl/ package serves as documentation only.
// Each BC should create its own acl/ sub-packages as needed.
package acl
