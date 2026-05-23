package health

import (
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
)

func HandleGet(w http.ResponseWriter, r *http.Request) {
	api.OKResponse(w, map[string]string{"status": "ok"})
}
