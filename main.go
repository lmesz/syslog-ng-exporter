package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func dial() {
	conn, err := net.Dial("unix", "/var/lib/syslog-ng/syslog-ng.ctl")
	if err != nil {
		fmt.Printf("Failed to dial: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("STATS\n")); err != nil {
		fmt.Printf("DIAL: Error: %v\n", err)
	}

	buf := make([]byte, 2048)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Printf("LISTEN: Error: %v\n", err)
		os.Exit(1)
	}
	stat := strings.Split(string(buf), "\n")

	processed, _ := regexp.Compile(`processed;(\d+)$`)
	dropped, _ := regexp.Compile(`dropped;(\d+)$`)
	queued, _ := regexp.Compile(`queued;(\d+)$`)

	number_of_processed := 0
	number_of_dropped := 0
	number_of_queued:= 0
	for line := range stat {
		line := stat[line]

		if processed.MatchString(line) {
			current_value := strings.Trim(processed.FindStringSubmatch(line)[1], "  ")
			if proceed, err := strconv.Atoi(current_value); err == nil {
				number_of_processed = number_of_processed + proceed
			}
		}

		if dropped.MatchString(line) {
			current_value := strings.Trim(dropped.FindStringSubmatch(line)[1], "  ")
			if dropped, err := strconv.Atoi(current_value); err == nil {
				number_of_dropped = number_of_dropped + dropped
			}
        }

        if queued.MatchString(line) {
			current_value := strings.Trim(queued.FindStringSubmatch(line)[1], "  ")
			if queued, err := strconv.Atoi(current_value); err == nil {
				number_of_queued = number_of_queued + queued
			}
        }
	}
    fmt.Printf("Number of processed: %v\n", number_of_processed)
    fmt.Printf("Number of dropped: %v\n", number_of_dropped)
    fmt.Printf("Number of queued: %v\n", number_of_queued)
}

func main() {
	dial()
}
