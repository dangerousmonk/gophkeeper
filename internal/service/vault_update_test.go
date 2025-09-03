package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dangerousmonk/gophkeeper/internal/models"
	"github.com/dangerousmonk/gophkeeper/internal/postgres/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestVaultUpdate(t *testing.T) {
	testUserID := 1
	testVaultID := 123
	newEncryptedData := []byte("\x246fc350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe")
	repoError := errors.New("driver: bad connection")

	cases := []struct {
		name          string
		userID        int
		vaultID       int
		vaultName     string
		encryptedData []byte
		buildRepoStub func(s *mocks.MockVaultRepository)
		expectedError error
		wantError     bool
	}{
		{
			name:          "success",
			userID:        testUserID,
			vaultID:       testVaultID,
			vaultName:     "updated card",
			encryptedData: newEncryptedData,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: testUserID, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Update(gomock.Any(), testVaultID, "updated card", newEncryptedData).Times(1).
					Return(nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name:          "repository_error",
			userID:        testUserID,
			vaultID:       testVaultID,
			vaultName:     "updated card",
			encryptedData: newEncryptedData,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{}, repoError)
				r.EXPECT().
					Update(gomock.Any(), testVaultID, "updated card", newEncryptedData).Times(0)
			},
			wantError:     true,
			expectedError: repoError,
		},
		{
			name:          "owner_missmatch",
			userID:        testUserID,
			vaultID:       testVaultID,
			vaultName:     "updated card",
			encryptedData: newEncryptedData,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: 99, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Update(gomock.Any(), testVaultID, "updated card", newEncryptedData).Times(0)
			},
			wantError:     true,
			expectedError: ErrVaultOwnerMismatchUpdate,
		},
		{
			name:          "update_error",
			userID:        testUserID,
			vaultID:       testVaultID,
			vaultName:     "updated card",
			encryptedData: newEncryptedData,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: testUserID, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Update(gomock.Any(), testVaultID, "updated card", newEncryptedData).Times(1).
					Return(repoError)
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

			repo := mocks.NewMockVaultRepository(ctrl)
			tc.buildRepoStub(repo)

			s := NewVaultService(repo)
			err := s.Update(context.Background(), tc.vaultID, tc.userID, tc.vaultName, tc.encryptedData)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
