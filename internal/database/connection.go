package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgresConnection *gorm.DB
)

func createDsn(host string, port int, username string, password string) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable TimeZone=Asia/Shanghai",
		host,
		username,
		password,
		port,
	)
}

func CreateDatabaseConnection(host string, port int, username string, password string) error {
	dsn := createDsn(host, port, username, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	postgresConnection = db
	return nil
}

func GetDatabaseConnection() *gorm.DB {
	return postgresConnection
}

func GetDatabaseVersion() string {
	var version string
	postgresConnection.Raw("select version()").Scan(&version)
	return version
}
