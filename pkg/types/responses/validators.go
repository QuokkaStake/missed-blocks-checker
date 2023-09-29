package responses

type ValidatorsResponse struct {
	Result ValidatorsResult `json:"result"`
	Error  *ResponseError   `json:"error"`
}

type ValidatorsResult struct {
	Validators []HistoricalValidator `json:"validators"`
	Count      string                `json:"count"`
	Total      string                `json:"total"`
}

type HistoricalValidator struct {
	Address string `json:"address"`
}
