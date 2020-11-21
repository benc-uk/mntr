#!/bin/bash
set -e

INFLUX_VER=2.0.2

# Fetch and extract
wget https://dl.influxdata.com/influxdb/releases/influxdb-${INFLUX_VER}_linux_amd64.tar.gz
tar -xzf influxdb-${INFLUX_VER}_linux_amd64.tar.gz
mv influxdb-${INFLUX_VER}_linux_amd64/* .
rm influxdb-${INFLUX_VER}_linux_amd64.tar.gz
rm -rf influxdb-${INFLUX_VER}_linux_amd64/
chmod +x influx
chmod +x influxd

# Check version
./influx version
./influxd version
