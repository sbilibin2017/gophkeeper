package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/stretchr/testify/require"
)

func TestClientRegisterApp_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsernameValidator := NewMockUsernameValidator(ctrl)
	mockPasswordValidator := NewMockPasswordValidator(ctrl)
	mockRegisterer := NewMockRegisterer(ctrl)

	tests := []struct {
		name             string
		username         string
		password         string
		setupMocks       func()
		expectedErr      bool
		expectedResponse *models.AuthResponse
	}{
		{
			name:     "valid input, registration succeeds",
			username: "validUser",
			password: "validPass",
			setupMocks: func() {
				mockUsernameValidator.EXPECT().Validate("validUser").Return(nil)
				mockPasswordValidator.EXPECT().Validate("validPass").Return(nil)
				mockRegisterer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&models.AuthResponse{Token: "token123"}, nil)
			},
			expectedErr:      false,
			expectedResponse: &models.AuthResponse{Token: "token123"},
		},
		{
			name:     "invalid username",
			username: "badUser",
			password: "validPass",
			setupMocks: func() {
				mockUsernameValidator.EXPECT().Validate("badUser").Return(errors.New("invalid username"))
			},
			expectedErr:      true,
			expectedResponse: nil,
		},
		{
			name:     "invalid password",
			username: "validUser",
			password: "badPass",
			setupMocks: func() {
				mockUsernameValidator.EXPECT().Validate("validUser").Return(nil)
				mockPasswordValidator.EXPECT().Validate("badPass").Return(errors.New("invalid password"))
			},
			expectedErr:      true,
			expectedResponse: nil,
		},
		{
			name:     "registerer returns error",
			username: "validUser",
			password: "validPass",
			setupMocks: func() {
				mockUsernameValidator.EXPECT().Validate("validUser").Return(nil)
				mockPasswordValidator.EXPECT().Validate("validPass").Return(nil)
				mockRegisterer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("register failed"))
			},
			expectedErr:      true,
			expectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			app := NewClientRegisterUsecase(mockUsernameValidator, mockPasswordValidator, mockRegisterer)

			req := models.AuthRegisterRequest{
				Username: tt.username,
				Password: tt.password,
			}

			resp, err := app.Execute(context.Background(), req)
			if tt.expectedErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResponse, resp)
			}
		})
	}
}

func TestClientLoginApp_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoginer := NewMockLoginer(ctrl)

	tests := []struct {
		name             string
		username         string
		password         string
		setupMocks       func()
		expectedErr      bool
		expectedResponse *models.AuthResponse
	}{
		{
			name:     "successful login",
			username: "user1",
			password: "pass1",
			setupMocks: func() {
				mockLoginer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(&models.AuthResponse{Token: "tokenXYZ"}, nil)
			},
			expectedErr:      false,
			expectedResponse: &models.AuthResponse{Token: "tokenXYZ"},
		},
		{
			name:     "login failure",
			username: "user1",
			password: "wrongPass",
			setupMocks: func() {
				mockLoginer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("login failed"))
			},
			expectedErr:      true,
			expectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			app := NewClientLoginUsecase(mockLoginer)

			req := models.AuthLoginRequest{
				Username: tt.username,
				Password: tt.password,
			}

			resp, err := app.Execute(context.Background(), req)
			if tt.expectedErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResponse, resp)
			}
		})
	}
}
