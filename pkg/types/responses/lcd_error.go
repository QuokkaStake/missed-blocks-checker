package responses

type LCDError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
