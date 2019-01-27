package ui

import (
	"net/http"

	"github.com/gobuffalo/packr"
	"github.com/heroiclabs/nakama/runtime"
)

func RegisterUI(init runtime.Initializer) error {
	go func() {
		statics := packr.NewBox("./static")

		srv := http.NewServeMux()
		srv.Handle("/", http.FileServer(statics))

		http.ListenAndServe(":8090", srv)
	}()
	return nil
}
