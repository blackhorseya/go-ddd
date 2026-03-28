package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
)

func TestGetUserUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(repo *user.MockRepository)
		wantID  string
		wantErr bool
	}{
		{
			name: "success",
			id:   "u-1",
			mock: func(repo *user.MockRepository) {
				email, _ := user.NewEmail("alice@example.com")
				pwd, _ := user.NewHashedPassword("secret123")
				u, _ := user.NewUser(user.NewUserParams{
					ID:       "u-1",
					Email:    email,
					Password: pwd,
					Name:     "Alice",
				})
				repo.EXPECT().FindByID(gomock.Any(), "u-1").Return(u, nil)
			},
			wantID: "u-1",
		},
		{
			name: "not found",
			id:   "non-existent",
			mock: func(repo *user.MockRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "non-existent").Return(nil, user.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := user.NewMockRepository(ctrl)
			tt.mock(repo)

			uc := NewGetUserUseCase(repo)
			got, err := uc.Execute(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() unexpected error: %v", err)
			}

			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}
