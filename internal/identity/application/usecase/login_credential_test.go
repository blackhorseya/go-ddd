package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/blackhorseya/go-ddd/internal/identity/application/dto"
	"github.com/blackhorseya/go-ddd/internal/identity/application/port"
	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
)

func TestLoginUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.LoginInput
		mock    func(repo *credential.MockRepository, tokenSvc *port.MockTokenService)
		wantErr error
	}{
		{
			name:  "success",
			input: dto.LoginInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, tokenSvc *port.MockTokenService) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.NewCredential(credential.NewCredentialParams{
					ID: "c-1", Email: email, Password: pwd,
				})
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)
				tokenSvc.EXPECT().GenerateTokenPair(gomock.Any(), port.TokenClaims{
					CredentialID: "c-1", Email: "alice@example.com",
				}).Return(port.TokenPair{
					AccessToken: "at", RefreshToken: "rt", ExpiresIn: 900,
				}, nil)
			},
		},
		{
			name:    "invalid email format",
			input:   dto.LoginInput{Email: "bad", Password: "secret123"},
			mock:    func(_ *credential.MockRepository, _ *port.MockTokenService) {},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "email not found",
			input: dto.LoginInput{Email: "notfound@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, _ *port.MockTokenService) {
				email, _ := credential.NewEmail("notfound@example.com")
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(nil, credential.ErrNotFound)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "wrong password",
			input: dto.LoginInput{Email: "alice@example.com", Password: "wrongpass"},
			mock: func(repo *credential.MockRepository, _ *port.MockTokenService) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.NewCredential(credential.NewCredentialParams{
					ID: "c-1", Email: email, Password: pwd,
				})
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "suspended credential",
			input: dto.LoginInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, _ *port.MockTokenService) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.ReconstituteCredential(credential.ReconstituteCredentialParams{
					ID: "c-1", Email: email, Password: pwd, Status: credential.StatusSuspended,
				})
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)
			},
			wantErr: credential.ErrCredentialSuspended,
		},
		{
			name:  "token generation failure",
			input: dto.LoginInput{Email: "alice@example.com", Password: "secret123"},
			mock: func(repo *credential.MockRepository, tokenSvc *port.MockTokenService) {
				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.NewCredential(credential.NewCredentialParams{
					ID: "c-1", Email: email, Password: pwd,
				})
				repo.EXPECT().FindByEmail(gomock.Any(), email).Return(cred, nil)
				tokenSvc.EXPECT().GenerateTokenPair(gomock.Any(), gomock.Any()).
					Return(port.TokenPair{}, errors.New("signing error"))
			},
			wantErr: errors.New("signing error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := credential.NewMockRepository(ctrl)
			tokenSvc := port.NewMockTokenService(ctrl)
			tt.mock(repo, tokenSvc)

			uc := NewLoginUseCase(repo, tokenSvc)
			got, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("Execute() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() unexpected error: %v", err)
			}

			if got.Credential.ID != "c-1" {
				t.Errorf("Credential.ID = %q, want %q", got.Credential.ID, "c-1")
			}

			if got.AccessToken == "" {
				t.Error("AccessToken should not be empty")
			}

			if got.RefreshToken == "" {
				t.Error("RefreshToken should not be empty")
			}
		})
	}
}
