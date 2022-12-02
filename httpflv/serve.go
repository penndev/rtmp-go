package httpflv

import (
	"net/http"
)

type flvHandle func(http.ResponseWriter, *http.Request)

func Serve(f flvHandle) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", f)

	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		return err
	}
	return nil
}
