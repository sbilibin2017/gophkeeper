package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestRegisterService_Register_Table(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFacade := NewMockRegisterer(ctrl)

	service := NewRegisterService(mockFacade)

	testSecret := &models.UsernamePassword{
		Username: "testuser",
		Password: "testpass",
	}

	tests := []struct {
		name          string
		mockReturn    func()
		expectedToken string
		expectedError bool
	}{
		{
			name: "successful registration",
			mockReturn: func() {
				mockFacade.EXPECT().
					Register(gomock.Any(), testSecret).
					Return("token123", nil)
			},
			expectedToken: "token123",
			expectedError: false,
		},
		{
			name: "registration error",
			mockReturn: func() {
				mockFacade.EXPECT().
					Register(gomock.Any(), testSecret).
					Return("", errors.New("fail"))
			},
			expectedToken: "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockReturn()

			token, err := service.Register(context.Background(), testSecret)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedToken, token)
		})
	}
}
