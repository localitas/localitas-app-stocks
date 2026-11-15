package stocks

import (
	"net/http"

	client "github.com/localitas/localitas-go"
)

var cronSpec = client.CronSpec{
	Jobs: []client.CronJob{
		{
			ID:          "cron:stocks:refresh-quotes",
			Path:        "/api/refresh",
			Method:      "POST",
			Schedule:    "*/5 * * * *",
			Description: "Refreshes stock quotes for all portfolio holdings",
			Timeout:     "60s",
			Retry:       &client.CronRetry{MaxAttempts: 1},
		},
	},
}

func HandleCron(w http.ResponseWriter, r *http.Request) {
	cronSpec.ServeHTTP(w, r)
}
