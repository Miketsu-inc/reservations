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
	stratch *priceScratch
}

type FormattedPrice struct {
	currency.Amount
}

type priceScratch struct {
	number       pgtype.Numeric
	currencyCode string
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

// IsNull implements pgtype.CompositeIndexGetter.
// Price is a value type; use *Price for nullable columns.
func (p Price) IsNull() bool {
	return false
}

// Index implements pgtype.CompositeIndexGetter.
func (p Price) Index(i int) any {
	switch i {
	case 0:
		n := new(pgtype.Numeric)
		if err := n.Scan(p.Number()); err != nil {
			panic(fmt.Sprintf("price index: scanning numeric: %q: %v", p.Number(), err))
		}
		return n
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
	if p.stratch == nil {
		p.stratch = new(priceScratch)
	}

	switch i {
	case 0:
		return &p.stratch.number
	case 1:
		return &p.stratch.currencyCode
	default:
		panic(fmt.Sprintf("price scanIndex: unexpected index %d", i))
	}
}

func (p *Price) Scan(src any) error {
	if p.stratch == nil {
		return fmt.Errorf("price: Scan called without prior ScanIndex (src: %T)", src)
	}

	s := p.stratch
	p.stratch = nil

	if !s.number.Valid {
		return fmt.Errorf("price: number cannot be null")
	}

	if s.currencyCode == "" {
		return fmt.Errorf("price: empty currency code")
	}

	v, err := s.number.Value()
	if err != nil {
		return fmt.Errorf("price: reading numeric value: %w", err)
	}

	numStr, ok := v.(string)
	if !ok {
		return fmt.Errorf("price: unexpected numeric type: %T", v)
	}

	a, err := currency.NewAmount(numStr, s.currencyCode)
	if err != nil {
		return fmt.Errorf("price: error constructing amount (%q, %q): %w", numStr, s.currencyCode, err)
	}

	p.Amount = a
	return nil
}
