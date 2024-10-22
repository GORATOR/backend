package crud

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/service"
	"github.com/GORATOR/backend/internal/utils"
)

func Create[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, ok := before(w, r, entity, nil)
		if !ok {
			return
		}

		if !service.HasUserAccessToByUserId(uint(userId), models.ActionCreate, entity) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}

		body, err := utils.GetBodyBytes(r)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}
		var entityObject V
		err = json.Unmarshal(body, &entityObject)
		if err != nil {
			fmt.Print("create json.Unmarshal error", err)
			utils.HttpReturnBadRequest(w)
			return
		}

		//filter fields

		insertResult := database.GetDatabaseConnection().Create(&entityObject)
		if insertResult.Error != nil {
			fmt.Print("create db insert error", insertResult.Error)
			utils.HttpReturnBadRequest(w)
			return
		}
		utils.HttpReturnJson(w, entityObject)
	}
}
