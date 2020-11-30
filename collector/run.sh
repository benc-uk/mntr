#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
#GO_BIN=/usr/local/go/bin/go

# Build plugins
(cd $DIR/.. && make plugins)

# Build collector
(cd $DIR/.. && make collector)

# Run with sudo for certain monitors (ping)
sudo $DIR/collector
