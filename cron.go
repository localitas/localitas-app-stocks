package stocks

import (
	"encoding/json"
	"net/http"
)

func HandleCron(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"id":          "cron:stocks:refresh-quotes",
				"path":        "/api/refresh",
				"method":      "POST",
				"schedule":    "*/5 * * * *",
				"description": "Refreshes stock quotes for all portfolio holdings",
				"timeout":     "60s",
				"retry": map[string]interface{}{
					"max_attempts": 1,
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}
