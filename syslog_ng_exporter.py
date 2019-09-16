#!/usr/bin/python3


import argparse
import time
import logging
import socket
import sys
from prometheus_client.core import GaugeMetricFamily, REGISTRY
from prometheus_client import start_http_server


class CustomCollector(object):
    def __init__(self, socket, targets):
        self.socket = socket
        self.targets = targets

    def read_syslog_ng_stats(self):
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as s:
            try:
                s.connect(self.socket)
                s.sendall(b'STATS\n')
                return s.recv(65536).decode('utf-8')
            except FileNotFoundError:
                logging.info("Socket is not exists.")

    def parse_stats(self, stats):
        stat_by_name = {}
        for stat in stats.split("\n"):
            splitted_stat = stat.split(";")
            if (splitted_stat[0] in self.targets.split(",")):
                stat_by_name["%s_%s" % (splitted_stat[1], splitted_stat[4])] = splitted_stat[-1]
        return stat_by_name

    def collect(self):
        stats = self.read_syslog_ng_stats()
        if stats:
            parsed_stats = self.parse_stats(stats)
            for parsed_stat_name, parsed_stat_value in parsed_stats.items():
                g = GaugeMetricFamily("syslog_ng_destination_metric",
                                      "Syslog-ng %s" % parsed_stat_name,
                                      labels=["destination_name"])
                g.add_metric([parsed_stat_name], parsed_stat_value)
                yield g


def parse_parameters():
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", "--port", help="Port on which the exporter listens",
                        default="8000", type=int)
    parser.add_argument("-s", "--socket", help="Location of the socket that will be queried",
                        default="/var/lib/syslog-ng/syslog-ng.ctl")
    parser.add_argument("-t", "--targets",
                        help="Kind of target that needs to be measured in a comma separated list. E.g.: 'source,destination'. Default is destination only.",
                        default="destination")
    return parser.parse_args()


if __name__ == '__main__':
    parameters = parse_parameters()
    logging.basicConfig(stream=sys.stdout, level=logging.DEBUG)
    logging.info("Syslog-NG exporter listens on %s" % parameters.port)
    start_http_server(parameters.port)
    REGISTRY.register(CustomCollector(socket=parameters.socket, targets=parameters.targets))
    while True:
        time.sleep(1)
