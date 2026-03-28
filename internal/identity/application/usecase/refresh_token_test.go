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

func TestRefreshTokenUseCase_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   dto.RefreshInput
		mock    func(repo *credential.MockRepository, tokenSvc *port.MockTokenService)
		wantErr error
	}{
		{
			name:  "success",
			input: dto.RefreshInput{RefreshToken: "valid-rt"},
			mock: func(repo *credential.MockRepository, tokenSvc *port.MockTokenService) {
				tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "valid-rt").
					Return(port.TokenClaims{CredentialID: "c-1", Email: "alice@example.com"}, nil)

				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.NewCredential(credential.NewCredentialParams{
					ID: "c-1", Email: email, Password: pwd,
				})
				repo.EXPECT().FindByID(gomock.Any(), "c-1").Return(cred, nil)

				tokenSvc.EXPECT().GenerateTokenPair(gomock.Any(), port.TokenClaims{
					CredentialID: "c-1", Email: "alice@example.com",
				}).Return(port.TokenPair{
					AccessToken: "new-at", RefreshToken: "new-rt", ExpiresIn: 900,
				}, nil)
			},
		},
		{
			name:  "invalid refresh token",
			input: dto.RefreshInput{RefreshToken: "bad-token"},
			mock: func(_ *credential.MockRepository, tokenSvc *port.MockTokenService) {
				tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "bad-token").
					Return(port.TokenClaims{}, errors.New("invalid"))
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "credential not found",
			input: dto.RefreshInput{RefreshToken: "valid-rt"},
			mock: func(repo *credential.MockRepository, tokenSvc *port.MockTokenService) {
				tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "valid-rt").
					Return(port.TokenClaims{CredentialID: "c-1"}, nil)
				repo.EXPECT().FindByID(gomock.Any(), "c-1").Return(nil, credential.ErrNotFound)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:  "suspended credential",
			input: dto.RefreshInput{RefreshToken: "valid-rt"},
			mock: func(repo *credential.MockRepository, tokenSvc *port.MockTokenService) {
				tokenSvc.EXPECT().ValidateRefreshToken(gomock.Any(), "valid-rt").
					Return(port.TokenClaims{CredentialID: "c-1"}, nil)

				email, _ := credential.NewEmail("alice@example.com")
				pwd, _ := credential.NewHashedPassword("secret123")
				cred, _ := credential.ReconstituteCredential(credential.ReconstituteCredentialParams{
					ID: "c-1", Email: email, Password: pwd, Status: credential.StatusSuspended,
				})
				repo.EXPECT().FindByID(gomock.Any(), "c-1").Return(cred, nil)
			},
			wantErr: credential.ErrCredentialSuspended,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := credential.NewMockRepository(ctrl)
			tokenSvc := port.NewMockTokenService(ctrl)
			tt.mock(repo, tokenSvc)

			uc := NewRefreshTokenUseCase(repo, tokenSvc)
			got, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Execute() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() unexpected error: %v", err)
			}

			if got.AccessToken == "" {
				t.Error("AccessToken should not be empty")
			}
		})
	}
}
