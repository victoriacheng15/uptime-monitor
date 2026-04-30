package main

import (
	"fmt"

	"uptime-monitor/internal/models"
	"uptime-monitor/internal/monitor"
	"uptime-monitor/internal/storage"
)

func main() {
	fmt.Println("hello from uptime-monitor lambda backend")
	fmt.Println(models.Hello())
	fmt.Println(monitor.Hello())
	fmt.Println(storage.Hello())
}
