package stats

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"

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
	router.Use(middleware.Timeout(statsServerTimeout))
	router.Use(middleware.StripSlashes)
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Use(middleware.Compress(6))

	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(statsContainer.GetStatsJSON()); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warn("Cannot return JSON to client")
		}
	})

	srv := &http.Server{
		Addr:    net.JoinHostPort(conf.ProxyAPIIP, strconv.Itoa(conf.ProxyAPIPort)),
		Handler: router,
	}

	go statsContainer.Collect()
	go log.Fatal(srv.ListenAndServe())
}
