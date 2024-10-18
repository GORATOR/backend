package crud

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GORATOR/backend/internal/database"
	"github.com/GORATOR/backend/internal/models"
	"github.com/GORATOR/backend/internal/utils"
)

func Create[V models.Entity](entity string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !before(w, r, entity, nil) {
			return
		}

		/*if !service.HasUserAccessToByUserId(id, models.ActionCreate, entityInterface) {
			http.Error(w, fmt.Sprintf("Forbidden action \"%s\"", models.ActionRead), http.StatusForbidden)
			return
		}*/

		body, err := utils.GetBodyBytes(r)
		if err != nil {
			utils.HttpReturnBadRequest(w)
			return
		}
		var jsonObject V
		err = json.Unmarshal(body, &jsonObject)
		if err != nil {
			fmt.Print("create json.Unmarshal error", err)
			utils.HttpReturnBadRequest(w)
			return
		}

		//filter fields

		insertResult := database.GetDatabaseConnection().Create(&jsonObject)
		if insertResult.Error != nil {
			fmt.Print("create db insert error", insertResult.Error)
			utils.HttpReturnBadRequest(w)
			return
		}
		utils.HttpReturnJson(w, jsonObject)
	}
}
