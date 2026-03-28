package domain

import (
	"testing"
)

// ============================================================================
// SortOption Tests
// ============================================================================

func TestNewSortOption(t *testing.T) {
	tests := []struct {
		name              string
		field             string
		direction         SortDirection
		expectedDirection SortDirection
	}{
		{
			name:              "ascending direction",
			field:             "created_at",
			direction:         SortAsc,
			expectedDirection: SortAsc,
		},
		{
			name:              "descending direction",
			field:             "updated_at",
			direction:         SortDesc,
			expectedDirection: SortDesc,
		},
		{
			name:              "invalid direction defaults to asc",
			field:             "name",
			direction:         "invalid",
			expectedDirection: SortAsc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			opt := NewSortOption(tt.field, tt.direction)

			// Assert
			if opt.Field() != tt.field {
				t.Errorf("Field() = %v, want %v", opt.Field(), tt.field)
			}
			if opt.Direction() != tt.expectedDirection {
				t.Errorf("Direction() = %v, want %v", opt.Direction(), tt.expectedDirection)
			}
		})
	}
}

func TestSortOption_IsAscending(t *testing.T) {
	tests := []struct {
		name      string
		direction SortDirection
		want      bool
	}{
		{"asc returns true", SortAsc, true},
		{"desc returns false", SortDesc, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := NewSortOption("field", tt.direction)
			if got := opt.IsAscending(); got != tt.want {
				t.Errorf("IsAscending() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// PageRequest Tests
// ============================================================================

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  error
		wantPage int
		wantSize int
	}{
		{
			name:     "valid request",
			page:     1,
			pageSize: 20,
			wantErr:  nil,
			wantPage: 1,
			wantSize: 20,
		},
		{
			name:     "page at boundary",
			page:     1,
			pageSize: MaxPageSize,
			wantErr:  nil,
			wantPage: 1,
			wantSize: MaxPageSize,
		},
		{
			name:     "page zero is invalid",
			page:     0,
			pageSize: 20,
			wantErr:  ErrInvalidPage,
		},
		{
			name:     "negative page is invalid",
			page:     -1,
			pageSize: 20,
			wantErr:  ErrInvalidPage,
		},
		{
			name:     "page size zero is invalid",
			page:     1,
			pageSize: 0,
			wantErr:  ErrInvalidPageSize,
		},
		{
			name:     "page size exceeds max",
			page:     1,
			pageSize: MaxPageSize + 1,
			wantErr:  ErrInvalidPageSize,
		},
		{
			name:     "negative page size is invalid",
			page:     1,
			pageSize: -1,
			wantErr:  ErrInvalidPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			req, err := NewPageRequest(tt.page, tt.pageSize)

			// Assert
			if err != tt.wantErr {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				if req.Page() != tt.wantPage {
					t.Errorf("Page() = %v, want %v", req.Page(), tt.wantPage)
				}
				if req.PageSize() != tt.wantSize {
					t.Errorf("PageSize() = %v, want %v", req.PageSize(), tt.wantSize)
				}
			}
		})
	}
}

func TestNewPageRequestWithDefaults(t *testing.T) {
	// Act
	req := NewPageRequestWithDefaults()

	// Assert
	if req.Page() != DefaultPage {
		t.Errorf("Page() = %v, want %v", req.Page(), DefaultPage)
	}
	if req.PageSize() != DefaultPageSize {
		t.Errorf("PageSize() = %v, want %v", req.PageSize(), DefaultPageSize)
	}
}

func TestPageRequest_Offset(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		wantOffset int
	}{
		{"page 1", 1, 20, 0},
		{"page 2", 2, 20, 20},
		{"page 3 with size 10", 3, 10, 20},
		{"page 5 with size 50", 5, 50, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := NewPageRequest(tt.page, tt.pageSize)
			if got := req.Offset(); got != tt.wantOffset {
				t.Errorf("Offset() = %v, want %v", got, tt.wantOffset)
			}
		})
	}
}

