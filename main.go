package main

import (
	"log/slog"
	"time"
)

func main() {
	for {
		slog.Info("Running kubespiffed...")
		time.Sleep(5 * time.Second)
	}
}
