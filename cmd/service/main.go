package main

import (
	"log"
	"os"
	"os/signal"
	"reviewer-service/cmd/env"
	"reviewer-service/cmd/inits"
	"syscall"
)

func main() {
	env.InitEnv()

	cfg, err := inits.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := inits.InitDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database!")

	if err := inits.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations applied successfully!")

	services := inits.InitServices(db)

	handler := inits.SetupRoutes(services)

	go func() {
		if err := inits.StartServer(cfg, handler); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
}
