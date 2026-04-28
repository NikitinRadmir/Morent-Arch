package dtos

type LoginRequest struct {
	APIToken  string `json:"api_token"`
	APISecret string `json:"api_secret"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type TrimsResponse struct {
	Collection Collection `json:"collection"`
	Data       []Trim     `json:"data"`
}

type BodiesResponse struct {
	Collection Collection `json:"collection"`
	Data       []Body     `json:"data"`
}

type EnginesResponse struct {
	Collection Collection `json:"collection"`
	Data       []Engine   `json:"data"`
}

type Collection struct {
	URL   string `json:"url"`
	Count int    `json:"count"`
	Pages int    `json:"pages"`
	Total int    `json:"total"`
	Next  string `json:"next"`
	Prev  string `json:"prev"`
	First string `json:"first"`
	Last  string `json:"last"`
}

type Trim struct {
	ID          int     `json:"id"`
	MakeID      int     `json:"make_id"`
	ModelID     int     `json:"model_id"`
	SubmodelID  int     `json:"submodel_id"`
	Year        int     `json:"year"`
	Make        string  `json:"make"`
	Model       string  `json:"model"`
	Series      *string `json:"series"`
	Submodel    string  `json:"submodel"`
	Trim        string  `json:"trim"`
	Description string  `json:"description"`
	MSRP        int     `json:"msrp"`
	Invoice     int     `json:"invoice"`
	Created     string  `json:"created"`
	Modified    string  `json:"modified"`
}

type Body struct {
	ID         int    `json:"id"`
	MakeID     int    `json:"make_id"`
	ModelID    int    `json:"model_id"`
	SubmodelID int    `json:"submodel_id"`
	TrimID     int    `json:"trim_id"`
	Year       int    `json:"year"`
	Make       string `json:"make"`
	Model      string `json:"model"`
	Submodel   string `json:"submodel"`
	Trim       string `json:"trim"`
	Type       string `json:"type"`
	Doors      int    `json:"doors"`
	Seats      int    `json:"seats"`
}

type Engine struct {
	ID           int    `json:"id"`
	MakeID       int    `json:"make_id"`
	ModelID      int    `json:"model_id"`
	SubmodelID   int    `json:"submodel_id"`
	TrimID       int    `json:"trim_id"`
	Year         int    `json:"year"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Submodel     string `json:"submodel"`
	Trim         string `json:"trim"`
	EngineType   string `json:"engine_type"`
	FuelType     string `json:"fuel_type"`
	Cylinders    string `json:"cylinders"`
	Size         string `json:"size"`
	HorsepowerHP int    `json:"horsepower_hp"`
	TorqueFtLbs  *int   `json:"torque_ft_lbs"`
	Transmission string `json:"transmission"`
}
