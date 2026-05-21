package bank

// DTO для HTTP и будущих Kafka-сообщений
type ProfileResponse struct {
	Phone       string  `json:"phone"`
	DisplayName string  `json:"displayName"`
	Role        string  `json:"role"`
	Balance     float64 `json:"balance"`
	CardNumber  string  `json:"cardNumber,omitempty"`
	ExpDate     string  `json:"expDate,omitempty"`
	CVV         string  `json:"cvv,omitempty"`
	CardHolder  string  `json:"cardHolder,omitempty"`
}

type TransactionResponse struct {
	ID                uint    `json:"id"`
	Type              string  `json:"type"`
	Amount            float64 `json:"amount"`
	CounterpartyPhone      string  `json:"counterpartyPhone,omitempty"`
	CounterpartyCardNumber string  `json:"counterpartyCardNumber,omitempty"`
	BalanceAfter      float64 `json:"balanceAfter"`
	CreatedAt         string  `json:"createdAt"`
}
