package identity

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/identity/adapter/http/handler"
	"github.com/blackhorseya/go-ddd/internal/identity/application/usecase"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/idgen"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence/memory"
)

// ProviderSet collects all providers for the Identity bounded context.
var ProviderSet = wire.NewSet(
	memory.NewCredentialRepository,
	idgen.NewUUIDGenerator,
	usecase.NewRegisterUseCase,
	handler.NewAuthHandler,
)
