package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
)

func TestGetCredentialUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(repo *credential.MockRepository)
		wantID  string
		wantErr bool
	}{
		{
			name: "success",
			id:   "c-1",
			mock: func(repo *credential.MockRepository) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				c, _ := credential.NewCredential(credential.NewCredentialParams{
					ID:       "c-1",
					Email:    email,
					Password: pwd,
				})
				repo.EXPECT().FindByID(gomock.Any(), "c-1").Return(c, nil)
			},
			wantID: "c-1",
		},
		{
			name: "not found",
			id:   "non-existent",
			mock: func(repo *credential.MockRepository) {
				repo.EXPECT().FindByID(gomock.Any(), "non-existent").Return(nil, credential.ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := credential.NewMockRepository(ctrl)
			tt.mock(repo)

			uc := NewGetCredentialUseCase(repo)
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
