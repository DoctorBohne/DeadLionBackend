package main

import (
	"log"
	"os"

	"github.com/DoctorBohne/DeadLionBackend/internal/auth"
	"github.com/DoctorBohne/DeadLionBackend/internal/http/middleware"
)

func main() {
	issuer := mustEnv("ISS")
	jwksURL := issuer + "/protocol/openid-connect/certs"

	verifier, err := auth.NewVerifier(issuer, jwksURL)
	if err != nil {
		log.Fatal(err)
	}

	authMW := middleware.AuthMiddleware(verifier)

}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Missing required environment variable: %s", k)
	}
	return v
}
