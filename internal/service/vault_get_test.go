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

func TestVaultGet(t *testing.T) {
	testUserID := 123
	repoError := errors.New("driver: bad connection")
	testVaults := []models.Vault{
		{ID: 1, UserID: testUserID, Name: "test1", DataType: models.Credentials, EncryptedData: []byte("\x246fc350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe")},
		{ID: 2, UserID: testUserID, Name: "test2", DataType: models.Binary, EncryptedData: []byte("\x24c350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe")},
	}

	cases := []struct {
		name          string
		userID        int
		buildRepoStub func(s *mocks.MockVaultRepository)
		expectedError error
		wantError     bool
	}{
		{
			name:   "success",
			userID: testUserID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					GetByUserID(gomock.Any(), testUserID).Times(1).
					Return(testVaults, nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name:   "repository_error",
			userID: testUserID,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					GetByUserID(gomock.Any(), testUserID).Times(1).
					Return([]models.Vault{}, repoError)
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
			res, err := s.GetByUser(context.Background(), testUserID)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, testVaults, res)
			}

		})
	}
}
