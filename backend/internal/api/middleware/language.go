package middleware

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"golang.org/x/text/language"
)

// MatchStrings defaults to the first language in this list
// if it does not match anything
var supportedLanguages = language.NewMatcher([]language.Tag{
	language.Hungarian,
	language.English,
})

// Language middleware that puts the language tag in the context
// always should be called after the authentication middleware
func (m *Manager) Language(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var langTag language.Tag

		// if the userId is not in the context assume that
		// the request did not come from a user
		userId, ok := jwt.GetUserIDFromContext(ctx)
		if !ok {
			langTag = getLangFromHeader(r)
		} else {
			lang, err := m.userRepo.GetUserPreferredLanguage(ctx, userId)
			if err == nil && lang != nil {
				langTag = *lang
			} else {
				langTag = getLangFromHeader(r)
			}
		}

		ctx = lang.SetLangInContext(ctx, langTag)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getLangFromHeader(r *http.Request) language.Tag {
	match, _ := language.MatchStrings(supportedLanguages, r.Header.Get("Accept-Language"))
	tag, _ := match.Base()
	return language.Make(tag.String())
}
