package responses

import "encoding/json"

func (r *ValidatorsResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if result, ok := v["result"]; !ok {
		r.Result = nil
	} else {
		rawBytes, _ := json.Marshal(result) //nolint:errchkjson

		var resultParsed ValidatorsResult
		if err := json.Unmarshal(rawBytes, &resultParsed); err != nil {
			return err
		}

		r.Result = &resultParsed
	}

	if responseError, ok := v["error"]; !ok {
		r.Error = nil
	} else {
		rawBytes, _ := json.Marshal(responseError) //nolint:errchkjson

		var errorParsed ResponseError
		if err := json.Unmarshal(rawBytes, &errorParsed); err != nil {
			return err
		}

		r.Error = &errorParsed
	}

	return nil
}

type ValidatorsResponse struct {
	Result *ValidatorsResult `json:"result"`
	Error  *ResponseError    `json:"error"`
}

type ValidatorsResult struct {
	Validators []HistoricalValidator `json:"validators"`
	Count      string                `json:"count"`
	Total      string                `json:"total"`
}

type HistoricalValidator struct {
	Address string `json:"address"`
}
