package api

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler())
	mux.HandleFunc("/check", checkHandler())

	return mux
}
