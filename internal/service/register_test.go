package service

import (
	"context"
	"errors"
	"testing"

	encryptm "github.com/dangerousmonk/gophkeeper/internal/encryption/mocks"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	pgm "github.com/dangerousmonk/gophkeeper/internal/postgres/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	hash := "$2a$10$Jy14COgXoPSo1LbQJvNP0uXHZbpy0aEPwAlRnuU8oVTujwGfjnupW"
	testUser := models.User{ID: 1, Login: "guest", PasswordHash: hash, Active: true}
	repoError := errors.New("driver: bad connection")
	password := "guest"
	request := &models.RegisterUserRequest{Login: testUser.Login, Password: password, HashedPassword: hash}

	cases := []struct {
		name             string
		req              *models.RegisterUserRequest
		buildRepoStub    func(s *pgm.MockUserRepository)
		buildEncryptStub func(es *encryptm.MockPasswordEncryptor)
		expectedError    error
		wantError        bool
	}{
		{
			name: "success",
			req:  &models.RegisterUserRequest{Login: testUser.Login, Password: password},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Create(gomock.Any(), request).Times(1).
					Return(testUser.ID, nil)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					HashPassword(password).Times(1).
					Return(hash, nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name: "password_validation",
			req:  &models.RegisterUserRequest{Login: testUser.Login, Password: ""},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Create(gomock.Any(), request).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					HashPassword(password).Times(0)
			},
			wantError:     true,
			expectedError: nil,
		},
		{
			name: "password_validation",
			req:  &models.RegisterUserRequest{Login: testUser.Login, Password: "123a"},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Create(gomock.Any(), request).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					HashPassword(password).Times(0)
			},
			wantError:     true,
			expectedError: nil,
		},
		{
			name: "encrypt_password_error",
			req:  &models.RegisterUserRequest{Login: testUser.Login, Password: password},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Create(gomock.Any(), request).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					HashPassword(password).Times(1).
					Return("", errors.New("encrypt:some error"))
			},
			wantError:     true,
			expectedError: ErrPasswordEncryptionFailed,
		},
		{
			name: "repository_error",
			req:  &models.RegisterUserRequest{Login: testUser.Login, Password: password},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Create(gomock.Any(), request).Times(1).
					Return(-1, repoError)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					HashPassword(password).Times(1).
					Return(hash, nil)
			},
			wantError:     true,
			expectedError: repoError,
		},
	}

	for i := range cases {
		tc := cases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := pgm.NewMockUserRepository(ctrl)
			encryptor := encryptm.NewMockPasswordEncryptor(ctrl)
			tc.buildRepoStub(repo)
			tc.buildEncryptStub(encryptor)

			s := NewUserService(repo, encryptor)
			resp, err := s.Register(context.Background(), tc.req)

			if tc.wantError {
				require.Error(t, err)
				if tc.name != "password_validation" {
					require.ErrorIs(t, err, tc.expectedError)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, resp.Success, true)
				require.Equal(t, resp.Login, tc.req.Login)
			}

		})
	}
}
