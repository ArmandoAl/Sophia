package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/armandoalvarado/sofia-backend/internal/app"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	container, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}
	defer container.Close()

	log.Printf("server starting addr=%s env=%s persistence=%s firestore=%s", cfg.Addr(), cfg.Env, cfg.PersistenceDriver, container.FirestoreStatus)
	if err := container.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}
