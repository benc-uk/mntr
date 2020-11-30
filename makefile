# ------------------------------------------------------------
# Licensed under the MIT License.
# ------------------------------------------------------------

################################################################################
# Variables
################################################################################
CGO := 0
BUILD_INFO ?= "Makefile build"
SERVER_DIR = ./server
PLUGINS_DIR = ./plugins
COLLECTOR_DIR = ./collector
FRONTEND_DIR = ./frontend
INFLUX_DIR = ./influxdb
PLUGINS = web ping

# Most likely want to override these when calling `make docker`
DOCKER_REG ?= ghcr.io
DOCKER_REPO ?= benc-uk/mntr
DOCKER_TAG ?= latest
DOCKER_PREFIX := $(DOCKER_REG)/$(DOCKER_REPO)

# Change if you know what you're doing
CLIENT_ID ?= ""

################################################################################
# Lint & format check server
################################################################################
.PHONY: lint-server
lint-server : 
	@echo '### Linting and formatting $(SERVER_DIR)'
	npx prettier --check $(SERVER_DIR)
	deno --unstable lint $(SERVER_DIR)

################################################################################
# Lint & format check frontend
################################################################################
.PHONY: lint-frontend
lint-frontend : $(FRONTEND_DIR)/node_modules
	@echo '### Linting and formatting $(FRONTEND_DIR)'
	npx prettier --check $(FRONTEND_DIR)
	@cd $(FRONTEND_DIR); npm run lint
	
################################################################################
# Lint & format check plugins
################################################################################
.PHONY: lint-plugins
lint-plugins : 
	@echo '### Linting and formatting $(PLUGINS_DIR)'
	scripts/gofmt.sh $(PLUGINS_DIR)

################################################################################
# Lint & format check everything
################################################################################
.PHONY: lint
lint: lint-server lint-plugins lint-frontend

################################################################################
# Frontend / Vue.js
################################################################################
frontend : $(FRONTEND_DIR)/node_modules
	cd $(FRONTEND_DIR); npm run build
	cd $(SERVICE_DIR)/frontend-host; go build

$(FRONTEND_DIR)/node_modules: $(FRONTEND_DIR)/package.json
	cd $(FRONTEND_DIR); npm install --silent
	touch -m $(FRONTEND_DIR)/node_modules

$(FRONTEND_DIR)/package.json: 
	@echo "package.json was modified"

################################################################################
# Server + plugins
################################################################################
.PHONY: server
server : plugins
	@rm -rf $(SERVER_DIR)/plugins
	@mkdir -p $(SERVER_DIR)/plugins
	@for plugin in $(PLUGINS) ; do \
		cp $(PLUGINS_DIR)/$$plugin/*.yaml $(SERVER_DIR)/plugins ; \
	done	

################################################################################
# Collector
################################################################################
.PHONY: collector
collector : plugins
	go build -o $(COLLECTOR_DIR)/collector $(COLLECTOR_DIR)/main.go 
	@rm -rf $(COLLECTOR_DIR)/plugins
	@mkdir -p $(COLLECTOR_DIR)/plugins
	@for plugin in $(PLUGINS) ; do \
		cp $(PLUGINS_DIR)/$$plugin/*.so $(COLLECTOR_DIR)/plugins ; \
	done

################################################################################
# Plugins only
################################################################################
.PHONY: plugins
plugins : 
	@for plugin in $(PLUGINS) ; do \
		go build -buildmode=plugin -o $(PLUGINS_DIR)/$$plugin/$$plugin.so $(PLUGINS_DIR)/$$plugin ; \
	done

################################################################################
# Cleanup
################################################################################
.PHONY : clean
clean :
	-rm $(PLUGINS_DIR)/**/*.so
	-rm $(COLLECTOR_DIR)/collector
	-rm -rf $(SERVER_DIR)/plugins
	-rm -rf $(COLLECTOR_DIR)/plugins

################################################################################
# Cleanup DB
################################################################################
.PHONY : cleandb
cleandb :
	-rm -rf $(SERVER_DIR)/data
	-rm -rf $(INFLUX_DIR)/data