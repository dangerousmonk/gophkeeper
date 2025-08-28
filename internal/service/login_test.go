package service

import (
	"context"
	"errors"
	"testing"

	encryptm "github.com/dangerousmonk/gophkeeper/internal/encryption/mocks"
	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/postgres"
	pgm "github.com/dangerousmonk/gophkeeper/internal/postgres/mocks"
	"github.com/dangerousmonk/gophkeeper/internal/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	testUser := models.User{ID: 1, Login: "guest", PasswordHash: "$2a$10$JlBjqiVSWraOUZ8SkHwnmO38Vfscr3bloe8eDlObLBFwRImhJjsbq", Active: true}
	jwtAuthenticator, err := utils.NewJWTAuthenticator("secretkeysecretkeysecretkeykeyke")
	require.NoError(t, err)

	repoError := errors.New("driver: bad connection")

	cases := []struct {
		name             string
		buildRepoStub    func(s *pgm.MockUserRepository)
		buildEncryptStub func(es *encryptm.MockPasswordEncryptor)
		login            string
		password         string
		expectedError    error
		wantError        bool
	}{
		{
			name:     "success",
			login:    testUser.Login,
			password: "guest",
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(1).
					Return(nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name:     "repository_error",
			login:    testUser.Login,
			password: "guest",
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(models.User{}, repoError)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
			},
			wantError:     true,
			expectedError: repoError,
		},
		{
			name:     "not_found",
			login:    testUser.Login,
			password: "guest",
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(models.User{}, postgres.ErrUserNotFound)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
			},
			wantError:     true,
			expectedError: ErrNoUserWithLogin,
		},
		{
			name:     "bad_credentials",
			login:    testUser.Login,
			password: "foobar",
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("foobar", testUser.PasswordHash).Times(1).
					Return(ErrInvalidCredentials)
			},
			wantError:     true,
			expectedError: ErrInvalidCredentials,
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
			token, err := s.Login(context.Background(), tc.login, tc.password, jwtAuthenticator)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)
				require.GreaterOrEqual(t, len(token), 4)
			}

		})
	}
}
