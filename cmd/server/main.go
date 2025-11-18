package main

import (
	"context"
	"log"
	"net/http"

	"github.com/NetPo4ki/reward-system/internal/config"
	"github.com/NetPo4ki/reward-system/internal/db"
	"github.com/NetPo4ki/reward-system/internal/server"
)

func main() {
	cfg := config.FromEnv()
	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	r := server.NewRouter(cfg, pool)
	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
