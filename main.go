package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/GORATOR/backend/internal/api"
	"github.com/GORATOR/backend/internal/config"
	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
	"github.com/rs/cors"
	"gorm.io/gorm"
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
		setupDatabase()
		return
	default:
		//todo: print help
	}
}

func setupDatabase() {
	db := database.GetDatabaseConnection()
	err := db.AutoMigrate(
		&models.EventCommonSdk{},
		&models.EnvelopeEventCommon{},
		&models.EnvelopeEventExtra{},
	)
	if err != nil {
		panic(err)
	}
	uniqueIndexResult := db.Raw("CREATE UNIQUE INDEX unique_name_version ON event_common_sdks (name, version)")
	if uniqueIndexResult.Error != nil {
		panic(uniqueIndexResult.Error)
	}
	undefinedSdkResult := db.Create(&models.UndefinedSdk)
	if undefinedSdkResult.Error != nil &&
		errors.Is(undefinedSdkResult.Error, gorm.ErrDuplicatedKey) {
		panic(undefinedSdkResult.Error)
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
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
			Debug:            true,
		}).Handler(mux)
	log.Fatal(http.ListenAndServe(":"+appPort, handler))
}

func setupRouter(mux *http.ServeMux) {
	mux.HandleFunc(apiPrefix+"/healthcheck", api.Healthscheck)
	mux.HandleFunc(apiPrefix+"/api/{id}/envelope/", api.Envelope)
}
