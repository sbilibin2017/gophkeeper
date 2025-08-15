package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register_SuccessAndExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockUserReader(ctrl)
	mockUserWriter := NewMockUserWriter(ctrl)
	mockDeviceReader := NewMockDeviceReader(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)
	mockTokenGen := NewMockTokenGenerator(ctrl)

	auth := NewAuthService(
		mockUserReader,
		mockUserWriter,
		mockDeviceReader,
		mockDeviceWriter,
		mockTokenGen,
	)

	tests := []struct {
		name           string
		existingUser   bool
		existingDevice bool
		expectedErr    error
	}{
		{
			name:           "successful registration",
			existingUser:   false,
			existingDevice: false,
			expectedErr:    nil,
		},
		{
			name:           "user already exists",
			existingUser:   true,
			existingDevice: false,
			expectedErr:    ErrUserExists,
		},
		{
			name:           "device already exists",
			existingUser:   false,
			existingDevice: true,
			expectedErr:    ErrDeviceExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			username := "user1"
			password := "pass"
			deviceID := "dev1"

			if tt.existingUser {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(&models.UserDB{}, nil)
			} else {
				mockUserReader.EXPECT().GetByUsername(ctx, username).Return(nil, nil)
				mockUserWriter.EXPECT().Save(ctx, gomock.Any(), username, gomock.Any()).Return(nil)

				if tt.existingDevice {
					mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(&models.DeviceDB{}, nil)
				} else {
					mockDeviceReader.EXPECT().GetByID(ctx, deviceID).Return(nil, nil)
					mockDeviceWriter.EXPECT().Save(ctx, deviceID, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
					mockTokenGen.EXPECT().Generate(gomock.Any()).Return("token123", nil)
				}
			}

			privKey, token, err := auth.Register(ctx, username, password, deviceID)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Nil(t, privKey)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, privKey)
				assert.Equal(t, "token123", token)
			}
		})
	}
}

func TestAuthService_Register_ErrorCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(
			ur *MockUserReader,
			uw *MockUserWriter,
			dr *MockDeviceReader,
			dw *MockDeviceWriter,
			tg *MockTokenGenerator,
		)
		expectedErr error
	}{
		{
			name: "GetByUsername error",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "User already exists",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(&models.UserDB{}, nil)
			},
			expectedErr: ErrUserExists,
		},
		{
			name: "Save user error",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, nil)
				uw.EXPECT().Save(ctx, gomock.Any(), "user1", gomock.Any()).Return(errors.New("save user error"))
			},
			expectedErr: errors.New("save user error"),
		},
		{
			name: "GetByID error",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, nil)
				uw.EXPECT().Save(ctx, gomock.Any(), "user1", gomock.Any()).Return(nil)
				dr.EXPECT().GetByID(ctx, "dev1").Return(nil, errors.New("device lookup error"))
			},
			expectedErr: errors.New("device lookup error"),
		},
		{
			name: "Device already exists",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, nil)
				uw.EXPECT().Save(ctx, gomock.Any(), "user1", gomock.Any()).Return(nil)
				dr.EXPECT().GetByID(ctx, "dev1").Return(&models.DeviceDB{}, nil)
			},
			expectedErr: ErrDeviceExists,
		},
		{
			name: "Save device error",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, nil)
				uw.EXPECT().Save(ctx, gomock.Any(), "user1", gomock.Any()).Return(nil)
				dr.EXPECT().GetByID(ctx, "dev1").Return(nil, nil)
				dw.EXPECT().Save(ctx, "dev1", gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("save device error"))
			},
			expectedErr: errors.New("save device error"),
		},
		{
			name: "Token generation error",
			setupMocks: func(ur *MockUserReader, uw *MockUserWriter, dr *MockDeviceReader, dw *MockDeviceWriter, tg *MockTokenGenerator) {
				ur.EXPECT().GetByUsername(ctx, "user1").Return(nil, nil)
				uw.EXPECT().Save(ctx, gomock.Any(), "user1", gomock.Any()).Return(nil)
				dr.EXPECT().GetByID(ctx, "dev1").Return(nil, nil)
				dw.EXPECT().Save(ctx, "dev1", gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				tg.EXPECT().Generate(gomock.Any()).Return("", errors.New("token error"))
			},
			expectedErr: errors.New("token error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserReader := NewMockUserReader(ctrl)
			mockUserWriter := NewMockUserWriter(ctrl)
			mockDeviceReader := NewMockDeviceReader(ctrl)
			mockDeviceWriter := NewMockDeviceWriter(ctrl)
			mockTokenGen := NewMockTokenGenerator(ctrl)

			tt.setupMocks(mockUserReader, mockUserWriter, mockDeviceReader, mockDeviceWriter, mockTokenGen)

			service := NewAuthService(
				mockUserReader,
				mockUserWriter,
				mockDeviceReader,
				mockDeviceWriter,
				mockTokenGen,
			)

			privKey, token, err := service.Register(ctx, "user1", "pass", "dev1")

			assert.Error(t, err)
			assert.EqualError(t, err, tt.expectedErr.Error())
			assert.Nil(t, privKey)
			assert.Empty(t, token)
		})
	}
}
