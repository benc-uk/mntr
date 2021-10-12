#!/bin/bash
set -e

# Start server with local data dir
./influxd --bolt-path ./data/influxdb.bolt --engine-path ./data