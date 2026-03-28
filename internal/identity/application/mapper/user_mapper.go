package mapper

import (
	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
)

// ToUserOutput converts a domain User to a UserOutput DTO.
func ToUserOutput(u *user.User) dto.UserOutput {
	return dto.UserOutput{
		ID:        u.ID(),
		Email:     u.Email().Address(),
		Name:      u.Name(),
		Status:    string(u.Status()),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}
