package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretAddHandler(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
		mockSetup      func(ctrl *gomock.Controller) (SecretWriter, JWTParser)
	}{
		{
			name:       "success",
			authHeader: "Bearer validtoken",
			requestBody: SecretSaveRequest{
				SecretName: "mysecret",
				SecretType: "password",
				Ciphertext: []byte("encrypted"),
				AESKeyEnc:  []byte("keyenc"),
			},
			expectedStatus: http.StatusOK,
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				mockWriter := NewMockSecretWriter(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("validtoken").Return("alice", nil).Times(1)
				mockWriter.EXPECT().
					Save(gomock.Any(), "alice", "mysecret", "password", []byte("encrypted"), []byte("keyenc")).
					Return(nil).
					Times(1)

				return mockWriter, mockParser
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			requestBody:    nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				return nil, nil
			},
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidHeader",
			requestBody:    nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				return nil, nil
			},
		},
		{
			name:       "jwt parse error",
			authHeader: "Bearer invalidtoken",
			requestBody: SecretSaveRequest{
				SecretName: "s1",
				SecretType: "t1",
				Ciphertext: []byte("c"),
				AESKeyEnc:  []byte("k"),
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				mockParser := NewMockJWTParser(ctrl)
				mockParser.EXPECT().Parse("invalidtoken").Return("", errors.New("parse error")).Times(1)
				return nil, mockParser
			},
		},
		{
			name:           "invalid JSON body",
			authHeader:     "Bearer sometoken",
			requestBody:    "not-json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				mockParser := NewMockJWTParser(ctrl)
				mockParser.EXPECT().Parse("sometoken").Return("user", nil).Times(1)
				return nil, mockParser
			},
		},
		{
			name:       "save error",
			authHeader: "Bearer token123",
			requestBody: SecretSaveRequest{
				SecretName: "sn",
				SecretType: "st",
				Ciphertext: []byte("ct"),
				AESKeyEnc:  []byte("ak"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to save secret\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretWriter, JWTParser) {
				mockWriter := NewMockSecretWriter(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("token123").Return("bob", nil).Times(1)
				mockWriter.EXPECT().
					Save(gomock.Any(), "bob", "sn", "st", []byte("ct"), []byte("ak")).
					Return(errors.New("db failure")).
					Times(1)

				return mockWriter, mockParser
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			writer, parser := tt.mockSetup(ctrl)
			handler := NewSecretAddHandler(writer, parser)

			var bodyBytes []byte
			if tt.requestBody != nil {
				switch v := tt.requestBody.(type) {
				case string:
					bodyBytes = []byte(v)
				default:
					b, err := json.Marshal(v)
					if err != nil {
						t.Fatalf("marshal requestBody error: %v", err)
					}
					bodyBytes = b
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(bodyBytes))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestNewSecretGetHandler(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		secretType     string
		secretName     string
		expectedStatus int
		expectedBody   string
		mockSetup      func(ctrl *gomock.Controller) (SecretReader, JWTParser)
	}{
		{
			name:           "success",
			authHeader:     "Bearer validtoken",
			secretType:     "password",
			secretName:     "mysecret",
			expectedStatus: http.StatusOK,
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockReader := NewMockSecretReader(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("validtoken").Return("alice", nil).Times(1)
				mockReader.EXPECT().
					Get(gomock.Any(), "alice", "password", "mysecret").
					Return(&models.Secret{
						SecretName: "mysecret",
						SecretType: "password",
						Ciphertext: []byte("encrypted"),
						AESKeyEnc:  []byte("keyenc"),
					}, nil).
					Times(1)

				return mockReader, mockParser
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			secretType:     "password",
			secretName:     "mysecret",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				return nil, nil
			},
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidHeader",
			secretType:     "password",
			secretName:     "mysecret",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				return nil, nil
			},
		},
		{
			name:           "jwt parse error",
			authHeader:     "Bearer invalidtoken",
			secretType:     "password",
			secretName:     "mysecret",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockParser := NewMockJWTParser(ctrl)
				mockParser.EXPECT().Parse("invalidtoken").Return("", errors.New("parse error")).Times(1)
				return nil, mockParser
			},
		},
		{
			name:           "missing parameters",
			authHeader:     "Bearer validtoken",
			secretType:     "",
			secretName:     "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "missing secret_type or secret_name URL parameter\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockParser := NewMockJWTParser(ctrl)
				mockParser.EXPECT().Parse("validtoken").Return("alice", nil).Times(1)
				return nil, mockParser
			},
		},
		{
			name:           "get error",
			authHeader:     "Bearer token123",
			secretType:     "st",
			secretName:     "sn",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get secret\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockReader := NewMockSecretReader(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("token123").Return("bob", nil).Times(1)
				mockReader.EXPECT().
					Get(gomock.Any(), "bob", "st", "sn").
					Return(nil, errors.New("db failure")).
					Times(1)

				return mockReader, mockParser
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reader, parser := tt.mockSetup(ctrl)
			handler := NewSecretGetHandler(reader, parser)

			req := httptest.NewRequest(http.MethodGet, "/secrets/"+tt.secretType+"/"+tt.secretName, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// chi URL params setup (needed for chi.URLParam to work)
			routeCtx := chi.NewRouteContext()
			if tt.secretType != "" {
				routeCtx.URLParams.Add("secret_type", tt.secretType)
			}
			if tt.secretName != "" {
				routeCtx.URLParams.Add("secret_name", tt.secretName)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			} else if rec.Code == http.StatusOK {
				// On success, decode the response and verify fields
				var resp models.Secret
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.secretName, resp.SecretName)
				assert.Equal(t, tt.secretType, resp.SecretType)
			}
		})
	}
}

func TestNewSecretListHandler(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
		mockSetup      func(ctrl *gomock.Controller) (SecretReader, JWTParser)
	}{
		{
			name:           "success",
			authHeader:     "Bearer validtoken",
			expectedStatus: http.StatusOK,
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockReader := NewMockSecretReader(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("validtoken").Return("alice", nil).Times(1)
				mockReader.EXPECT().
					List(gomock.Any(), "alice").
					Return([]*models.Secret{
						{
							SecretName: "s1",
							SecretType: "t1",
							Ciphertext: []byte("c1"),
							AESKeyEnc:  []byte("k1"),
						},
						{
							SecretName: "s2",
							SecretType: "t2",
							Ciphertext: []byte("c2"),
							AESKeyEnc:  []byte("k2"),
						},
					}, nil).
					Times(1)

				return mockReader, mockParser
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				return nil, nil
			},
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "InvalidHeader",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				return nil, nil
			},
		},
		{
			name:           "jwt parse error",
			authHeader:     "Bearer invalidtoken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   ErrUnauthorized.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockParser := NewMockJWTParser(ctrl)
				mockParser.EXPECT().Parse("invalidtoken").Return("", errors.New("parse error")).Times(1)
				return nil, mockParser
			},
		},
		{
			name:           "list error",
			authHeader:     "Bearer token123",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to list secrets\n",
			mockSetup: func(ctrl *gomock.Controller) (SecretReader, JWTParser) {
				mockReader := NewMockSecretReader(ctrl)
				mockParser := NewMockJWTParser(ctrl)

				mockParser.EXPECT().Parse("token123").Return("bob", nil).Times(1)
				mockReader.EXPECT().
					List(gomock.Any(), "bob").
					Return(nil, errors.New("db failure")).
					Times(1)

				return mockReader, mockParser
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			reader, parser := tt.mockSetup(ctrl)
			handler := NewSecretListHandler(reader, parser)

			req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			} else if rec.Code == http.StatusOK {
				// On success, decode the response and verify
				var resp []*models.Secret
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp)
			}
		})
	}
}
