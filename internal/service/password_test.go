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

func TestPassword(t *testing.T) {
	testUser := models.User{ID: 1, Login: "guest", PasswordHash: "$2a$10$JlBjqiVSWraOUZ8SkHwnmO38Vfscr3bloe8eDlObLBFwRImhJjsbq", Active: true}
	repoError := errors.New("driver: bad connection")
	newPass := "guest-new"
	newHash := "$2a$10$Jy14COgXoPSo1LbQJvNP0uXHZbpy0aEPwAlRnuU8oVTujwGfjnupW"

	cases := []struct {
		name             string
		req              *models.ChangePasswordRequest
		buildRepoStub    func(s *pgm.MockUserRepository)
		buildEncryptStub func(es *encryptm.MockPasswordEncryptor)
		expectedError    error
		wantError        bool
	}{
		{
			name: "success",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: newPass},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(1).
					Return(nil)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(1).
					Return(nil)
				r.EXPECT().
					HashPassword(newPass).Times(1).
					Return(newHash, nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name: "password_not_changed",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: "guest"},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(0)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
				r.EXPECT().
					HashPassword(newPass).Times(0)
			},
			wantError:     true,
			expectedError: ErrPasswordNotChanged,
		},
		{
			name: "repository_error",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: newPass},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(models.User{}, repoError)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
				r.EXPECT().
					HashPassword(newPass).Times(0)
			},
			wantError:     true,
			expectedError: repoError,
		},
		{
			name: "password_validation",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: ""},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(0)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
				r.EXPECT().
					HashPassword(newPass).Times(0)
			},
			wantError:     true,
			expectedError: nil,
		},
		{
			name: "password_validation",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: "1"},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(0)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(0)
				r.EXPECT().
					HashPassword(newPass).Times(0)
			},
			wantError:     true,
			expectedError: nil,
		},
		{
			name: "check_password_failed",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: newPass},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(1).
					Return(errors.New("decrypt:some error"))
				r.EXPECT().
					HashPassword(newPass).Times(0)
			},
			wantError:     true,
			expectedError: ErrPasswordDecryptionFailed,
		},
		{
			name: "encrypt_password_failed",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: newPass},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(0)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(1).
					Return(nil)
				r.EXPECT().
					HashPassword(newPass).Times(1).
					Return("", errors.New("encrypt:some error"))
			},
			wantError:     true,
			expectedError: ErrPasswordEncryptionFailed,
		},
		{
			name: "update_repo_failed",
			req:  &models.ChangePasswordRequest{Login: testUser.Login, CurrentPassword: "guest", NewPassword: newPass},
			buildRepoStub: func(r *pgm.MockUserRepository) {
				r.EXPECT().
					Get(gomock.Any(), testUser.Login).Times(1).
					Return(testUser, nil)
				r.EXPECT().
					UpdatePassword(gomock.Any(), testUser.ID, newHash).Times(1).
					Return(repoError)
			},
			buildEncryptStub: func(r *encryptm.MockPasswordEncryptor) {
				r.EXPECT().
					CheckPassword("guest", testUser.PasswordHash).Times(1).
					Return(nil)
				r.EXPECT().
					HashPassword(newPass).Times(1).
					Return(newHash, nil)
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
			resp, err := s.ChangePassword(context.Background(), testUser.ID, tc.req)

			if tc.wantError {
				require.Error(t, err)
				if tc.name != "password_validation" {
					require.ErrorIs(t, err, tc.expectedError)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, resp.Success, true)
			}

		})
	}
}
