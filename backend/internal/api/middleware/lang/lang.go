package lang

import (
	"context"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"golang.org/x/text/language"
)

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
