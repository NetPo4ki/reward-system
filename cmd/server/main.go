package main

import (
	"log"
	"net/http"

	"github.com/NetPo4ki/reward-system/internal/config"
	"github.com/NetPo4ki/reward-system/internal/server"
)

func main() {
	cfg := config.FromEnv()
	r := server.NewRouter(cfg)
	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
