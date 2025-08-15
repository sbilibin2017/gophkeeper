package middlewares

import (
	"context"
	"net/http"
)

// TxSetter описывает объект, который умеет устанавливать транзакцию в контекст.
type TxSetter interface {
	Set(ctx context.Context) (context.Context, error)
}

// NewTxMiddleware создает middleware, который устанавливает транзакцию в контекст запроса.
func NewTxMiddleware(txSetter TxSetter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := txSetter.Set(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
