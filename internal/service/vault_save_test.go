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

func TestVaultSave(t *testing.T) {
	repoError := errors.New("driver: bad connection")
	testVault := &models.Vault{
		UserID:        1,
		Name:          "test content",
		DataType:      models.Credentials,
		EncryptedData: []byte("\x246fc350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe"),
	}

	cases := []struct {
		name          string
		req           *models.Vault
		buildRepoStub func(s *mocks.MockVaultRepository)
		expectedError error
		wantError     bool
	}{
		{
			name: "success",
			req:  testVault,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Insert(gomock.Any(), testVault).Times(1).
					Return(nil)
			},
			wantError:     false,
			expectedError: nil,
		},
		{
			name: "validation_errors",
			req: &models.Vault{
				UserID:        -100,
				Name:          "test content",
				DataType:      models.Credentials,
				EncryptedData: []byte("\x246fc350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe"),
			},
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Insert(gomock.Any(), &models.Vault{
						UserID:        -100,
						Name:          "test content",
						DataType:      models.Credentials,
						EncryptedData: []byte("\x246fc350aa50c4c02361a530e8e70112c7303f59402d541c3d82995340fa02a73ddafe"),
					}).Times(0)
			},
			wantError:     true,
			expectedError: nil,
		},
		{
			name: "repository_error",
			req:  testVault,
			buildRepoStub: func(r *mocks.MockVaultRepository) {
				r.EXPECT().
					Insert(gomock.Any(), testVault).Times(1).
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
			_, err := s.Save(context.Background(), tc.req)

			if tc.wantError {
				require.Error(t, err)
				if tc.name != "validation_errors" {
					require.ErrorIs(t, err, tc.expectedError)
				}
			} else {
				require.NoError(t, err)
			}

		})
	}
}
