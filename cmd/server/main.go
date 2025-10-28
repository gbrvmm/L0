package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gbrvmm/L0/internal/cache"
	"github.com/gbrvmm/L0/internal/config"
	"github.com/gbrvmm/L0/internal/db"
	"github.com/gbrvmm/L0/internal/stan"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo, err := db.New(ctx, cfg.PGConnString())
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("db migrate error: %v", err)
	}

	c := cache.New()

	// восстанавливаем кэш из бд
	orders, err := repo.LoadAllOrders(ctx)
	if err != nil {
		log.Fatalf("load cache from db error: %v", err)
	}
	c.SetMany(orders)
	log.Printf("cache primed with %d orders", c.Size())

	// подписка на NATS Streaming
	_, err = stan.Start(ctx, cfg, c, repo)
	if err != nil {
		log.Fatalf("stan subscribe error: %v", err)
	}

	// HTTP
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/api/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/orders/"), "/")
		id := strings.TrimSpace(parts[0])
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing id"))
			return
		}
		if v, ok := c.Get(id); ok {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			_ = enc.Encode(v)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	// статика (UI)
	http.Handle("/", http.FileServer(http.Dir("web/static")))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("http server on %s", cfg.HTTPAddr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	log.Println("graceful shutdown complete")
}
