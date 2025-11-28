package middleware

import (
	"net/http"
	"strings"

	"SLGaming/back/services/gateway/internal/svc"

	"github.com/golang-jwt/jwt/v4"
)

type JWTMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewJWTMiddleware(svcCtx *svc.ServiceContext) *JWTMiddleware {
	return &JWTMiddleware{svcCtx: svcCtx}
}

func (m *JWTMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.svcCtx == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		secret := m.svcCtx.Config.JWT.Secret
		if secret == "" {
			secret = "slgaming-gateway-secret"
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
