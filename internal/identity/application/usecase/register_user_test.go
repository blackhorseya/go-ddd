package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
)

func TestRegisterUserUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RegisterUserInput
		mock    func(repo *user.MockRepository, idGen *port.MockIDGenerator)
		wantID  string
		wantErr bool
	}{
		{
			name:  "success",
			input: dto.RegisterUserInput{Email: "alice@example.com", Password: "secret123", Name: "Alice"},
			mock: func(repo *user.MockRepository, idGen *port.MockIDGenerator) {
				idGen.EXPECT().Generate().Return("u-1")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantID: "u-1",
		},
		{
			name:    "invalid email",
			input:   dto.RegisterUserInput{Email: "bad", Password: "secret123", Name: "Alice"},
			mock:    func(repo *user.MockRepository, idGen *port.MockIDGenerator) {},
			wantErr: true,
		},
		{
			name:    "password too short",
			input:   dto.RegisterUserInput{Email: "alice@example.com", Password: "ab", Name: "Alice"},
			mock:    func(repo *user.MockRepository, idGen *port.MockIDGenerator) {},
			wantErr: true,
		},
		{
			name:  "empty name",
			input: dto.RegisterUserInput{Email: "alice@example.com", Password: "secret123", Name: ""},
			mock: func(repo *user.MockRepository, idGen *port.MockIDGenerator) {
				idGen.EXPECT().Generate().Return("u-x")
			},
			wantErr: true,
		},
		{
			name:  "duplicate email",
			input: dto.RegisterUserInput{Email: "alice@example.com", Password: "secret123", Name: "Alice"},
			mock: func(repo *user.MockRepository, idGen *port.MockIDGenerator) {
				idGen.EXPECT().Generate().Return("u-2")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(user.ErrEmailDuplicated)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := user.NewMockRepository(ctrl)
			idGen := port.NewMockIDGenerator(ctrl)
			tt.mock(repo, idGen)

			uc := NewRegisterUserUseCase(repo, idGen)
			got, err := uc.Execute(context.Background(), tt.input)

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

			if got.Email != tt.input.Email {
				t.Errorf("Email = %q, want %q", got.Email, tt.input.Email)
			}

			if got.Status != string(user.StatusInactive) {
				t.Errorf("Status = %q, want %q", got.Status, user.StatusInactive)
			}
		})
	}
}
