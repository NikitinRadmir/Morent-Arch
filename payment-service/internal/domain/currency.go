package domain

var supportedCurrencies = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"GBP": {},
	"RUB": {},
}

func IsSupportedCurrency(code string) bool {
	_, ok := supportedCurrencies[code]
	return ok
}
