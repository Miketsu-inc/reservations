package middleware

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"golang.org/x/text/language"
)

// Language middleware that puts the language tag in the context
// always should be called after the authentication middleware
func (m *Manager) Language(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var langTag language.Tag

		langTag = lang.GetLangFromHeader(r)

		// if the userId is not in the context assume that
		// the request did not come from a user
		userId, ok := jwt.GetUserIDFromContext(ctx)
		if ok {
			lang, err := m.userRepo.GetUserLanguage(ctx, userId)
			if err == nil {
				langTag = lang
			}
		}

		ctx = lang.SetLangInContext(ctx, langTag)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
