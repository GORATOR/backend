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

// bindIfMissing reports whether the join table does not link owner and
// related yet. Join tables have no unique constraints, so appending blindly
// would create duplicate rows on re-run.
func bindIfMissing(db *gorm.DB, joinTable string, ownerColumn string, ownerID uint, relatedColumn string, relatedID uint) bool {
	var bound int64
	result := db.Table(joinTable).
		Where(fmt.Sprintf("%s = ? and %s = ?", ownerColumn, relatedColumn), ownerID, relatedID).
		Count(&bound)
	if result.Error != nil {
		panic(result.Error)
	}
	return bound == 0
}

func bindTeamOrganization(db *gorm.DB, team *models.Team, org *models.Organization) {
	if !bindIfMissing(db, "org_teams", "team_id", team.ID, "organization_id", org.ID) {
		return
	}
	if err := db.Model(team).Association("Organizations").Append(org); err != nil {
		panic(err)
	}
}

func bindUserTeamAndOrganization(db *gorm.DB, user *models.User, team *models.Team, org *models.Organization) {
	if bindIfMissing(db, "team_users", "user_id", user.ID, "team_id", team.ID) {
		if err := db.Model(user).Association("Teams").Append(team); err != nil {
			panic(err)
		}
	}
	if bindIfMissing(db, "org_users", "user_id", user.ID, "organization_id", org.ID) {
		if err := db.Model(user).Association("Organizations").Append(org); err != nil {
			panic(err)
		}
	}
}

func bindRoleUser(db *gorm.DB, role *models.Role, user *models.User) {
	if !bindIfMissing(db, "role_users", "role_id", role.ID, "user_id", user.ID) {
		return
	}
	if err := db.Model(role).Association("Users").Append(user); err != nil {
		panic(err)
	}
}

func setupRole(name string, db *gorm.DB, user *models.User, cb RoleRulesCallback) {
	role := models.Role{
		Name:   name,
		Active: true,
	}
	result := db.Where("name = ?", name).First(&role)
	switch {
	case result.Error == nil:
		bindRoleUser(db, &role, user)
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		role.Users = []*models.User{user}
		tryCreateRecord(db, &role)
	default:
		panic(result.Error)
	}

	// Recreate rules deterministically, same approach as Role.UpdateModel.
	// Unscoped: rules are read by raw SQL (service.HasUserAccessToByUserId)
	// without a deleted_at filter, so old rules must be removed for real.
	db.Unscoped().Where("role_id = ?", role.ID).Delete(&models.Rule{})
	for _, entity := range []string{
		models.TeamModelName,
		models.UserModelName,
		models.OrganizationModelName,
		models.ProjectModelName,
		models.EnvelopeEventCommonModelName,
		models.TransactionModelName,
		models.RoleModelName,
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
			tryCreateRecord(db, &rule)
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
		tryCreateRecord(db, &rule)
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
			tryCreateRecord(db, &rule)
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
		tryCreateRecord(db, &rule)
	})

}

func tryFindOrCreateOrganization(db *gorm.DB, name string) *models.Organization {
	var org models.Organization
	result := db.Where("name = ?", name).First(&org)
	if result.Error == nil {
		return &org
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(result.Error)
	}
	org = models.Organization{
		Name:   name,
		Active: true,
	}
	tryCreateRecord(db, &org)
	return &org
}

func tryFindOrCreateTeam(db *gorm.DB, name string, org *models.Organization) *models.Team {
	var team models.Team
	result := db.Where("name = ?", name).First(&team)
	if result.Error == nil {
		bindTeamOrganization(db, &team, org)
		return &team
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(result.Error)
	}
	team = models.Team{
		Organizations: []*models.Organization{org},
		Name:          name,
		Active:        true,
	}
	tryCreateRecord(db, &team)
	return &team
}

func tryCreateUser(db *gorm.DB, email string, username string, team *models.Team, org *models.Organization) *models.User {
	user := models.User{
		Username: username,
		Email:    email,
		Active:   true,
	}
	result := db.Where("username = ? and email = ?", username, email).First(&user)
	if result.Error == nil {
		bindUserTeamAndOrganization(db, &user, team, org)
		return &user
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(result.Error)
	}
	user.Teams = []*models.Team{team}
	user.Organizations = []*models.Organization{org}

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
		&models.Transaction{},
		&models.Span{},
	)
	if err != nil {
		panic(err)
	}

	if err := DeduplicateData(db); err != nil {
		panic(err)
	}

	uniqueIndexes := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_name_version ON event_common_sdks (name, version)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_ee_tag ON envelope_tags (name, value)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS unique_envelope_key ON projects (envelope_key)`,
	}
	for _, idx := range uniqueIndexes {
		if err := db.Exec(idx).Error; err != nil {
			panic(fmt.Errorf("failed to create unique index (%s): %w", idx, err))
		}
	}

	var undefinedSdk models.EventCommonSdk
	undefinedSdkResult := db.Where(
		"name = ? and version = ?",
		models.UndefinedSdk.Name,
		models.UndefinedSdk.Version,
	).First(&undefinedSdk)
	if errors.Is(undefinedSdkResult.Error, gorm.ErrRecordNotFound) {
		tryCreateRecord(db, &models.UndefinedSdk)
	} else if undefinedSdkResult.Error != nil {
		panic(undefinedSdkResult.Error)
	}

	// Run migration for exception fields
	if err := MigrateExceptionFields(db); err != nil {
		fmt.Printf("Warning: Exception fields migration failed: %v\n", err)
	}

	if err := MigrateExtraField(db); err != nil {
		fmt.Printf("Warning: Extra field migration failed: %v\n", err)
	}

	if err := MigrateStacktraceHash(db); err != nil {
		fmt.Printf("Warning: Stacktrace hash migration failed: %v\n", err)
	}

	if err := MigrateTraces(db); err != nil {
		fmt.Printf("Warning: Traces migration failed: %v\n", err)
	}

	if config.IsDebug() {
		org := tryFindOrCreateOrganization(db, "Test Organization")
		team := tryFindOrCreateTeam(db, "Test Team", org)
		userAdmin := tryCreateUser(db, "user1@email.com", "userAdmin", team, org)
		userViewer := tryCreateUser(db, "user2@email.com", "userViewer", team, org)
		setupRoles(db, userAdmin, userViewer)
	}
	fmt.Println("database setup is done")
}
