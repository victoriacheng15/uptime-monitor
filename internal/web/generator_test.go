package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator("tmpl", "cnt", "out")
	if gen.TemplatesDir != "tmpl" || gen.ContentDir != "cnt" || gen.OutputDir != "out" {
		t.Errorf("constructor failed to assign fields correctly")
	}
}

func TestAverageLatencyByURL(t *testing.T) {
	// 1. Nil history response
	if got := averageLatencyByURL(nil); got != nil {
		t.Errorf("expected nil for nil history, got %v", got)
	}

	// 2. Empty history response
	empty := &HistoryResponse{History: map[string][]CheckResult{}}
	if got := averageLatencyByURL(empty); got != nil {
		t.Errorf("expected nil for empty history, got %v", got)
	}

	// 3. History with empty slice for a URL
	hasEmptySlice := &HistoryResponse{
		History: map[string][]CheckResult{
			"https://example.com": {},
		},
	}
	if got := averageLatencyByURL(hasEmptySlice); got != nil {
		t.Errorf("expected nil when all slices are empty, got %v", got)
	}

	// 4. Normal history latency average and ordering
	normal := &HistoryResponse{
		History: map[string][]CheckResult{
			"https://site-a.com": {
				{URL: "https://site-a.com", LatencyMS: 100},
				{URL: "https://site-a.com", LatencyMS: 200},
			},
			"https://site-b.com": {
				{URL: "https://site-b.com", LatencyMS: 300},
				{URL: "https://site-b.com", LatencyMS: 500},
			},
		},
	}
	averages := averageLatencyByURL(normal)
	if len(averages) != 2 {
		t.Fatalf("expected 2 averages, got %d", len(averages))
	}
	// Sorted by AverageLatencyMS descending: site-b (400) should be before site-a (150)
	if averages[0].URL != "https://site-b.com" || averages[0].AverageLatencyMS != 400 {
		t.Errorf("expected first to be site-b with average 400, got %s with %d", averages[0].URL, averages[0].AverageLatencyMS)
	}
	if averages[1].URL != "https://site-a.com" || averages[1].AverageLatencyMS != 150 {
		t.Errorf("expected second to be site-a with average 150, got %s with %d", averages[1].URL, averages[1].AverageLatencyMS)
	}
}

func TestReverseEvolution(t *testing.T) {
	cfg := &EvolutionConfig{
		Chapters: []Chapter{
			{
				Title: "Chapter 1",
				Timeline: []Timeline{
					{Title: "Event A"},
					{Title: "Event B"},
				},
			},
			{
				Title: "Chapter 2",
				Timeline: []Timeline{
					{Title: "Event C"},
				},
			},
		},
	}

	gen := NewGenerator("", "", "")
	gen.reverseEvolution(cfg)

	// Chapters should be reversed: Chapter 2 first, Chapter 1 second
	if cfg.Chapters[0].Title != "Chapter 2" {
		t.Errorf("expected first chapter to be Chapter 2, got %q", cfg.Chapters[0].Title)
	}
	// Inside Chapter 1, Timeline should be reversed: Event B first, Event A second
	if cfg.Chapters[1].Timeline[0].Title != "Event B" {
		t.Errorf("expected first event of Chapter 1 to be Event B, got %q", cfg.Chapters[1].Timeline[0].Title)
	}
}

func TestGetFuncs(t *testing.T) {
	gen := NewGenerator("", "", "")
	funcs := gen.getFuncs()

	// Test dict function
	dictFn, ok := funcs["dict"].(func(...interface{}) (map[string]interface{}, error))
	if !ok {
		t.Fatal("dict func not found or has invalid signature")
	}

	// 1. Valid call
	res, err := dictFn("key1", "val1", "key2", 42)
	if err != nil {
		t.Fatalf("unexpected dict error: %v", err)
	}
	if res["key1"] != "val1" || res["key2"] != 42 {
		t.Errorf("incorrect dict results: %v", res)
	}

	// 2. Odd number of arguments
	_, err = dictFn("key1")
	if err == nil {
		t.Error("expected error for odd number of arguments")
	}

	// 3. Non-string key
	_, err = dictFn(123, "value")
	if err == nil {
		t.Error("expected error for non-string key")
	}

	// Test list function
	listFn, ok := funcs["list"].(func(...interface{}) []interface{})
	if !ok {
		t.Fatal("list func not found or has invalid signature")
	}

	listRes := listFn("a", 1, true)
	if len(listRes) != 3 || listRes[0] != "a" || listRes[1] != 1 || listRes[2] != true {
		t.Errorf("incorrect list results: %v", listRes)
	}
}

