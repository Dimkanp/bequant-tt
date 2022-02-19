package rest

import (
	"log"
	"net/http"
)

type priceRequestData struct {
	Fsyms []string `schema:"fsyms"`
	Tsyms []string `schema:"tsyms"`
}

func (rest *Rest) priceHandler(w http.ResponseWriter, r *http.Request) {
	var data priceRequestData
	err := rest.decoder.Decode(&data, r.Form)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	compares, err := rest.s.App.Get(r.Context(), data.Fsyms, data.Tsyms)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	_, err = w.Write(ComparesToResponse(compares).Message())
	if err != nil {
		log.Printf("write response: %s\n", err)
		return
	}
}
