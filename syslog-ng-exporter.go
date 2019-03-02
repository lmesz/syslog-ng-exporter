package main

import (
  "log"
	"net"
  "os"
  "path"
  "regexp"
  "time"
  "strings"
  "strconv"
  "gopkg.in/alecthomas/kingpin.v2"
	"github.com/prometheus/client_golang/prometheus"
  "gopkg.in/ini.v1"
)

func parseReceivedData(data string) {
  for _, line := range strings.Split(data, "\n") {
    match := pattern.FindStringSubmatch(line)
    if (match[5] == "processed") {
      if num, err := strconv.ParseFloat(match[6], 64); err == nil {
        syslogngProcessedMessages.Add(num)
      }
    }
  }
}

func querySyslogNG() {
	c, err := net.Dial("unix", "/tmp/go.sock")
	if err != nil {
		log.Print("Dial error", err)
    return
	}
	defer c.Close()

  msg := "stats"
  _, err = c.Write([]byte(msg))
  if err != nil {
	  log.Print("Write error:", err)
    return
  }

	buf := make([]byte, 10240)

  _, err = c.Read(buf[:])
  if err != nil {
    return
  }
  parseReceivedData(string(buf))
}

func collectMetrics() {
  go func() {
    for {
      querySyslogNG()
      time.Sleep(2 * time.Second)
    }
  }()
}

//Define the metrics we wish to expose
var (
  syslogngProcessedMessages = prometheus.NewGauge(prometheus.GaugeOpts{
	  Name: "syslog_ng_processed_messages", Help: "Sum of all processed messages"})
  pattern = regexp.MustCompile(`(.*);(.*);(.*);(.*);(.*);(.*)`)
  config = kingpin.Flag(
    "config.syslog-ng",
    "Path to .syslog-ng.cnf file for some information. (socket location, observable elements)",
    ).Default(path.Join(os.Getenv("HOME"), ".syslog-ng.cnf")).String()
)

func parseConfig(config interface{}) {
	opts := ini.LoadOptions{
		AllowBooleanKeys: true,
	}

	_, err := ini.LoadSources(opts, config)
	if err != nil {
    log.Fatal("Failed to read config")
		return
	}
}

func init() {
  parseConfig(*config)
  collectMetrics()
	prometheus.MustRegister(syslogngProcessedMessages)
}
