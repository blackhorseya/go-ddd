package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blackhorseya/go-ddd/internal/adapter/http/response"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// Add request ID to context
	ctx := contextx.WithRequestID(c.Request.Context(), "test-request-id")
	c.Request = c.Request.WithContext(ctx)

	return c, w
}

func TestOK(t *testing.T) {
	c, w := setupTestContext()

	data := map[string]string{"message": "hello"}
	response.OK(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "test-request-id", resp.Meta.RequestID)
	assert.False(t, resp.Meta.Timestamp.IsZero())
}

func TestCreated(t *testing.T) {
	c, w := setupTestContext()

	data := map[string]string{"id": "123"}
	response.Created(c, data)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
}

func TestNoContent(t *testing.T) {
	// Create a real test router for NoContent since c.Status() alone
	// doesn't finalize the status code in httptest context
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		response.NoContent(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestList(t *testing.T) {
	c, w := setupTestContext()

	items := []string{"a", "b", "c"}
	response.List(c, items, 1, 10, 25)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Success)
	require.NotNil(t, resp.Meta.Pagination)
	assert.Equal(t, 1, resp.Meta.Pagination.Page)
	assert.Equal(t, 10, resp.Meta.Pagination.PageSize)
	assert.Equal(t, 25, resp.Meta.Pagination.Total)
	assert.Equal(t, 3, resp.Meta.Pagination.TotalPages)
}

func TestList_ZeroPageSize(t *testing.T) {
	c, w := setupTestContext()

	items := []string{"a"}
	response.List(c, items, 1, 0, 10)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 0, resp.Meta.Pagination.TotalPages)
}

func TestErr(t *testing.T) {
	c, w := setupTestContext()

	response.Err(c, http.StatusBadRequest, "INVALID_INPUT", "invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
	assert.Equal(t, "invalid input", resp.Error.Message)
}

func TestErrWithDetails(t *testing.T) {
	c, w := setupTestContext()

	details := []response.FieldError{
		{Field: "email", Message: "invalid email format"},
		{Field: "age", Message: "must be positive"},
	}
	response.ErrWithDetails(c, http.StatusBadRequest, response.CodeValidationFailed, "validation failed", details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, response.CodeValidationFailed, resp.Error.Code)
	require.Len(t, resp.Error.Details, 2)
	assert.Equal(t, "email", resp.Error.Details[0].Field)
}

func TestConvenienceErrors(t *testing.T) {
	tests := []struct {
		name       string
		callFunc   func(c *gin.Context)
		wantStatus int
		wantCode   string
	}{
		{
			name:       "BadRequest",
			callFunc:   func(c *gin.Context) { response.BadRequest(c, "bad request") },
			wantStatus: http.StatusBadRequest,
			wantCode:   response.CodeBadRequest,
		},
		{
			name:       "Unauthorized",
			callFunc:   func(c *gin.Context) { response.Unauthorized(c, "unauthorized") },
			wantStatus: http.StatusUnauthorized,
			wantCode:   response.CodeUnauthorized,
		},
		{
			name:       "Forbidden",
			callFunc:   func(c *gin.Context) { response.Forbidden(c, "forbidden") },
			wantStatus: http.StatusForbidden,
			wantCode:   response.CodeForbidden,
		},
		{
			name:       "NotFound",
			callFunc:   func(c *gin.Context) { response.NotFound(c, "not found") },
			wantStatus: http.StatusNotFound,
			wantCode:   response.CodeNotFound,
		},
		{
			name:       "Conflict",
			callFunc:   func(c *gin.Context) { response.Conflict(c, "conflict") },
			wantStatus: http.StatusConflict,
			wantCode:   response.CodeConflict,
		},
		{
			name:       "TooManyRequests",
			callFunc:   func(c *gin.Context) { response.TooManyRequests(c, "rate limited") },
			wantStatus: http.StatusTooManyRequests,
			wantCode:   response.CodeTooManyRequests,
		},
		{
			name:       "InternalError",
			callFunc:   func(c *gin.Context) { response.InternalError(c, "internal error") },
			wantStatus: http.StatusInternalServerError,
			wantCode:   response.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext()

			tt.callFunc(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp response.Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.False(t, resp.Success)
			require.NotNil(t, resp.Error)
			assert.Equal(t, tt.wantCode, resp.Error.Code)
		})
	}
}

func TestValidationFailed(t *testing.T) {
	c, w := setupTestContext()

	details := []response.FieldError{
		{Field: "name", Message: "required"},
	}
	response.ValidationFailed(c, details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	assert.Equal(t, response.CodeValidationFailed, resp.Error.Code)
	assert.Len(t, resp.Error.Details, 1)
}