func TestPageRequest_WithSort(t *testing.T) {
	// Arrange
	req, _ := NewPageRequest(1, 20)
	sort1 := NewSortOption("created_at", SortDesc)
	sort2 := NewSortOption("id", SortAsc)

	// Act
	newReq := req.WithSort(sort1, sort2)

	// Assert - original unchanged (immutability)
	if len(req.Sort()) != 0 {
		t.Error("original request should not be modified")
	}

	// Assert - new request has sort options
	if len(newReq.Sort()) != 2 {
		t.Errorf("Sort() length = %v, want 2", len(newReq.Sort()))
	}
	if newReq.Sort()[0].Field() != "created_at" {
		t.Errorf("Sort()[0].Field() = %v, want created_at", newReq.Sort()[0].Field())
	}
}

// ============================================================================
// PageResult Tests
// ============================================================================

func TestNewPageResult(t *testing.T) {
	tests := []struct {
		name           string
		itemCount      int
		page           int
		pageSize       int
		totalItems     int64
		wantTotalPages int
		wantHasNext    bool
		wantHasPrev    bool
	}{
		{
			name:           "first page of multiple",
			itemCount:      20,
			page:           1,
			pageSize:       20,
			totalItems:     50,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    false,
		},
		{
			name:           "middle page",
			itemCount:      20,
			page:           2,
			pageSize:       20,
			totalItems:     50,
			wantTotalPages: 3,
			wantHasNext:    true,
			wantHasPrev:    true,
		},
		{
			name:           "last page",
			itemCount:      10,
			page:           3,
			pageSize:       20,
			totalItems:     50,
			wantTotalPages: 3,
			wantHasNext:    false,
			wantHasPrev:    true,
		},
		{
			name:           "single page",
			itemCount:      5,
			page:           1,
			pageSize:       20,
			totalItems:     5,
			wantTotalPages: 1,
			wantHasNext:    false,
			wantHasPrev:    false,
		},
		{
			name:           "exact page boundary",
			itemCount:      20,
			page:           2,
			pageSize:       20,
			totalItems:     40,
			wantTotalPages: 2,
			wantHasNext:    false,
			wantHasPrev:    true,
		},
		{
			name:           "empty result",
			itemCount:      0,
			page:           1,
			pageSize:       20,
			totalItems:     0,
			wantTotalPages: 0,
			wantHasNext:    false,
			wantHasPrev:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			items := make([]int, tt.itemCount)

			// Act
			result := NewPageResult(items, tt.page, tt.pageSize, tt.totalItems)

			// Assert
			if result.TotalPages() != tt.wantTotalPages {
				t.Errorf("TotalPages() = %v, want %v", result.TotalPages(), tt.wantTotalPages)
			}
			if result.HasNext() != tt.wantHasNext {
				t.Errorf("HasNext() = %v, want %v", result.HasNext(), tt.wantHasNext)
			}
			if result.HasPrev() != tt.wantHasPrev {
				t.Errorf("HasPrev() = %v, want %v", result.HasPrev(), tt.wantHasPrev)
			}
			if result.Page() != tt.page {
				t.Errorf("Page() = %v, want %v", result.Page(), tt.page)
			}
			if result.TotalItems() != tt.totalItems {
				t.Errorf("TotalItems() = %v, want %v", result.TotalItems(), tt.totalItems)
			}
		})
	}
}

func TestPageResult_IsEmpty(t *testing.T) {
	t.Run("empty items", func(t *testing.T) {
		result := NewPageResult([]int{}, 1, 20, 0)
		if !result.IsEmpty() {
			t.Error("IsEmpty() should return true for empty items")
		}
	})

	t.Run("non-empty items", func(t *testing.T) {
		result := NewPageResult([]int{1, 2, 3}, 1, 20, 3)
		if result.IsEmpty() {
			t.Error("IsEmpty() should return false for non-empty items")
		}
	})
}

