package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func getProcessed(lines []string) int {
	return getGivenTypeOfMessages("processed", lines)
}

func getDropped(lines []string) int {
	return getGivenTypeOfMessages("dropped", lines)
}

func getQueued(lines []string) int {
	return getGivenTypeOfMessages("queued", lines)
}

func getGivenTypeOfMessages(message_type string, lines []string) int {
	message_type_regexp, _ := regexp.Compile(fmt.Sprintf(`%v;(\d+)$`, message_type))
	number_of_messages := 0
	for line := range lines {
		line := lines[line]

		if message_type_regexp.MatchString(line) {
			current_value := strings.Trim(message_type_regexp.FindStringSubmatch(line)[1], "  ")
			if message, err := strconv.Atoi(current_value); err == nil {
				number_of_messages = number_of_messages + message
			}
		}
	}
	return number_of_messages
}

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

	number_of_processed := getProcessed(stat)
	number_of_dropped := getDropped(stat)
	number_of_queued := getQueued(stat)

	fmt.Printf("Number of processed: %v\n", number_of_processed)
	fmt.Printf("Number of dropped: %v\n", number_of_dropped)
	fmt.Printf("Number of queued: %v\n", number_of_queued)
}

func main() {
	dial()
}
