package currencyx

import (
	"encoding/json"

	"github.com/bojanz/currency"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	curr "golang.org/x/text/currency"
	"golang.org/x/text/language"
)

var currencyLocale = map[string]string{
	"HUF": "hu",
	"EUR": "en",
	"USD": "en",
	"GBP": "en",
}

var currencyPrecision = map[string]int{
	"HUF": 0,
	"EUR": 2,
	"USD": 2,
	"GBP": 2,
}

var formatters = make(map[string]*currency.Formatter)

func init() {
	for currencyCode, loc := range currencyLocale {
		locale := currency.NewLocale(loc)
		formatter := currency.NewFormatter(locale)

		prec, ok := currencyPrecision[currencyCode]
		assert.True(ok, "Precision for this currency code is not defined", currencyCode, currencyPrecision, currencyLocale)

		formatter.MaxDigits = uint8(prec)
		formatters[currencyCode] = formatter
	}
}

func Format(amount currency.Amount) string {
	formatter, ok := formatters[amount.CurrencyCode()]
	assert.True(ok, "Currency code does not exist in currencyLocale map", amount.CurrencyCode(), currencyLocale)

	return formatter.Format(amount)
}

// finds the most likely currency based on the user's language
// falls back on HUF if currency is not supported
func FindBest(lang language.Tag) string {
	currencyUnit, _ := curr.FromTag(lang)
	currency := currencyUnit.String()

	_, ok := currencyLocale[currency]
	if !ok {
		currency = "HUF"
	}

	return currency
}

type Price struct {
	currency.Amount
}

type FormattedPrice struct {
	currency.Amount
}

func (f FormattedPrice) MarshalJSON() ([]byte, error) {
	var priceStr string
	if f.IsZero() {
		priceStr = ""
	} else {
		priceStr = Format(f.Amount)
	}
	return json.Marshal(priceStr)
}
