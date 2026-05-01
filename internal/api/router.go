package api

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler())
	mux.HandleFunc("/check", checkHandler())
	mux.HandleFunc("/latest", latestHandler())
	mux.HandleFunc("/history", historyHandler())

	return mux
}
