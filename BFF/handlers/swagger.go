package handlers

import(
	"net/http"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
)

func SwaggerHandler() *mux.Router{
	pr := mux.NewRouter()

	pr.Handle("/swagger.json", http.FileServer(http.Dir("./")))
	opts := middleware.SwaggerUIOpts{SpecURL: "swagger.json"}
	sh := middleware.SwaggerUI(opts, nil)
	pr.Handle("/docs", sh)

	return pr

}