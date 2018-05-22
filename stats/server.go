package stats

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

const (
	statsServerTimeout      = 2 * time.Second
	statsConcurrentRequests = 10
)

func RunStats(statsContainer *Stats, conf *config.Config) {
	router := chi.NewRouter()

	router.Use(middleware.GetHead)
	router.Use(middleware.Throttle(statsConcurrentRequests))
	router.Use(middleware.Timeout(time.Second * 2))
	router.Use(middleware.StripSlashes)
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Use(middleware.Compress(6))

	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(statsContainer.GetStatsJSON())
	})

	srv := &http.Server{
		Addr:    net.JoinHostPort(conf.ProxyAPIIP, strconv.Itoa(conf.ProxyAPIPort)),
		Handler: router,
	}

	go statsContainer.Collect()
	go srv.ListenAndServe()
}
