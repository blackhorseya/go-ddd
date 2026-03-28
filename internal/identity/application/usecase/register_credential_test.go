package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RegisterInput
		mock    func(repo *credential.MockRepository, idGen *port.MockIDGenerator)
		wantID  string
		wantErr bool
	}{
		{
			name:  "success",
			input: dto.RegisterInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, idGen *port.MockIDGenerator) {
				idGen.EXPECT().Generate().Return("c-1")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantID: "c-1",
		},
		{
			name:    "invalid email",
			input:   dto.RegisterInput{Email: "bad", Password: "secret123"},
			mock:    func(repo *credential.MockRepository, idGen *port.MockIDGenerator) {},
			wantErr: true,
		},
		{
			name:    "password too short",
			input:   dto.RegisterInput{Email: "alice@example.com", Password: "ab"},
			mock:    func(repo *credential.MockRepository, idGen *port.MockIDGenerator) {},
			wantErr: true,
		},
		{
			name:  "duplicate email",
			input: dto.RegisterInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, idGen *port.MockIDGenerator) {
				idGen.EXPECT().Generate().Return("c-2")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(credential.ErrEmailDuplicated)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := credential.NewMockRepository(ctrl)
			idGen := port.NewMockIDGenerator(ctrl)
			tt.mock(repo, idGen)

			uc := NewRegisterUseCase(repo, idGen)
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

			if got.Status != string(credential.StatusInactive) {
				t.Errorf("Status = %q, want %q", got.Status, credential.StatusInactive)
			}
		})
	}
}
