package lang

import (
	"context"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"golang.org/x/text/language"
)

// MatchStrings defaults to the first language in this list
// if it does not match anything
var supportedLanguages = language.NewMatcher([]language.Tag{
	language.Hungarian,
	language.English,
})

type contextKey struct {
	name string
}

var langCtxKey = &contextKey{"lang"}

// Return the user's language from the context
func LangFromContext(ctx context.Context) language.Tag {
	lang, ok := ctx.Value(langCtxKey).(language.Tag)
	assert.True(ok, "Language not found in route context", ctx, lang)

	return lang
}

func SetLangInContext(ctx context.Context, langTag language.Tag) context.Context {
	return context.WithValue(ctx, langCtxKey, langTag)
}

func GetLangFromHeader(r *http.Request) language.Tag {
	match, _ := language.MatchStrings(supportedLanguages, r.Header.Get("Accept-Language"))
	tag, _ := match.Base()
	return language.Make(tag.String())
}

func GetDefaultLang() language.Tag {
	tag, _, _ := supportedLanguages.Match(language.Tag{})
	return tag
}
