package main

import (
  "net/http"

  log "github.com/Sirupsen/logrus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
  //This section will start the HTTP server and expose
  //any metrics on the /metrics endpoint.
  http.Handle("/metrics", promhttp.Handler())
  log.Info("Beginning to serve on port :8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
