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

	"github.com/scrapinghub/crawlera-headless-proxy/config"
)

const (
	statsServerTimeout      = 2 * time.Second
	statsConcurrentRequests = 10
)

// RunStats runs statistics collector and API service.
func RunStats(statsContainer *Stats, conf *config.Config) {
	router := chi.NewRouter()

	router.Use(middleware.GetHead)
	router.Use(middleware.Throttle(statsConcurrentRequests))
	router.Use(middleware.Timeout(statsServerTimeout))
	router.Use(middleware.StripSlashes)
	router.Use(middleware.SetHeader("Content-Type", "application/json"))
	router.Use(middleware.Compress(6))

	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")

		statsContainer.statsLock.Lock()
		defer statsContainer.statsLock.Unlock()

		if err := encoder.Encode(statsContainer); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warn("Cannot return JSON to client")
		}
	})

	srv := &http.Server{
		Addr:    net.JoinHostPort(conf.ProxyAPIIP, strconv.Itoa(conf.ProxyAPIPort)),
		Handler: router,
	}

	log.Fatal(srv.ListenAndServe())
}
