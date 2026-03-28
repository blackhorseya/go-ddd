package idgen

import (
	"github.com/google/uuid"

	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
)

var _ port.IDGenerator = (*uuidGen)(nil)

type uuidGen struct{}

// NewUUIDGenerator creates an IDGenerator backed by UUID v4.
func NewUUIDGenerator() port.IDGenerator {
	return &uuidGen{}
}

func (i *uuidGen) Generate() string {
	return uuid.New().String()
}
