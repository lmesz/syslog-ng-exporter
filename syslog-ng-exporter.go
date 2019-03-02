package main

import (
  "log"
	"net"
  "time"
	"github.com/prometheus/client_golang/prometheus"
)

func querySyslogNG() {
	c, err := net.Dial("unix", "/tmp/go.sock")
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()

  msg := "stats"
  _, err := c.Write([]byte(msg))
  if err != nil {
	  log.Fatal("Write error:", err)
  }

	buf := make([]byte, 1024)

  n, err := c.Read(buf[:])
  if err != nil {
    return
  }
  println("Client got:", string(buf[0:n]))
}

func collectMetrics() {
  go func() {
    for {
      syslogngProcessedMessages.Add(10)
      time.Sleep(2 * time.Second)
    }
  }()
}

//Define the metrics we wish to expose
var syslogngProcessedMessages = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "syslog_ng_processed_messages", Help: "Sum of all processed messages"})

func init() {
  collectMetrics()

	//Register metrics with prometheus
	prometheus.MustRegister(syslogngProcessedMessages)
}
