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

func TestVaultDeactivate(t *testing.T) {
	testUserID := 1
	testVaultID := 123
	repoError := errors.New("driver: bad connection")

	cases := []struct {
		name          string
		userID        int
		vaultID       int
		buildRepoStub func(s *mocks.MockVaultRepository)
		expectedError error
		wantError     bool
	}{
		{
			name:    "success",
			userID:  testUserID,
			vaultID: testVaultID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: testUserID, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Deactivate(gomock.Any(), testVaultID).Times(1).
					Return(nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name:    "repository_lookup_error",
			userID:  testUserID,
			vaultID: testVaultID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{}, repoError)
				r.EXPECT().
					Deactivate(gomock.Any(), testVaultID).Times(0)
			},
			wantError:     true,
			expectedError: repoError,
		},
		{
			name:    "owner_missmatch",
			userID:  testUserID,
			vaultID: testVaultID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: 99, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Deactivate(gomock.Any(), testVaultID).Times(0)
			},
			wantError:     true,
			expectedError: ErrVaultOwnerMismatch,
		},
		{
			name:    "repository_deactivate_error",
			userID:  testUserID,
			vaultID: testVaultID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Get(gomock.Any(), testVaultID).Times(1).
					Return(models.Vault{ID: testVaultID, UserID: testUserID, Name: "test", DataType: models.Credentials}, nil)
				r.EXPECT().
					Deactivate(gomock.Any(), testVaultID).Times(1).
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
			err := s.Deactivate(context.Background(), tc.userID, tc.vaultID)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}

		})
	}
}
