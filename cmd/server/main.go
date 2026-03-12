package main

import (
	"log"
	"os"

	"github.com/DoctorBohne/DeadLionBackend/internal/auth"
	"github.com/DoctorBohne/DeadLionBackend/internal/db"
	"github.com/DoctorBohne/DeadLionBackend/internal/http"
	"github.com/DoctorBohne/DeadLionBackend/internal/http/middleware"
)

func main() {
	cfg := db.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Pass:     os.Getenv("DB_PASS"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		TimeZone: "Europe/Berlin",
	}

	gdb, err := db.OpenPostgres(cfg)
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	autoMigrate := os.Getenv("DB_AUTOMIGRATE")
	if autoMigrate == "" {
		autoMigrate = "true"
	}
	if err = db.Migrate(gdb); err != nil {
		log.Fatalf("DB migrate failed: %v", err)
	}
	issuer := mustEnv("ISS")
	jwksURL := issuer + "/protocol/openid-connect/certs"

	verifier, err := auth.NewVerifier(jwksURL, issuer)
	if err != nil {
		log.Fatal(err)
	}

	authMW := middleware.AuthMiddleware(verifier)

	deps := http.Deps{
		DB:             gdb,
		AuthMiddleWare: authMW,
	}
	r := http.NewRouter(deps)
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Missing required environment variable: %s", k)
	}
	return v
}
