package domain

import (
	"encoding/base64"
	"errors"
	"strings"
)

// Pagination errors
var (
	ErrInvalidPage     = errors.New("page number must be greater than 0")
	ErrInvalidPageSize = errors.New("page size must be between 1 and max page size")
	ErrInvalidCursor   = errors.New("invalid cursor format")
)

// Default pagination constants
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 1000
)

// SortDirection represents the sort order direction
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// SortOption represents a sorting configuration
type SortOption struct {
	field     string
	direction SortDirection
}

// NewSortOption creates a new sort option with validation
func NewSortOption(field string, direction SortDirection) SortOption {
	if direction != SortAsc && direction != SortDesc {
		direction = SortAsc
	}
	return SortOption{
		field:     field,
		direction: direction,
	}
}

func (x SortOption) Field() string            { return x.field }
func (x SortOption) Direction() SortDirection { return x.direction }
func (x SortOption) IsAscending() bool        { return x.direction == SortAsc }

// ============================================================================
// Offset-based Pagination (傳統頁碼分頁)
// ============================================================================

// PageRequest represents an offset-based pagination request (Value Object)
type PageRequest struct {
	page     int
	pageSize int
	sort     []SortOption
}

// NewPageRequest creates a validated page request
func NewPageRequest(page, pageSize int) (PageRequest, error) {
	if page < 1 {
		return PageRequest{}, ErrInvalidPage
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		return PageRequest{}, ErrInvalidPageSize
	}
	return PageRequest{
		page:     page,
		pageSize: pageSize,
	}, nil
}

// NewPageRequestWithDefaults creates a page request with default values
func NewPageRequestWithDefaults() PageRequest {
	return PageRequest{
		page:     DefaultPage,
		pageSize: DefaultPageSize,
	}
}

// WithSort returns a new PageRequest with sort options
func (x PageRequest) WithSort(sort ...SortOption) PageRequest {
	return PageRequest{
		page:     x.page,
		pageSize: x.pageSize,
		sort:     sort,
	}
}

// Getters
func (x PageRequest) Page() int          { return x.page }
func (x PageRequest) PageSize() int      { return x.pageSize }
func (x PageRequest) Sort() []SortOption { return x.sort }
func (x PageRequest) Offset() int        { return (x.page - 1) * x.pageSize }
func (x PageRequest) Limit() int         { return x.pageSize }

// PageResult represents a paginated result with metadata
type PageResult[T any] struct {
	items      []T
	page       int
	pageSize   int
	totalItems int64
	totalPages int
}

// NewPageResult creates a new page result
func NewPageResult[T any](items []T, page, pageSize int, totalItems int64) PageResult[T] {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize > 0 {
		totalPages++
	}
	return PageResult[T]{
		items:      items,
		page:       page,
		pageSize:   pageSize,
		totalItems: totalItems,
		totalPages: totalPages,
	}
}

// Getters
func (x PageResult[T]) Items() []T        { return x.items }
func (x PageResult[T]) Page() int         { return x.page }
func (x PageResult[T]) PageSize() int     { return x.pageSize }
func (x PageResult[T]) TotalItems() int64 { return x.totalItems }
func (x PageResult[T]) TotalPages() int   { return x.totalPages }
func (x PageResult[T]) HasNext() bool     { return x.page < x.totalPages }
func (x PageResult[T]) HasPrev() bool     { return x.page > 1 }
func (x PageResult[T]) IsEmpty() bool     { return len(x.items) == 0 }

// EmptyPageResult creates an empty page result
func EmptyPageResult[T any]() PageResult[T] {
	return PageResult[T]{
		items:      []T{},
		page:       1,
		pageSize:   DefaultPageSize,
		totalItems: 0,
		totalPages: 0,
	}
}

// ============================================================================
// Cursor-based Pagination (游標分頁，適合大資料集)
// ============================================================================

// CursorRequest represents a cursor-based pagination request
type CursorRequest struct {
	cursor   string
	pageSize int
	sort     []SortOption
}

// NewCursorRequest creates a validated cursor request
func NewCursorRequest(cursor string, pageSize int) (CursorRequest, error) {
	if pageSize < 1 || pageSize > MaxPageSize {
		return CursorRequest{}, ErrInvalidPageSize
	}
	return CursorRequest{
		cursor:   cursor,
		pageSize: pageSize,
	}, nil
}

// NewCursorRequestWithDefaults creates a cursor request with defaults
func NewCursorRequestWithDefaults() CursorRequest {
	return CursorRequest{
		cursor:   "",
		pageSize: DefaultPageSize,
	}
}

// WithSort returns a new CursorRequest with sort options
func (x CursorRequest) WithSort(sort ...SortOption) CursorRequest {
	return CursorRequest{
		cursor:   x.cursor,
		pageSize: x.pageSize,
		sort:     sort,
	}
}

// Getters
func (x CursorRequest) Cursor() string     { return x.cursor }
func (x CursorRequest) PageSize() int      { return x.pageSize }
func (x CursorRequest) Sort() []SortOption { return x.sort }
func (x CursorRequest) Limit() int         { return x.pageSize }
func (x CursorRequest) HasCursor() bool    { return x.cursor != "" }

// CursorResult represents a cursor-based paginated result
type CursorResult[T any] struct {
	items      []T
	nextCursor string
	prevCursor string
	hasMore    bool
}

// NewCursorResult creates a new cursor result
func NewCursorResult[T any](items []T, nextCursor, prevCursor string, hasMore bool) CursorResult[T] {
	return CursorResult[T]{
		items:      items,
		nextCursor: nextCursor,
		prevCursor: prevCursor,
		hasMore:    hasMore,
	}
}

// Getters
func (x CursorResult[T]) Items() []T         { return x.items }
func (x CursorResult[T]) NextCursor() string { return x.nextCursor }
func (x CursorResult[T]) PrevCursor() string { return x.prevCursor }
func (x CursorResult[T]) HasMore() bool      { return x.hasMore }
func (x CursorResult[T]) IsEmpty() bool      { return len(x.items) == 0 }

// EmptyCursorResult creates an empty cursor result
func EmptyCursorResult[T any]() CursorResult[T] {
	return CursorResult[T]{
		items:      []T{},
		nextCursor: "",
		prevCursor: "",
		hasMore:    false,
	}
}

// ============================================================================
// Cursor Encoding (Base64)
// ============================================================================

// cursorSeparator is used to join multiple values in a cursor.
// Using null byte as it won't appear in normal string values.
const cursorSeparator = "\x00"

// EncodeCursor encodes values into a base64 cursor string.
// Supports single value or multiple values.
// Example: EncodeCursor("2024-01-01T10:30:00Z", "abc123") -> base64 encoded string
func EncodeCursor(values ...string) string {
	if len(values) == 0 {
		return ""
	}
	joined := strings.Join(values, cursorSeparator)
	return base64.URLEncoding.EncodeToString([]byte(joined))
}

// DecodeCursor decodes a base64 cursor string back to its values.
// Returns the original values.
// Example: DecodeCursor(encoded) -> ["2024-01-01T10:30:00Z", "abc123"], nil
func DecodeCursor(cursor string) ([]string, error) {
	if cursor == "" {
		return nil, nil
	}
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, ErrInvalidCursor
	}
	return strings.Split(string(decoded), cursorSeparator), nil
}

// DecodeCursorSingle decodes a cursor expecting exactly one value.
func DecodeCursorSingle(cursor string) (string, error) {
	values, err := DecodeCursor(cursor)
	if err != nil {
		return "", err
	}
	if len(values) != 1 {
		return "", ErrInvalidCursor
	}
	return values[0], nil
}
