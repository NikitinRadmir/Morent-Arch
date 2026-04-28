package dtos

type VehicleInfo struct {
	Id      int64  `db:"id" json:"id"`
	Name    string `db:"name" json:"manufacturer"`
	Model   string `db:"model" json:"model"`
	Year    int    `db:"year" json:"year"`
	Price   int    `db:"price" json:"price"`
	Mileage int    `db:"mileage" json:"mileage"`
}

type ActionMessage struct {
	Action string `json:"action"`
	Source string `json:"source"`
}
