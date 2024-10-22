package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/GORATOR/backend/internal/api"
	"github.com/GORATOR/backend/internal/api/crud"
	"github.com/GORATOR/backend/internal/config"
	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/setup"
	"github.com/GORATOR/backend/internal/utils"
	"github.com/rs/cors"
)

var (
	varPrefix     = "GORATOR_"
	apiPrefix     = utils.StringFromEnv(varPrefix+"API_PREFIX", "")
	appPort       = utils.StringFromEnv(varPrefix+"APP_PORT", "8080")
	allowedOrigin = utils.StringFromEnv(varPrefix+"ALLOWED_ORIGIN", "http://localhost:4000")
	dbHostname    = utils.StringFromEnv(varPrefix+"DB_HOSTNAME", "localhost")
	dbPort        = utils.StringFromEnv(varPrefix+"DB_PORT", "5432")
	dbUsername    = utils.StringFromEnv(varPrefix+"DB_USERNAME", "postgres")
	dbPassword    = utils.StringFromEnv(varPrefix+"DB_PASSWORD", "postgres")
)

func printAppMode() {
	var appMode string
	if config.IsDebug() {
		appMode = "development"
	} else {
		appMode = "production"
	}
	fmt.Printf("Running application in %s mode\n", appMode)
}

func runCliMode() {
	switch os.Args[1] {
	case "-s":
		setup.SetupDatabase()
		return
	default:
		//todo: print help
	}
}

func main() {
	printAppMode()

	port, _ := strconv.Atoi(dbPort)
	err := database.CreateDatabaseConnection(dbHostname, port, dbUsername, dbPassword)
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 1 {
		runCliMode()
		return
	}

	mux := http.NewServeMux()
	setupRouter(mux)
	allowedOrigins := []string{allowedOrigin}
	handler := cors.New(
		cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: true,
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodOptions,
				http.MethodDelete,
			},
			Debug: true,
		}).Handler(mux)
	log.Fatal(http.ListenAndServe(":"+appPort, handler))
}

func setupRouter(mux *http.ServeMux) {
	mux.HandleFunc(apiPrefix+"/healthcheck", api.Healthscheck)
	mux.HandleFunc(apiPrefix+"/api/{id}/envelope/", api.Envelope)
	mux.HandleFunc(apiPrefix+"/login", api.Login)

	setupEntityEndpoints[models.User](mux, models.UserEntityName)
	setupEntityEndpoints[models.Organization](mux, models.OrganizationEntityName)
	setupEntityEndpoints[models.Team](mux, models.TeamEntityName)

	mux.HandleFunc("GET /user/current", api.UserCurrent)

}

func setupEntityEndpoints[V models.Entity](mux *http.ServeMux, entityName string) {
	mux.HandleFunc(fmt.Sprintf("%s %s/%s", http.MethodPut, apiPrefix, entityName), crud.Update[V](entityName))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s", http.MethodPost, apiPrefix, entityName), crud.Create[V](entityName))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s/{id}", http.MethodGet, apiPrefix, entityName), crud.Read[V](entityName))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s/{id}", http.MethodDelete, apiPrefix, entityName), crud.Delete[V](entityName))
}
