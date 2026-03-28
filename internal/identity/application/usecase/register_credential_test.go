package usecase

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/internal/shared/domain/event"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RegisterInput
		mock    func(repo *credential.MockRepository, idGen *port.MockIDGenerator, bus *event.MockEventBus)
		wantID  string
		wantErr bool
	}{
		{
			name:  "success",
			input: dto.RegisterInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, idGen *port.MockIDGenerator, bus *event.MockEventBus) {
				email, _ := credential.NewEmail("alice@example.com")
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(nil, credential.ErrNotFound)
				idGen.EXPECT().Generate().Return("c-1")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				bus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantID: "c-1",
		},
		{
			name:  "success even if event publish fails",
			input: dto.RegisterInput{Email: "bob@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, idGen *port.MockIDGenerator, bus *event.MockEventBus) {
				email, _ := credential.NewEmail("bob@example.com")
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(nil, credential.ErrNotFound)
				idGen.EXPECT().Generate().Return("c-2")
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				bus.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			wantID: "c-2",
		},
		{
			name:    "invalid email",
			input:   dto.RegisterInput{Email: "bad", Password: "secret123"},
			mock:    func(_ *credential.MockRepository, _ *port.MockIDGenerator, _ *event.MockEventBus) {},
			wantErr: true,
		},
		{
			name:    "password too short",
			input:   dto.RegisterInput{Email: "alice@example.com", Password: "ab"},
			mock:    func(_ *credential.MockRepository, _ *port.MockIDGenerator, _ *event.MockEventBus) {},
			wantErr: true,
		},
		{
			name:  "duplicate email",
			input: dto.RegisterInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, _ *port.MockIDGenerator, _ *event.MockEventBus) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				existing, _ := credential.NewCredential(credential.NewCredentialParams{
					ID: "c-existing", Email: email, Password: pwd,
				})
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(existing, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := credential.NewMockRepository(ctrl)
			idGen := port.NewMockIDGenerator(ctrl)
			bus := event.NewMockEventBus(ctrl)
			tt.mock(repo, idGen, bus)

			uc := NewRegisterUseCase(repo, idGen, bus)
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

			if got.Status != string(credential.StatusActive) {
				t.Errorf("Status = %q, want %q", got.Status, credential.StatusActive)
			}
		})
	}
}
