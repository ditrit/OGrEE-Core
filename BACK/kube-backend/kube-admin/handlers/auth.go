package handlers
import (
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"kube-admin/auth"
	"kube-admin/models"
)
// swagger:operation POST /login Authentication Authenticate
// Generates a new JWT Key for the client.
// Create a new JWT Key. This can also be used to verify credentials
// The authorize and 'Try it out' buttons don't work
// ---
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: email and password.'
//     required: true
//     format: object
//     example: '{"email": "user@test.com", "password": "secret123"}'
//
// responses:
//     '200':
//         description: Authenticated
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error
func Login(c *gin.Context) {
	var userIn models.User
	if err := c.BindJSON(&userIn); err != nil {
		println("ERROR:")
		println(err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	} else {
		// Check credentials
		if userIn.Email != "admin" ||
			bcrypt.CompareHashAndPassword([]byte(os.Getenv("ADM_PASSWORD")), []byte(userIn.Password)) != nil {
			println("Credentials error")
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
			return
		}

		println("Generate")
		// Generate token
		token, err := auth.GenerateToken(userIn.Email)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Respond
		response := make(map[string]map[string]string)
		response["account"] = make(map[string]string)
		response["account"]["Email"] = userIn.Email
		response["account"]["token"] = token
		response["account"]["isTenant"] = "true"
		c.IndentedJSON(http.StatusOK, response)
	}
}