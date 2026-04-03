package currencyx

import (
	"encoding/json"
	"fmt"

	"github.com/bojanz/currency"
	"github.com/jackc/pgx/v5/pgtype"
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
	scanState *priceScanState
}

type FormattedPrice struct {
	currency.Amount
}

// Used for assembling the currency (Price struct)
type priceScanState struct {
	numStr string
	parent *Price
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

func (p Price) ToFormatted() FormattedPrice {
	return FormattedPrice{Amount: p.Amount}
}

func FormatPrice(price *Price) *FormattedPrice {
	if price == nil {
		return nil
	}

	fp := price.ToFormatted()
	return &fp
}

// ScanText implements pgtype.TextScanner.
// This is called automatically by pgx v5 when it finishes reading the char(3) field.
func (s *priceScanState) ScanText(v pgtype.Text) error {
	if !v.Valid || v.String == "" {
		return fmt.Errorf("price: empty currency code")
	}

	if s.numStr == "" {
		return fmt.Errorf("price: number scanned as empty")
	}

	a, err := currency.NewAmount(s.numStr, v.String)
	if err != nil {
		return fmt.Errorf("price: error constructing amount: %w", err)
	}

	// Update the parent Price struct and clear the state
	s.parent.Amount = a
	s.parent.scanState = nil
	return nil
}

// IsNull implements pgtype.CompositeIndexGetter.
// Price is a value type; use *Price for nullable columns.
func (p Price) IsNull() bool {
	return false
}

// Index implements pgtype.CompositeIndexGetter.
func (p Price) Index(i int) any {
	switch i {
	case 0:
		return p.Number()
	case 1:
		return p.CurrencyCode()
	default:
		panic(fmt.Sprintf("price index: unexpected index %d", i))
	}
}

// ScanNull implements pgtype.CompositeIndexScanner.
func (p *Price) ScanNull() error {
	*p = Price{}
	return nil
}

// ScanIndex implements pgtype.CompositeIndexScanner.
func (p *Price) ScanIndex(i int) any {
	if p.scanState == nil {
		p.scanState = &priceScanState{parent: p}
	}

	switch i {
	case 0:
		// Scan numeric directly into the state's string
		return &p.scanState.numStr
	case 1:
		// Return the state itself as the TextScanner
		return p.scanState
	default:
		panic(fmt.Sprintf("price scanIndex: unexpected index %d", i))
	}
}
