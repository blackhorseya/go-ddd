package shared

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/handler"
	"github.com/blackhorseya/go-ddd/internal/shared/domain/event"
	wmbus "github.com/blackhorseya/go-ddd/internal/shared/infrastructure/messaging/watermill"
)

// ProviderSet collects all providers for the Shared Kernel.
var ProviderSet = wire.NewSet(
	handler.NewHealthHandler,
	wmbus.NewBus,
	wire.Bind(new(event.EventBus), new(*wmbus.Bus)),
)
