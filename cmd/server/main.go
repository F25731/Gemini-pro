package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"banana-pro-wrapper/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	server := app.NewServer(cfg)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Router(),
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 30*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 30*time.Second,
	}

	go func() {
		log.Printf("banana pro wrapper listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}