func TestFetchAPIData(t *testing.T) {
	latestData := LatestResponse{
		Sites: []CheckResult{
			{URL: "https://example.com", IsUp: true},
		},
		UpdatedAt: time.Now(),
	}

	historyData := HistoryResponse{
		History: map[string][]CheckResult{
			"https://example.com": {
				{URL: "https://example.com", IsUp: true},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(latestData)
			return
		}
		if r.URL.Path == "/history" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(historyData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	gen := NewGenerator("", "", "")
	gen.APIBaseURL = server.URL

	// 1. Fetch Uptime success
	latestRes, err := gen.fetchUptimeData()
	if err != nil {
		t.Fatalf("unexpected fetchUptimeData error: %v", err)
	}
	if len(latestRes.Sites) != 1 || latestRes.Sites[0].URL != "https://example.com" {
		t.Errorf("incorrect fetchUptimeData result: %v", latestRes)
	}

	// 2. Fetch History success
	historyRes, err := gen.fetchHistoryData()
	if err != nil {
		t.Fatalf("unexpected fetchHistoryData error: %v", err)
	}
	if len(historyRes.History["https://example.com"]) != 1 {
		t.Errorf("incorrect fetchHistoryData result: %v", historyRes)
	}
}

func TestFetchAPIDataErrors(t *testing.T) {
	// Server returning HTTP error
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errServer.Close()

	gen := NewGenerator("", "", "")

	// Invalid API Base URL (unreachable)
	gen.APIBaseURL = "http://invalid-dns-name-unreachable-1234.local"
	_, err := gen.fetchUptimeData()
	if err == nil {
		t.Error("expected error for unreachable fetchUptimeData")
	}
	_, err = gen.fetchHistoryData()
	if err == nil {
		t.Error("expected error for unreachable fetchHistoryData")
	}

	// Server returning 500 error
	gen.APIBaseURL = errServer.URL
	_, err = gen.fetchUptimeData()
	if err == nil || err.Error() == "" {
		t.Error("expected status error for fetchUptimeData")
	}
	_, err = gen.fetchHistoryData()
	if err == nil || err.Error() == "" {
		t.Error("expected status error for fetchHistoryData")
	}

	// Server returning malformed JSON
	malformedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json}"))
	}))
	defer malformedServer.Close()

	gen.APIBaseURL = malformedServer.URL
	_, err = gen.fetchUptimeData()
	if err == nil {
		t.Error("expected json unmarshal error for fetchUptimeData")
	}
	_, err = gen.fetchHistoryData()
	if err == nil {
		t.Error("expected json unmarshal error for fetchHistoryData")
	}
}

func TestGenerateAndRendering(t *testing.T) {
	// Set up temporary output directory
	outDir, err := os.MkdirTemp("", "monitor-out-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(outDir)

	// Since we run in the web package, templates/ and templates/content/ are local
	templatesDir := "templates"
	contentDir := filepath.Join("templates", "content")

	// 1. Test Generate with empty API (no dynamic fetching)
	gen := NewGenerator(templatesDir, contentDir, outDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate() failed with empty APIBaseURL: %v", err)
	}

	// Verify outputs exist
	expectedFiles := []string{"index.html", "monitor.html", "evolution.html", "llms.txt"}
	for _, filename := range expectedFiles {
		path := filepath.Join(outDir, filename)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected output file %s does not exist: %v", filename, err)
		}
	}

	// 2. Test Generate with a mocked API server returning valid responses
	latestData := LatestResponse{
		Sites: []CheckResult{
			{URL: "https://example.com", IsUp: true, LatencyMS: 50, StatusCode: 200, Timestamp: time.Now()},
		},
		UpdatedAt: time.Now(),
	}

	historyData := HistoryResponse{
		History: map[string][]CheckResult{
			"https://example.com": {
				{URL: "https://example.com", IsUp: true, LatencyMS: 50, StatusCode: 200, Timestamp: time.Now()},
			},
		},
	}

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/latest" {
			json.NewEncoder(w).Encode(latestData)
		} else if r.URL.Path == "/history" {
			json.NewEncoder(w).Encode(historyData)
		}
	}))
	defer apiServer.Close()

	genWithAPI := NewGenerator(templatesDir, contentDir, outDir)
	genWithAPI.APIBaseURL = apiServer.URL
	if err := genWithAPI.Generate(); err != nil {
		t.Fatalf("Generate() failed with valid APIBaseURL: %v", err)
	}

	// 3. Test Generate with a failing API server (should print warnings but still output successfully)
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	genWithFailingAPI := NewGenerator(templatesDir, contentDir, outDir)
	genWithFailingAPI.APIBaseURL = failingServer.URL
	if err := genWithFailingAPI.Generate(); err != nil {
		t.Fatalf("Generate() should succeed and log warnings even if API fails, but got error: %v", err)
	}
}