func TestEmptyPageResult(t *testing.T) {
	result := EmptyPageResult[string]()

	if !result.IsEmpty() {
		t.Error("EmptyPageResult should be empty")
	}
	if result.TotalItems() != 0 {
		t.Errorf("TotalItems() = %v, want 0", result.TotalItems())
	}
	if result.TotalPages() != 0 {
		t.Errorf("TotalPages() = %v, want 0", result.TotalPages())
	}
}

// ============================================================================
// CursorRequest Tests
// ============================================================================

func TestNewCursorRequest(t *testing.T) {
	tests := []struct {
		name     string
		cursor   string
		pageSize int
		wantErr  error
	}{
		{
			name:     "valid request without cursor",
			cursor:   "",
			pageSize: 20,
			wantErr:  nil,
		},
		{
			name:     "valid request with cursor",
			cursor:   "abc123",
			pageSize: 20,
			wantErr:  nil,
		},
		{
			name:     "page size at max",
			cursor:   "",
			pageSize: MaxPageSize,
			wantErr:  nil,
		},
		{
			name:     "page size zero is invalid",
			cursor:   "",
			pageSize: 0,
			wantErr:  ErrInvalidPageSize,
		},
		{
			name:     "page size exceeds max",
			cursor:   "",
			pageSize: MaxPageSize + 1,
			wantErr:  ErrInvalidPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewCursorRequest(tt.cursor, tt.pageSize)

			if err != tt.wantErr {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == nil {
				if req.Cursor() != tt.cursor {
					t.Errorf("Cursor() = %v, want %v", req.Cursor(), tt.cursor)
				}
				if req.PageSize() != tt.pageSize {
					t.Errorf("PageSize() = %v, want %v", req.PageSize(), tt.pageSize)
				}
			}
		})
	}
}

