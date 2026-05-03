package main

import (
	"log"
	"os"
	"uptime-monitor/internal/web"
)

func main() {
	log.Println("Starting Uptime Monitor SSG...")

	apiBaseURL := os.Getenv("API_BASE_URL")
	if apiBaseURL == "" {
		log.Println("Warning: API_BASE_URL environment variable not set. Frontend will use placeholder.")
	}

	gen := web.NewGenerator(
		"internal/web/templates",
		"internal/web/templates/content",
		"dist",
	)
	gen.APIBaseURL = apiBaseURL

	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate site: %v", err)
	}

	log.Println("Site generated successfully in dist/")
}
