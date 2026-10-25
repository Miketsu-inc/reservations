package lang

import (
	"net/http"

	"golang.org/x/text/language"
)

// MatchStrings defaults to the first language in this list
// if it does not match anything
var supportedLanguages = language.NewMatcher([]language.Tag{
	language.Hungarian,
	language.English,
})

func GetLangFromHeader(r *http.Request) language.Tag {
	match, _ := language.MatchStrings(supportedLanguages, r.Header.Get("Accept-Language"))
	tag, _ := match.Base()
	return language.Make(tag.String())
}

func GetDefaultLang() language.Tag {
	tag, _, _ := supportedLanguages.Match(language.Tag{})
	return tag
}
