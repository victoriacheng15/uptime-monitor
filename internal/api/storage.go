package api

import (
	"errors"
	"net/http"
)

var errMissingHistoryBucket = errors.New("HISTORY_BUCKET is not configured")

func latestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		store, err := newPersistence(r.Context())
		if err != nil {
			http.Error(w, "storage is not available", http.StatusInternalServerError)
			return
		}

		latest, err := store.Latest(r.Context())
		if err != nil {
			http.Error(w, "failed to load latest results", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, latest)
	}
}

func historyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		store, err := newPersistence(r.Context())
		if err != nil {
			http.Error(w, "storage is not available", http.StatusInternalServerError)
			return
		}

		history, err := store.History(r.Context())
		if err != nil {
			http.Error(w, "failed to load history results", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, history)
	}
}
