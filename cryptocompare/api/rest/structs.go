package rest

import (
	"bequant-tt/core"
	"encoding/json"
)

type CompareResponse struct {
	Raw     map[string]Compares `json:"RAW"`
	Display map[string]Compares `json:"DISPLAY"`
}

func (c *CompareResponse) Message() []byte {
	bytes, err := json.Marshal(c)
	if err != nil {
		return []byte(err.Error())
	}

	return bytes
}

func ComparesToResponse(compares []*core.Compare) *CompareResponse {
	resp := &CompareResponse{
		Raw:     make(map[string]Compares),
		Display: make(map[string]Compares),
	}

	for _, compare := range compares {
		resp.Raw[compare.Fsym], resp.Display[compare.Fsym] = convertTsyms(compare.Tsyms)
	}

	return resp
}

type Compares map[string]json.RawMessage

func convertTsyms(tsyms map[string]*core.CompareData) (raw Compares, display Compares) {
	raw = make(map[string]json.RawMessage)
	display = make(map[string]json.RawMessage)

	for tsym, value := range tsyms {
		raw[tsym] = json.RawMessage(value.Raw)
		display[tsym] = json.RawMessage(value.Display)
	}

	return raw, display
}
