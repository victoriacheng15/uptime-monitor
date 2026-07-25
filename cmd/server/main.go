// cmd/server/main.go
// Local development server only. Never deployed to Lambda.
// Runs the Go API router over HTTP and serves the SSG frontend with live reload.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"uptime-monitor/internal/api"
	"uptime-monitor/internal/models"
	"uptime-monitor/internal/web"
)

const liveReloadSnippet = `<script>
(() => {
	function connect() {
		const es = new EventSource('/dev-reload');
		es.onerror = () => {
			es.close();
			const interval = setInterval(async () => {
				try {
					await fetch('/dev-reload');
					clearInterval(interval);
					location.reload();
				} catch (e) {}
			}, 200);
		};
	}
	connect();
})();
</script>`

// localStore replaces AWS S3 with a local JSON file store in the dist/ directory.
type localStore struct {
	mu  sync.RWMutex
	dir string
}

func (s *localStore) Save(ctx context.Context, results []models.CheckResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	history := models.HistoryResponse{History: map[string][]models.CheckResult{}}
	if data, err := os.ReadFile(filepath.Join(s.dir, "history.json")); err == nil {
		_ = json.Unmarshal(data, &history)
	}
	if history.History == nil {
		history.History = map[string][]models.CheckResult{}
	}
	for _, r := range results {
		entries := append(history.History[r.URL], r)
		if len(entries) > 5 {
			entries = entries[len(entries)-5:]
		}
		history.History[r.URL] = entries
	}
	latest := models.LatestResponse{Sites: results, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
	if b, err := json.Marshal(latest); err == nil {
		_ = os.WriteFile(filepath.Join(s.dir, "latest.json"), b, 0644)
	}
	if b, err := json.Marshal(history); err == nil {
		_ = os.WriteFile(filepath.Join(s.dir, "history.json"), b, 0644)
	}
	return nil
}

func (s *localStore) Latest(ctx context.Context) (models.LatestResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var v models.LatestResponse
	data, err := os.ReadFile(filepath.Join(s.dir, "latest.json"))
	if err != nil {
		return models.LatestResponse{Sites: []models.CheckResult{}}, nil
	}
	_ = json.Unmarshal(data, &v)
	return v, nil
}

func (s *localStore) History(ctx context.Context) (models.HistoryResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := models.HistoryResponse{History: map[string][]models.CheckResult{}}
	data, err := os.ReadFile(filepath.Join(s.dir, "history.json"))
	if err != nil {
		return v, nil
	}
	_ = json.Unmarshal(data, &v)
	if v.History == nil {
		v.History = map[string][]models.CheckResult{}
	}
	return v, nil
}

// buildFrontend runs the Go SSG generator and the Tailwind CSS compiler.
func buildFrontend() {
	log.Println("[frontend] building static pages...")
	gen := web.NewGenerator(
		"internal/web/templates",
		"internal/web/templates/content",
		"dist",
	)
	gen.APIBaseURL = os.Getenv("API_BASE_URL")
	if err := gen.Generate(); err != nil {
		log.Printf("[frontend] ssg error: %v", err)
	}

	log.Println("[frontend] compiling tailwind css...")
	cmd := exec.Command("tailwindcss",
		"-i", "./internal/web/templates/styles.css",
		"-o", "./dist/styles.css",
		"--minify",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[frontend] tailwind error: %v", err)
	}
	log.Println("[frontend] build complete")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Override AWS S3 storage with local file storage.
	store := &localStore{dir: "dist"}
	api.SetNewPersistence(func(ctx context.Context) (api.ExportedPersistence, error) {
		return store, nil
	})

	if os.Getenv("MONITOR_TARGETS") == "" {
		_ = os.Setenv("MONITOR_TARGETS", "https://google.com,https://github.com")
	}
	_ = os.Setenv("API_BASE_URL", "http://localhost:"+port)

	// Build frontend synchronously on startup.
	buildFrontend()

	// Trigger a startup check to populate the local data store.
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("[backend] running startup check...")
		resp, err := http.Post(fmt.Sprintf("http://localhost:%s/check", port), "application/json", nil)
		if err != nil {
			log.Printf("[backend] startup check warning: %v", err)
			return
		}
		resp.Body.Close()
		log.Println("[backend] startup check done, rebuilding frontend with live data...")
		buildFrontend()
	}()

	apiRouter := api.NewRouter()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Backend API routes
		switch path {
		case "/health", "/check", "/latest", "/history":
			apiRouter.ServeHTTP(w, r)
			return
		}

		// SSE endpoint for browser live reload
		if path == "/dev-reload" {
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "streaming not supported", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			fmt.Fprintf(w, "data: connected\n\n")
			flusher.Flush()

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-r.Context().Done():
					return
				case <-ticker.C:
					fmt.Fprintf(w, "event: heartbeat\ndata: \n\n")
					flusher.Flush()
				}
			}
		}

		// Disable browser caching for development resources
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")

		// Frontend static file serving from dist/
		if path == "/" {
			path = "/index.html"
		}
		fullPath := filepath.Join("dist", path)
		log.Printf("[dev] serving static file: %s", fullPath)

		if strings.HasSuffix(path, ".html") {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			html := string(content) + liveReloadSnippet
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(html))
			return
		}
		http.ServeFile(w, r, fullPath)
	})

	log.Printf("[dev] server running at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
