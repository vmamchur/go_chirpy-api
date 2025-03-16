package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/vmamchur/go_chirpy-api/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFROM must be set")
	}
	jwtSecret := os.Getenv("JWT_TOKEN")
	if jwtSecret == "" {
		log.Fatal("JWT_TOKEN must be set")
	}

	dbConnection, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConnection)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", http.StripPrefix("/app", fsHandler))

	mux.Handle("GET /api/healthz", http.HandlerFunc(handlerReadiness))

	mux.Handle("POST /api/polka/webhooks", http.HandlerFunc(apiCfg.handlerPolkaWebhook))

	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.handlerUsersCreate))
	mux.Handle("PUT /api/users", http.HandlerFunc(apiCfg.handlerUsersUpdate))

	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.handlerLogin))
	mux.Handle("POST /api/refresh", http.HandlerFunc(apiCfg.handlerRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(apiCfg.handlerRevoke))

	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handlerChirpsCreate))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.handlerChirpsRetrieve))
	mux.Handle("GET /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.handlerChirpsGet))
	mux.Handle("DELETE /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.handlerChirpsDelete))

	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.handlerReset))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
