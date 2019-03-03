package main

import (
  "log"
  "net"
  "path"
  "regexp"
  "time"
  "strings"
  "strconv"
  "gopkg.in/alecthomas/kingpin.v2"
  "github.com/prometheus/common/version"
  "github.com/prometheus/client_golang/prometheus"
  "gopkg.in/ini.v1"
)

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func parseReceivedData(data string) {
  for _, line := range strings.Split(data, "\n") {
    match := pattern.FindStringSubmatch(line)
    if (stringInSlice(match[2], syslogngconfig.sourceIds)) {
      if num, err := strconv.ParseFloat(match[6], 64); err == nil {
        syslogngProcessedMessages.Add(num)
        log.Print("Found ", match[2])
      }
    }
  }
}

func querySyslogNG() {
  c, err := net.Dial("unix", syslogngconfig.socket)
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

var (
  syslogngProcessedMessages = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "syslog_ng_processed_messages", Help: "Sum of all processed messages"})
  pattern = regexp.MustCompile(`(.*);(.*);(.*);(.*);(.*);(.*)`)
  config = kingpin.Flag(
    "config.syslog-ng",
    "Path to .syslog-ng.cnf file for some information. (socket location, list of source ids separated by ,)",
    ).Default(path.Join(path.Base(""), ".syslog-ng.cnf")).String()
    syslogngconfig = new(syslogNgConfig)
)

type syslogNgConfig struct {
  socket string
  sourceIds []string
}

func parseConfig(config interface{}) {
  opts := ini.LoadOptions{
    AllowBooleanKeys: true,
  }

  cfg, err := ini.LoadSources(opts, config)
  if err != nil {
    log.Fatal("Failed to read config")
    return
  }
  syslogngconfig.socket = "/tmp/go.sock"
  if (cfg.Section("").HasKey("socket")) {
    syslogngconfig.socket = cfg.Section("").Key("socket").String()
  }
  syslogngconfig.sourceIds = append(syslogngconfig.sourceIds, "d_splunk")
  if (cfg.Section("").HasKey("sourceids")) {
    syslogngconfig.sourceIds = strings.Split(cfg.Section("").Key("sourceids").String(), ",")
  }
}

func init() {
  kingpin.Version(version.Print("syslog_ng_exporter"))
  kingpin.HelpFlag.Short('h')
  kingpin.Parse()

  parseConfig(*config)
  collectMetrics()
  prometheus.MustRegister(syslogngProcessedMessages)
  prometheus.MustRegister(temps)
}
