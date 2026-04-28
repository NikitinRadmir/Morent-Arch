package dtos

type Filters struct {
	YearAfter   int `json:"year_after"`
	YearBefore  int `json:"year_before"`
	MileageFrom int `json:"mileage_from"`
	MileageTo   int `json:"mileage_to"`
	PriceFrom   int `json:"price_from"`
	PriceTo     int `json:"price_to"`
}

type UserSearchQuery struct {
	Query    string  `json:"q" binding:"required"`
	Criteria Filters `json:"filters"`
}

type SearchResponse struct {
	Info string `json:"info"`
}

type SearchTrimItem struct {
	ID           int     `json:"id"`
	Year         int     `json:"year"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Trim         string  `json:"trim"`
	Description  string  `json:"description"`
	MSRP         int     `json:"msrp"`
	Transmission string  `json:"transmission"`
	Seats        int     `json:"seats"`
	Fuel         float64 `json:"fuel"`
	ImageURL     string  `json:"imageUrl"`
}

type SearchTrimsResponse struct {
	Query string           `json:"query"`
	Count int              `json:"count"`
	Cars  []SearchTrimItem `json:"cars"`
}
