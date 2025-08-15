package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -source=tx_middleware.go -destination=mocks_txsetter_test.go -package=middlewares TxSetter

func TestNewTxMiddleware_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxSetter := NewMockTxSetter(ctrl)

	ctx := context.Background()
	expectedCtx := context.WithValue(ctx, "tx", "mock-tx")

	mockTxSetter.EXPECT().
		Set(ctx).
		Return(expectedCtx, nil)

	mw := NewTxMiddleware(mockTxSetter)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, expectedCtx, r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called, "Handler должен быть вызван")
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestNewTxMiddleware_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTxSetter := NewMockTxSetter(ctrl)

	ctx := context.Background()
	mockTxSetter.EXPECT().
		Set(ctx).
		Return(ctx, errors.New("tx error"))

	mw := NewTxMiddleware(mockTxSetter)

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, called, "Handler не должен быть вызван при ошибке")
	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
