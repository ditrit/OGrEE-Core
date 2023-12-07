package e2e

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"p3/app"
	"p3/models"
	"p3/router"
	_ "p3/test/integration"

	"github.com/elliotchance/pie/v2"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var appRouter *mux.Router
var AdminId primitive.ObjectID
var AdminToken string

func init() {
	appRouter = router.Router(app.JwtAuthentication)
	createAdminAccount()
}

func createAdminAccount() {
	// Create admin account
	admin := models.Account{}
	admin.Email = "admin@admin.com"
	admin.Password = "admin123"
	admin.Roles = map[string]models.Role{"*": "manager"}

	newAcc, err := admin.Create(map[string]models.Role{"*": "manager"})
	if err != nil {
		log.Fatalln("Error while creating admin account:", err.Error())
	}

	if newAcc != nil {
		AdminId = newAcc.ID
		AdminToken = newAcc.Token
	}
}

func MakeRequest(method, url string, requestBody []byte) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	request.Header.Set("Authorization", "Bearer "+AdminToken)
	appRouter.ServeHTTP(recorder, request)

	return recorder
}

func GetObjects(queryParams string) (*httptest.ResponseRecorder, []map[string]any) {
	response := MakeRequest(http.MethodGet, router.GenericObjectsURL+"?"+queryParams, nil)

	var objects []map[string]any
	if response.Code == http.StatusOK {
		var responseBody map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &responseBody)
		objects = pie.Map(responseBody["data"].([]any), func(objectAny any) map[string]any {
			return objectAny.(map[string]any)
		})
	}

	return response, objects
}
