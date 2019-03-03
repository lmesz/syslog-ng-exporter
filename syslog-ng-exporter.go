package main

import (
  "log"
  "net"
  "path"
  "regexp"
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

func parseReceivedData(data string) map[string]map[string]float64 {
  var result = map[string]map[string]float64{}
  for _, sid := range syslogngconfig.sourceIds {
    result[sid] = map[string]float64{}
  }
  for _, line := range strings.Split(data, "\n") {
    match := pattern.FindStringSubmatch(line)
    if (stringInSlice(match[2], syslogngconfig.sourceIds)) {
      if num, err := strconv.ParseFloat(match[6], 64); err == nil {
        result[match[2]][match[5]] = num
      }
    }
  }
  return result
}

func querySyslogNG() map[string]map[string]float64 {
  var result = make(map[string]map[string]float64)
  c, err := net.Dial("unix", syslogngconfig.socket)
  if err != nil {
    log.Print("Dial error", err)
    return result
  }
  defer c.Close()
  msg := "stats"
  _, err = c.Write([]byte(msg))
  if err != nil {
    log.Print("Write error:", err)
    return result
  }
  buf := make([]byte, 10240)
  _, err = c.Read(buf[:])
  if err != nil {
    return result
  }
  result = parseReceivedData(string(buf))
  return result
}

var (
  pattern = regexp.MustCompile(`(.*);(.*);(.*);(.*);(.*);(.*)`)
  config = kingpin.Flag(
    "config.syslog-ng",
    "Path to .syslog-ng.cnf file for some information. (socket location, list of source ids separated by ,)",
    ).Default(path.Join(path.Base(""), ".syslog-ng.cnf")).String()
    syslogngconfig = new(syslogNgConfig)
)

type syslogNgCollector struct {
	syslogNgMetric *prometheus.Desc
}

func newSyslogNgCollector() *syslogNgCollector {
	return &syslogNgCollector{
		syslogNgMetric: prometheus.NewDesc("syslog_ng_metric",
			"Shows syslo-ng destination statistics based on given config",
			[]string{"source_id", "type"}, nil,
		),
	}
}

func (collector *syslogNgCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.syslogNgMetric
}

func (collector *syslogNgCollector) Collect(ch chan<- prometheus.Metric) {
	var currentData = querySyslogNG()
	for source := range currentData {
		for stateType := range currentData[source] {
			ch <- prometheus.MustNewConstMetric(collector.syslogNgMetric, prometheus.CounterValue, currentData[source][stateType], source, stateType)
		}
	}
}


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
  prometheus.MustRegister(newSyslogNgCollector())
}