func TestCursorRequest_HasCursor(t *testing.T) {
	tests := []struct {
		name   string
		cursor string
		want   bool
	}{
		{"empty cursor", "", false},
		{"non-empty cursor", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := NewCursorRequest(tt.cursor, 20)
			if got := req.HasCursor(); got != tt.want {
				t.Errorf("HasCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCursorRequest_WithSort(t *testing.T) {
	// Arrange
	req, _ := NewCursorRequest("cursor", 20)
	sort := NewSortOption("created_at", SortDesc)

	// Act
	newReq := req.WithSort(sort)

	// Assert - original unchanged
	if len(req.Sort()) != 0 {
		t.Error("original request should not be modified")
	}

	// Assert - new request has sort
	if len(newReq.Sort()) != 1 {
		t.Errorf("Sort() length = %v, want 1", len(newReq.Sort()))
	}
}

// ============================================================================
// CursorResult Tests
// ============================================================================

func TestNewCursorResult(t *testing.T) {
	// Arrange
	items := []string{"a", "b", "c"}

	// Act
	result := NewCursorResult(items, "next", "prev", true)

	// Assert
	if len(result.Items()) != 3 {
		t.Errorf("Items() length = %v, want 3", len(result.Items()))
	}
	if result.NextCursor() != "next" {
		t.Errorf("NextCursor() = %v, want next", result.NextCursor())
	}
	if result.PrevCursor() != "prev" {
		t.Errorf("PrevCursor() = %v, want prev", result.PrevCursor())
	}
	if !result.HasMore() {
		t.Error("HasMore() should be true")
	}
}

func TestCursorResult_IsEmpty(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := NewCursorResult([]int{}, "", "", false)
		if !result.IsEmpty() {
			t.Error("IsEmpty() should return true")
		}
	})

	t.Run("non-empty", func(t *testing.T) {
		result := NewCursorResult([]int{1}, "next", "", true)
		if result.IsEmpty() {
			t.Error("IsEmpty() should return false")
		}
	})
}

func TestEmptyCursorResult(t *testing.T) {
	result := EmptyCursorResult[int]()

	if !result.IsEmpty() {
		t.Error("should be empty")
	}
	if result.HasMore() {
		t.Error("HasMore() should be false")
	}
	if result.NextCursor() != "" {
		t.Error("NextCursor() should be empty")
	}
}

// ============================================================================
// Cursor Encoding Tests
// ============================================================================

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name   string
		values []string
	}{
		{
			name:   "empty values",
			values: []string{},
		},
		{
			name:   "single value",
			values: []string{"abc123"},
		},
		{
			name:   "multiple values",
			values: []string{"2024-01-01", "order-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeCursor(tt.values...)

			if len(tt.values) == 0 {
				if encoded != "" {
					t.Errorf("EncodeCursor() = %v, want empty", encoded)
				}
				return
			}

			// Verify by decoding
			decoded, err := DecodeCursor(encoded)
			if err != nil {
				t.Errorf("DecodeCursor() error = %v", err)
				return
			}
			if len(decoded) != len(tt.values) {
				t.Errorf("round-trip length = %v, want %v", len(decoded), len(tt.values))
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	t.Run("empty cursor", func(t *testing.T) {
		got, err := DecodeCursor("")
		if err != nil {
			t.Errorf("error = %v, want nil", err)
		}
		if got != nil {
			t.Errorf("DecodeCursor() = %v, want nil", got)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := DecodeCursor("not-valid-base64!!!")
		if err != ErrInvalidCursor {
			t.Errorf("error = %v, want %v", err, ErrInvalidCursor)
		}
	})

	t.Run("single value round-trip", func(t *testing.T) {
		original := []string{"abc123"}
		encoded := EncodeCursor(original...)
		decoded, err := DecodeCursor(encoded)

		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(decoded) != 1 || decoded[0] != "abc123" {
			t.Errorf("DecodeCursor() = %v, want %v", decoded, original)
		}
	})

	t.Run("multiple values round-trip", func(t *testing.T) {
		original := []string{"2024-01-01", "order-123"}
		encoded := EncodeCursor(original...)
		decoded, err := DecodeCursor(encoded)

		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(decoded) != 2 {
			t.Fatalf("length = %v, want 2", len(decoded))
		}
		if decoded[0] != "2024-01-01" || decoded[1] != "order-123" {
			t.Errorf("DecodeCursor() = %v, want %v", decoded, original)
		}
	})
}

func TestDecodeCursorSingle(t *testing.T) {
	t.Run("valid single value", func(t *testing.T) {
		encoded := EncodeCursor("abc123")
		got, err := DecodeCursorSingle(encoded)

		if err != nil {
			t.Errorf("error = %v, want nil", err)
		}
		if got != "abc123" {
			t.Errorf("DecodeCursorSingle() = %v, want abc123", got)
		}
	})

	t.Run("multiple values returns error", func(t *testing.T) {
		encoded := EncodeCursor("value1", "value2")
		_, err := DecodeCursorSingle(encoded)

		if err != ErrInvalidCursor {
			t.Errorf("error = %v, want %v", err, ErrInvalidCursor)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := DecodeCursorSingle("invalid!!!")
		if err != ErrInvalidCursor {
			t.Errorf("error = %v, want %v", err, ErrInvalidCursor)
		}
	})
}

func TestCursorRoundTrip(t *testing.T) {
	// Test that encode -> decode returns original values
	tests := []struct {
		name   string
		values []string
	}{
		{"single value", []string{"test-id-123"}},
		{"multiple values", []string{"2024-01-15T10:30:00Z", "uuid-456", "extra"}},
		{"special characters", []string{"hello world", "foo/bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded := EncodeCursor(tt.values...)

			// Decode
			decoded, err := DecodeCursor(encoded)
			if err != nil {
				t.Fatalf("DecodeCursor() error = %v", err)
			}

			// Verify round-trip
			if len(decoded) != len(tt.values) {
				t.Fatalf("round-trip length = %v, want %v", len(decoded), len(tt.values))
			}
			for i := range decoded {
				if decoded[i] != tt.values[i] {
					t.Errorf("round-trip[%d] = %v, want %v", i, decoded[i], tt.values[i])
				}
			}
		})
	}
}
