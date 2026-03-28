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
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
)

func setupRouter(t *testing.T) (*gin.Engine, *user.MockRepository, *port.MockIDGenerator) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := user.NewMockRepository(ctrl)
	idGen := port.NewMockIDGenerator(ctrl)

	registerUC := usecase.NewRegisterUserUseCase(repo, idGen)
	h := NewUserHandler(registerUC)

	r := gin.New()
	h.Register(r)

	return r, repo, idGen
}

func TestUserHandler_RegisterUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, repo, idGen := setupRouter(t)

		idGen.EXPECT().Generate().Return("u-1")
		repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		body, _ := json.Marshal(dto.RegisterUserInput{
			Email:    "alice@example.com",
			Password: "secret123",
			Name:     "Alice",
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		}

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		if resp["success"] != true {
			t.Errorf("success = %v, want true", resp["success"])
		}

		data, _ := resp["data"].(map[string]any)
		if data["id"] != "u-1" {
			t.Errorf("data.id = %v, want %q", data["id"], "u-1")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		r, _, _ := setupRouter(t)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte("bad")))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		r, repo, idGen := setupRouter(t)

		idGen.EXPECT().Generate().Return("u-2")
		repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(user.ErrEmailDuplicated)

		body, _ := json.Marshal(dto.RegisterUserInput{
			Email:    "alice@example.com",
			Password: "secret123",
			Name:     "Alice",
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		r, _, _ := setupRouter(t)

		body, _ := json.Marshal(dto.RegisterUserInput{
			Email:    "bad",
			Password: "secret123",
			Name:     "Alice",
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
