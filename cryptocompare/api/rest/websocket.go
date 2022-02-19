package rest

import (
	"bequant-tt/pkg/websocket"
	"encoding/json"
	"net/http"
)

type wsRequestData struct {
	Fsyms string `json:"fsyms"`
	Tsyms string `json:"tsyms"`
}

func (rest *Rest) websocketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw := &websocket.RW{
		W: w,
		R: r,
	}

	handleFunc := func(in []byte, f func(out []byte)) {
		var data wsRequestData
		err := json.Unmarshal(in, &data)
		if err != nil {
			f([]byte(err.Error()))
			return
		}

		compares, err := rest.s.App.Get(ctx, []string{data.Fsyms}, []string{data.Tsyms})
		if err != nil {
			f([]byte(err.Error()))
			return
		}

		f(ComparesToResponse(compares).Message())
	}

	_, err := websocket.NewConnection(rw, handleFunc, func() {})
	if err != nil {
		return
	}
}
