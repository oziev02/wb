package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/oziev02/wb/internal/adapters/httpapi"
	"github.com/oziev02/wb/internal/adapters/mq/kafka"
	"github.com/oziev02/wb/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	c, err := app.NewContainer(ctx, cfg)
	if err != nil {
		log.Fatalf("container: %v", err)
	}

	mux := http.NewServeMux()
	h := httpapi.NewHandler(c.Svc)
	h.Routes(mux)
	httpapi.ServeStatic(mux, "./web")

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: mux}

	consumer := kafka.NewConsumer(kafka.Config{
		Brokers: cfg.KafkaBrokers, Topic: cfg.KafkaTopic, GroupID: cfg.KafkaGroup,
	}, c.Svc)

	go func() {
		log.Printf("HTTP listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	go func() {
		if err := consumer.Run(ctx); err != nil {
			log.Printf("kafka stopped: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}
