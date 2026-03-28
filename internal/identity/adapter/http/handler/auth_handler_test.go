package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/application/usecase"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/internal/shared/adapter/http/response"
	"github.com/blackhorseya/go-ddd/internal/shared/domain/event"
)

type testDeps struct {
	router   *gin.Engine
	repo     *credential.MockRepository
	idGen    *port.MockIDGenerator
	tokenSvc *port.MockTokenService
	bus      *event.MockEventBus
}

func setupRouter(t *testing.T) testDeps {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := credential.NewMockRepository(ctrl)
	idGen := port.NewMockIDGenerator(ctrl)
	tokenSvc := port.NewMockTokenService(ctrl)
	bus := event.NewMockEventBus(ctrl)

	registerUC := usecase.NewRegisterUseCase(repo, idGen, bus)
	loginUC := usecase.NewLoginUseCase(repo, tokenSvc)
	refreshUC := usecase.NewRefreshTokenUseCase(repo, tokenSvc)
	h := NewAuthHandler(registerUC, loginUC, refreshUC)

	r := gin.New()
	h.Register(r)

	return testDeps{router: r, repo: repo, idGen: idGen, tokenSvc: tokenSvc, bus: bus}
}

func TestAuthHandler_RegisterCredential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deps := setupRouter(t)

		email, _ := credential.NewEmail("alice@example.com")
		deps.repo.EXPECT().FindByEmail(gomock.Any(), email).Return(nil, credential.ErrNotFound)
		deps.idGen.EXPECT().Generate().Return("c-1")
		deps.repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		deps.bus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

		body, _ := json.Marshal(dto.RegisterInput{
			Email:    "alice@example.com",
			Password: "secret123",
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}

		var resp response.Response
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Unmarshal response: %v", err)
		}

		data, ok := resp.Data.(map[string]any)
		if !ok {
			t.Fatalf("Data is not map[string]any: %T", resp.Data)
		}
		if data["id"] != "c-1" {
			t.Errorf("data.id = %v, want %q", data["id"], "c-1")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		deps := setupRouter(t)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("bad")))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		deps := setupRouter(t)

		email, _ := credential.NewEmail("alice@example.com")
		pwd, _ := credential.NewHashedPassword("secret123")
		existing, _ := credential.NewCredential(credential.NewCredentialParams{
			ID: "c-existing", Email: email, Password: pwd,
		})
		deps.repo.EXPECT().FindByEmail(gomock.Any(), email).Return(existing, nil)

		body, _ := json.Marshal(dto.RegisterInput{
			Email:    "alice@example.com",
			Password: "secret123",
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deps := setupRouter(t)

		email, _ := credential.NewEmail("alice@example.com")
		pwd, _ := credential.NewHashedPassword("secret123")
		cred, _ := credential.NewCredential(credential.NewCredentialParams{
			ID: "c-1", Email: email, Password: pwd,
		})
		deps.repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)
		deps.tokenSvc.EXPECT().GenerateTokenPair(gomock.Any(), gomock.Any()).
			Return(port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 900}, nil)

		body, _ := json.Marshal(dto.LoginInput{Email: "alice@example.com", Password: "secret123"})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		deps := setupRouter(t)

		email, _ := credential.NewEmail("alice@example.com")
		pwd, _ := credential.NewHashedPassword("secret123")
		cred, _ := credential.NewCredential(credential.NewCredentialParams{
			ID: "c-1", Email: email, Password: pwd,
		})
		deps.repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)

		body, _ := json.Marshal(dto.LoginInput{Email: "alice@example.com", Password: "wrong"})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		deps := setupRouter(t)

		deps.tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "valid-rt").
			Return(port.TokenClaims{CredentialID: "c-1", Email: "alice@example.com"}, nil)

		email, _ := credential.NewEmail("alice@example.com")
		pwd, _ := credential.NewHashedPassword("secret123")
		cred, _ := credential.NewCredential(credential.NewCredentialParams{
			ID: "c-1", Email: email, Password: pwd,
		})
		deps.repo.EXPECT().FindByID(gomock.Any(), "c-1").Return(cred, nil)
		deps.tokenSvc.EXPECT().GenerateTokenPair(gomock.Any(), gomock.Any()).
			Return(port.TokenPair{AccessToken: "new-at", RefreshToken: "new-rt", ExpiresIn: 900}, nil)

		body, _ := json.Marshal(dto.RefreshInput{RefreshToken: "valid-rt"})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		deps := setupRouter(t)

		deps.tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "bad").
			Return(port.TokenClaims{}, usecase.ErrInvalidCredentials)

		body, _ := json.Marshal(dto.RefreshInput{RefreshToken: "bad"})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		deps.router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
