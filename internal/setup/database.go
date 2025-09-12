package setup

import (
	"errors"
	"fmt"
	"log"

	"github.com/GORATOR/backend/internal/config"
	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type RoleRulesCallback func(db *gorm.DB, role models.Role, tableName string)

func tryCreateRecord(db *gorm.DB, value interface{}) {
	result := db.Create(value)
	if result.Error == nil {
		return
	}
	var err *pgconn.PgError
	if errors.As(result.Error, &err) && err.Code != "23505" {
		panic(err)
	}
}

func setupRole(name string, db *gorm.DB, user *models.User, cb RoleRulesCallback) {
	role := models.Role{
		Name:  name,
		Users: []*models.User{user},
	}
	tryCreateRecord(db, &role)
	for _, entity := range []string{
		models.TeamModelName,
		models.UserModelName,
		models.OrganizationModelName,
		models.ProjectModelName,
		models.EnvelopeEventCommonModelName,
	} {
		cb(db, role, entity+"s")
	}
}

func setupRoles(db *gorm.DB, userAdmin *models.User, userViewer *models.User) {
	setupRole("admin", db, userAdmin, func(db *gorm.DB, role models.Role, entity string) {
		for _, action := range []models.RuleAction{
			models.ActionCreate,
			models.ActionDelete,
			models.ActionRead,
			models.ActionUpdate,
		} {
			rule := models.Rule{
				Action:  action,
				Table:   entity,
				Allowed: true,
			}
			rule.Role = role
			rule.RoleID = role.ID
			db.Create(&rule)
		}
	})

	setupRole("viewer", db, userViewer, func(db *gorm.DB, role models.Role, entity string) {
		rule := models.Rule{
			Action:  models.ActionRead,
			Table:   entity,
			Allowed: true,
		}
		rule.Role = role
		rule.RoleID = role.ID
		db.Create(&rule)
		for _, action := range []models.RuleAction{
			models.ActionCreate,
			models.ActionDelete,
			models.ActionUpdate,
		} {
			rule := models.Rule{
				Action:  action,
				Table:   entity,
				Allowed: false,
			}
			rule.Role = role
			rule.RoleID = role.ID
			db.Create(&rule)
		}

	})

	setupRole("viewerImplicit", db, userViewer, func(db *gorm.DB, role models.Role, entity string) {
		rule := models.Rule{
			Action:  models.ActionRead,
			Table:   entity,
			Allowed: true,
		}
		rule.Role = role
		rule.RoleID = role.ID
		db.Create(&rule)
	})

}

func tryCreateUser(db *gorm.DB, email string, username string, team models.Team, org models.Organization) *models.User {
	user := models.User{
		Username: username,
		Email:    email,
		Active:   true,
	}
	result := db.Model(&models.User{}).Where("username = ? and email = ?", username, email).Find(&user)
	if result.Error != nil {
		fmt.Print(result.Error)
		return nil
	}
	if result.RowsAffected == 1 {
		return &user
	}
	user.Teams = []*models.Team{&team}
	user.Organizations = []*models.Organization{&org}

	salt := utils.StringFromEnv("GORATOR_SALT", "")
	if salt == "" {
		log.Printf("Empty env GORATOR_SALT")
	}

	user.CreateHashedPassword(utils.StringFromEnv("GORATOR_DEBUG_USERS_PASSWORD", "pwd"), salt)
	tryCreateRecord(db, &user)
	return &user
}

func SetupDatabase() {
	db := database.GetDatabaseConnection()

	ruleActionResult := db.Debug().Exec(`
    DO $$ BEGIN
        CREATE TYPE public.rule_action AS ENUM ('create','read','update', 'delete');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END $$;`)
	if ruleActionResult.Error != nil {
		panic(ruleActionResult.Error)
	}

	err := db.AutoMigrate(
		&models.EventCommonSdk{},
		&models.EnvelopeEventCommon{},
		&models.EnvelopeEventExtra{},
		&models.EnvelopeTag{},
		&models.User{},
		&models.Team{},
		&models.Organization{},
		&models.Role{},
		&models.Rule{},
		&models.Project{},
		&models.ClientReport{},
		&models.ClientReportDiscardedEvent{},
	)
	if err != nil {
		panic(err)
	}
	uniqueIndexResult := db.Raw("CREATE UNIQUE INDEX unique_name_version ON event_common_sdks (name, version)")
	if uniqueIndexResult.Error != nil {
		panic(uniqueIndexResult.Error)
	}

	uniqueIndexResult = db.Raw("CREATE UNIQUE INDEX unique_ee_tag ON envelope_tags (name, value)")
	if uniqueIndexResult.Error != nil {
		panic(uniqueIndexResult.Error)
	}

	envelopeKeyIndexResult := db.Raw("CREATE UNIQUE INDEX unique_envelope_key ON projects (envelope_key)")
	if envelopeKeyIndexResult.Error != nil {
		panic(uniqueIndexResult.Error)
	}

	tryCreateRecord(db, &models.UndefinedSdk)

	if config.IsDebug() {
		org := models.Organization{
			Name:   "Test Organization",
			Active: true,
		}
		tryCreateRecord(db, &org)

		team := models.Team{
			Organizations: []*models.Organization{&org},
			Name:          "Test Team",
			Active:        true,
		}
		tryCreateRecord(db, &team)

		userAdmin := tryCreateUser(db, "user1@email.com", "userAdmin", team, org)
		userViewer := tryCreateUser(db, "user2@email.com", "userViewer", team, org)
		if userAdmin != nil && userViewer != nil {
			setupRoles(db, userAdmin, userViewer)
		}
	}
	fmt.Println("database setup is done")
}
