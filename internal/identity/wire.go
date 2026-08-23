package identity

import (
	"github.com/google/wire"

	"github.com/blackhorseya/go-ddd/internal/identity/adapter/http/handler"
	"github.com/blackhorseya/go-ddd/internal/identity/application/usecase"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/auth"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/idgen"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence/postgres"
)

// ProviderSet collects all providers for the Identity bounded context.
var ProviderSet = wire.NewSet(
	persistence.NewDB,
	postgres.NewCredentialRepository,
	idgen.NewUUIDGenerator,
	auth.NewJWTTokenService,
	auth.NewTokenValidatorAdapter,
	usecase.NewRegisterUseCase,
	usecase.NewLoginUseCase,
	usecase.NewRefreshTokenUseCase,
	handler.NewAuthHandler,
)
