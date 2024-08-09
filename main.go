package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GORATOR/backend/internal/api"
	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/utils"
	"github.com/rs/cors"
)

var (
	varPrefix     = "SHOVISHE"
	apiPrefix     = utils.StringFromEnv(varPrefix+"API_PREFIX", "")
	appPort       = utils.StringFromEnv(varPrefix+"APP_PORT", "8080")
	allowedOrigin = utils.StringFromEnv(varPrefix+"ALLOWED_ORIGIN", "http://localhost:4000")
)

func main() {
	fmt.Println("work in progress...")
	err := database.CreateDatabaseConnection("localhost", 5432, "postgres", "postgres")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	setupRouter(mux)
	allowedOrigins := []string{allowedOrigin}
	handler := cors.New(
		cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
			Debug:            true,
		}).Handler(mux)
	log.Fatal(http.ListenAndServe(":"+appPort, handler))
}

func setupRouter(mux *http.ServeMux) {
	mux.HandleFunc(apiPrefix+"/healthcheck", api.Healthscheck)
}
