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
	dbHostname    = utils.StringFromEnv(varPrefix+"DB_HOSTNAME", "127.0.0.1")
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
			AllowedHeaders: []string{
				api.SessionHeader,
				"content-type",
			},
			Debug: true,
		}).Handler(mux)
	log.Fatal(http.ListenAndServe(":"+appPort, handler))
}

func setupRouter(mux *http.ServeMux) {
	mux.HandleFunc(apiPrefix+"/healthcheck", api.Healthscheck)
	mux.HandleFunc(apiPrefix+"/api/{id}/envelope/", api.Envelope)
	mux.HandleFunc(apiPrefix+"/login", api.Login)

	setupUserEndpoints(mux)
	setupEntityEndpoints(mux, &models.Organization{})
	setupEntityEndpoints(mux, &models.Team{})
	setupEntityEndpoints(mux, &models.Project{})
	setupEntityEndpoints(mux, &models.Role{})

	mux.HandleFunc(fmt.Sprintf("GET %s/user/current", apiPrefix), api.UserCurrent)
	mux.HandleFunc(fmt.Sprintf("GET %s/teams", apiPrefix), crud.ReadEntities(&models.Team{}, false))
	mux.HandleFunc(fmt.Sprintf("GET %s/organizations", apiPrefix), crud.ReadEntities(&models.Organization{}, false))
	mux.HandleFunc(fmt.Sprintf("GET %s/projects", apiPrefix), crud.ReadEntities(&models.Project{}, false))
	mux.HandleFunc(fmt.Sprintf("GET %s/roles", apiPrefix), crud.ReadEntities(&models.Role{}, false))
	mux.HandleFunc(fmt.Sprintf("GET %s/envelopes", apiPrefix), crud.ReadEntities(&models.EnvelopeEventCommon{}, true))
	mux.HandleFunc(fmt.Sprintf("%s %s/envelopes/count", http.MethodGet, apiPrefix), crud.CountEntities(&models.EnvelopeEventCommon{}, true))
	mux.HandleFunc(fmt.Sprintf("GET %s/issues-aggregated", apiPrefix), api.IssuesAggregated)
	mux.HandleFunc(fmt.Sprintf("GET %s/issues-aggregated/count", apiPrefix), api.IssuesAggregatedCount)
	mux.HandleFunc(fmt.Sprintf("GET %s/issues/stats", apiPrefix), api.IssuesStats)
	mux.HandleFunc(fmt.Sprintf("GET %s/issue/{id}", apiPrefix), crud.Read(&models.EnvelopeEventCommon{}))
	mux.HandleFunc(fmt.Sprintf("GET %s/issue/{id}/events", apiPrefix), api.IssueEvents)
	mux.HandleFunc(fmt.Sprintf("GET %s/issue/{id}/events/count", apiPrefix), api.IssueEventsCount)
	mux.HandleFunc(fmt.Sprintf("GET %s/issue/{id}/events/stats", apiPrefix), api.IssueEventsStats)
}

func setupUserEndpoints(mux *http.ServeMux) {
	userModel := &models.User{}
	mux.HandleFunc(fmt.Sprintf("%s %s/user", http.MethodPut, apiPrefix), crud.UserUpdate(userModel))
	mux.HandleFunc(fmt.Sprintf("%s %s/user", http.MethodPost, apiPrefix), crud.UserCreate(userModel))
	mux.HandleFunc(fmt.Sprintf("%s %s/user/{id}", http.MethodGet, apiPrefix), crud.UserRead(userModel))
	mux.HandleFunc(fmt.Sprintf("%s %s/users", http.MethodGet, apiPrefix), crud.UserReadEntities(userModel, false))
	mux.HandleFunc(fmt.Sprintf("%s %s/users/count", http.MethodGet, apiPrefix), crud.UserCountEntities(userModel, false))
	mux.HandleFunc(fmt.Sprintf("%s %s/user/{id}", http.MethodDelete, apiPrefix), crud.UserDelete(userModel))
}

func setupEntityEndpoints(mux *http.ServeMux, m models.Model) {
	mux.HandleFunc(fmt.Sprintf("%s %s/%s", http.MethodPut, apiPrefix, m.GetName()), crud.Update(m))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s", http.MethodPost, apiPrefix, m.GetName()), crud.Create(m))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s/{id}", http.MethodGet, apiPrefix, m.GetName()), crud.Read(m))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s/count", http.MethodGet, apiPrefix, m.GetName()+"s"), crud.CountEntities(m, false))
	mux.HandleFunc(fmt.Sprintf("%s %s/%s/{id}", http.MethodDelete, apiPrefix, m.GetName()), crud.Delete(m))
}