func TestGenerateErrors(t *testing.T) {
	// 1. Output directory creation error
	// On Linux, creating a directory at a path containing a file (instead of a directory) will fail.
	tmpFile, err := os.CreateTemp("", "temp-file-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	genBadOut := NewGenerator("templates", filepath.Join("templates", "content"), tmpFile.Name())
	if err := genBadOut.Generate(); err == nil {
		t.Error("expected error when output directory cannot be created")
	}

	// 2. Missing Landing Config
	genBadLanding := NewGenerator("templates", "invalid-content-dir", "out")
	if err := genBadLanding.Generate(); err == nil {
		t.Error("expected error when landing.yaml cannot be loaded")
	}

	// 3. Missing Evolution Config
	// Let's create a temp content dir with landing.yaml but missing evolution.yaml
	tempContentDir, err := os.MkdirTemp("", "content-dir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempContentDir)

	landingData := []byte("header:\n  project_name: Test\n")
	if err := os.WriteFile(filepath.Join(tempContentDir, "landing.yaml"), landingData, 0644); err != nil {
		t.Fatalf("failed to write landing.yaml: %v", err)
	}

	genBadEvo := NewGenerator("templates", tempContentDir, "out")
	if err := genBadEvo.Generate(); err == nil {
		t.Error("expected error when evolution.yaml cannot be loaded")
	}

	// 4. Bad YAML syntax
	badYaml := []byte("{invalid-yaml:")
	if err := os.WriteFile(filepath.Join(tempContentDir, "landing.yaml"), badYaml, 0644); err != nil {
		t.Fatalf("failed to write landing.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempContentDir, "evolution.yaml"), badYaml, 0644); err != nil {
		t.Fatalf("failed to write evolution.yaml: %v", err)
	}

	genBadYamlLanding := NewGenerator("templates", tempContentDir, "out")
	if err := genBadYamlLanding.Generate(); err == nil {
		t.Error("expected unmarshal error for invalid landing.yaml")
	}

	// Recover valid landing.yaml, test bad evolution.yaml unmarshal
	if err := os.WriteFile(filepath.Join(tempContentDir, "landing.yaml"), landingData, 0644); err != nil {
		t.Fatalf("failed to write landing.yaml: %v", err)
	}
	genBadYamlEvo := NewGenerator("templates", tempContentDir, "out")
	if err := genBadYamlEvo.Generate(); err == nil {
		t.Error("expected unmarshal error for invalid evolution.yaml")
	}

	// 5. Template file rendering errors (missing template files)
	// Create a temp templates dir with base.html but missing index.html
	tempTemplatesDir, err := os.MkdirTemp("", "templates-dir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempTemplatesDir)

	if err := os.WriteFile(filepath.Join(tempTemplatesDir, "base.html"), []byte("{{template \"content\" .}}"), 0644); err != nil {
		t.Fatalf("failed to write base.html: %v", err)
	}

	// Write valid content configs to tempContentDir
	if err := os.WriteFile(filepath.Join(tempContentDir, "evolution.yaml"), []byte("page_title: Evolution\n"), 0644); err != nil {
		t.Fatalf("failed to write evolution.yaml: %v", err)
	}

	tempOutDir, err := os.MkdirTemp("", "out-dir-*")
	if err != nil {
		t.Fatalf("failed to create temp out dir: %v", err)
	}
	defer os.RemoveAll(tempOutDir)

	genBadTemplate := NewGenerator(tempTemplatesDir, tempContentDir, tempOutDir)
	if err := genBadTemplate.Generate(); err == nil {
		t.Error("expected error due to missing index.html template file")
	}
}
