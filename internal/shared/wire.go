package shared

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/handler"
)

// ProviderSet collects all providers for the Shared Kernel.
var ProviderSet = wire.NewSet(
	handler.NewHealthHandler,
)
