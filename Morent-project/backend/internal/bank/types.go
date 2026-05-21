package bank

// DTO для HTTP и будущих Kafka-сообщений
type ProfileResponse struct {
	Phone       string  `json:"phone"`
	DisplayName string  `json:"displayName"`
	Role        string  `json:"role"`
	Balance     float64 `json:"balance"`
}

type TransactionResponse struct {
	ID                uint    `json:"id"`
	Type              string  `json:"type"`
	Amount            float64 `json:"amount"`
	CounterpartyPhone string  `json:"counterpartyPhone,omitempty"`
	BalanceAfter      float64 `json:"balanceAfter"`
	CreatedAt         string  `json:"createdAt"`
}
