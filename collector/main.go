package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"github.com/benc-uk/mntr/collector/results"
	"github.com/benc-uk/mntr/collector/types"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxapi "github.com/influxdata/influxdb-client-go/v2/api"
	_ "github.com/joho/godotenv/autoload" // Autoloads .env file if it exists
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
)

var monitorCancelers []context.CancelFunc
var configHash = ""
var collectorName = ""

const maxDelay = 1000 // The random delay added to start of new monitor loops
const pluginDir = "./plugins"
const serverCheckInterval = 60 // Seconds between server checkins

type pluginSymbols struct {
	RunFunc     plugin.Symbol
	MonitorFunc plugin.Symbol
}

type configData struct {
	serverEndpoint   string
	influxDBEndpoint string
	influxDBToken    string
	influxDBOrg      string
	influxDBBucket   string
}

var plugins map[string]pluginSymbols
var config configData
var influxAPI influxapi.WriteAPI

//
// Entry point to start the collector
//
func main() {
	rand.Seed(time.Now().UnixNano())

	// List of all plugins key'ed on plugin name
	plugins = make(map[string]pluginSymbols)
	config = configData{}

	log.Println("🚀 Mntr collector is starting...")

	// Get config env vars
	if config.serverEndpoint = os.Getenv("MNTR_SERVER_ENDPOINT"); config.serverEndpoint == "" {
		log.Fatalln("💥 FATAL! MNTR_SERVER_ENDPOINT is not set")
	}
	if config.influxDBEndpoint = os.Getenv("MNTR_INFLUXDB_ENDPOINT"); config.influxDBEndpoint == "" {
		log.Fatalln("💥 FATAL! MNTR_INFLUXDB_ENDPOINT is not set")
	}
	if config.influxDBToken = os.Getenv("MNTR_INFLUXDB_TOKEN"); config.influxDBToken == "" {
		log.Fatalln("💥 FATAL! MNTR_INFLUXDB_TOKEN is not set")
	}
	if config.influxDBOrg = os.Getenv("MNTR_INFLUXDB_ORG"); config.influxDBOrg == "" {
		config.influxDBOrg = "mntr"
	}
	if config.influxDBBucket = os.Getenv("MNTR_INFLUXDB_BUCKET"); config.influxDBBucket == "" {
		config.influxDBBucket = "mntr"
	}
	log.Printf("%+v", config)

	// Find hostname
	collectorName, _ = os.Hostname()
	hostnameOverride := os.Getenv("MNTR_HOSTNAME")
	if hostnameOverride != "" {
		collectorName = hostnameOverride
	}
	if collectorName == "" {
		log.Fatalln("💥 FATAL! Could not get hostname from OS and MNTR_HOSTNAME is not set")
	}
	log.Printf("🌐 Hostname is: %s\n", collectorName)
	log.Printf("🏠 Monitor config will be fetched from: %s\n", config.serverEndpoint)

	// Create InfluxDB client
	client := influxdb2.NewClientWithOptions(config.influxDBEndpoint, config.influxDBToken,
		influxdb2.DefaultOptions().SetBatchSize(20))
	// Get non-blocking write client
	influxAPI = client.WriteAPI(config.influxDBOrg, config.influxDBBucket)
	defer client.Close()

	// Check disk for plugins and load them
	loadPlugins()

	// Main loop, stops exit and polls for config changes
	for {
		loadMonitors(config.serverEndpoint)
		time.Sleep(serverCheckInterval * time.Second)
	}
}

//
// Wrapper loop for running monitors on a given frequency
//
func monitorRunner(ctx context.Context, runFunc plugin.Symbol, monBase types.Monitor, monitor interface{}) {
	// Random delay to stop all monitors starting at exactly the same time
	delayMs := rand.Intn(maxDelay)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Monitor loop, infinite until ctx is canceled
	for {
		select {
		case <-ctx.Done():
			log.Printf("💢 '%s/%s' was canceled\n", monBase.Plugin, monBase.Name)
			return
		default:
			// Invoke the plugin run function
			log.Printf("🔹 Starting '%s/%s'\n", monBase.Plugin, monBase.Name)
			//Slight weird type assertion using func() and the function signature
			result := runFunc.(func(interface{}) results.Result)(monitor)

			symbol := ""
			switch result.Status {
			case results.StatusOK:
				symbol = "✅"
			case results.StatusFailed:
				symbol = "❌"
			case results.StatusFatal:
				symbol = "💥"
			default:
				symbol = "❓"
			}

			log.Printf("%s Result of '%s/%s' was: %v with %d errors\n", symbol, monBase.Plugin, monBase.Name, result.Status, len(result.Errors))
			// Log errors
			for _, err := range result.Errors {
				log.Printf("   - %s", err)
			}

			// Log metrics
			for name, val := range result.Metrics {
				log.Printf("   - %s=%f", name, val)
			}
			result.AddFloat("status", float64(result.Status))

			// Create and send data point to InfuxDB
			point := influxdb2.NewPoint(
				monBase.Plugin,
				map[string]string{
					"name":      monBase.Name,
					"collector": collectorName,
				},
				result.Metrics,
				time.Now())

			influxAPI.WritePoint(point)

			// TODO: Maybe remove? Force all unwritten data to be sent
			influxAPI.Flush()
		}

		// Wait until next run
		time.Sleep(time.Duration(monBase.Frequency) * time.Second)
	}
}

//
// Load monitor config from the server and set up all monitors
//
func loadMonitors(serverEndpoint string) {
	// Try to load monitor config from server
	log.Printf("👋 Contacting server for monitor config\n")
	resp, err := http.Get(fmt.Sprintf("%s/api/monitors/config", serverEndpoint))
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("💥 FATAL! Failed to get monitor config from server %s\n", err)
	}
	defer resp.Body.Close()
	configBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("💥 FATAL! Failed reading monitor config data %v\n", err)
	}

	// Compare config hash to detect changes
	newConfigHash := fmt.Sprintf("%x", md5.Sum(configBytes))
	if newConfigHash != configHash {
		configHash = newConfigHash
		log.Println("🧾 Config has changed, will cancel & restart all monitors")

		// Cancel all existing monitors (if any)
		for _, cancel := range monitorCancelers {
			cancel()
		}
		// Reset canceller array
		monitorCancelers = make([]context.CancelFunc, 0)

		// Split the config doc on YAML doc seperator
		// Sounds lame, trust me, it took a LONG time to this as a solution
		monDocs := strings.Split(string(configBytes), "\n---\n")

		// Parse each one as a separate monitor configuration
		for _, monYamlDoc := range monDocs {
			// STEP 1 - YAML parse generic/shared top level fields
			base := types.Monitor{
				Enabled: true,
			}

			err = yaml.Unmarshal([]byte(monYamlDoc), &base)
			if err != nil {
				log.Printf("🔥 WARNING! Failed parsing monitor data %v\n", err)
				continue
			}

			// Monitor config checks
			if base.Name == "" || base.Plugin == "" {
				log.Printf("⛔ Monitor is missing name and/or plugin, can not be configured\n")
				continue
			}
			if base.Frequency <= 0 {
				log.Printf("⛔ Monitor '%s/%s' is missing frequency, can not be configured\n", base.Plugin, base.Name)
				continue
			}
			if !base.Enabled {
				log.Printf("⛔ Monitor '%s/%s' is diasbled\n", base.Plugin, base.Name)
				continue
			}
			if len(base.RunsOn) > 0 {
				if !funk.ContainsString(base.RunsOn, collectorName) {
					log.Printf("⛔ Monitor '%s/%s' isn't set to run on this collector\n", base.Plugin, base.Name)
					continue
				}
			}

			// Use the plugin name fetched from first YAML parse
			// - to get hold of a function to create a fully typed config struct for this plugin
			newConfigFunc := plugins[base.Plugin].MonitorFunc
			if newConfigFunc == nil {
				log.Printf("🔥 WARNING! Unable to find plugin: %s\n", base.Plugin)
				continue
			}
			// Create the full monitor struct
			monitor := newConfigFunc.(func() interface{})()

			// STEP 2 - YAML parse into full monitor params struct
			err = yaml.Unmarshal([]byte(monYamlDoc), monitor)
			if err != nil {
				log.Printf("🔥 WARNING! Failed parsing monitor data %v\n", err)
				continue
			}

			// Set defaults in monitor params
			if err := defaults.Set(monitor); err != nil {
				log.Printf("🔥 WARNING! Failed setting defaults for %s - %v\n", base.Name, err)
				continue
			}

			// Run validation checks on monitor params
			var validate *validator.Validate
			validate = validator.New()
			if err := validate.Struct(monitor); err != nil {
				log.Printf("🔥 WARNING! Validation failed for monitor %s - %v\n", base.Name, err)
				continue
			}

			// Find plugin, and the symbol to invoke/run the plugin
			runnerFunc := plugins[base.Plugin].RunFunc
			if runnerFunc == nil {
				log.Printf("🔥 WARNING! Unable to find plugin: %s\n", base.Plugin)
				continue
			}

			log.Printf("⚡ Monitor %s/%s was configured OK\n", base.Plugin, base.Name)

			// Set up the monitor loop to run in a separate go routine
			ctx, cancel := context.WithCancel(context.Background())
			go monitorRunner(ctx, runnerFunc, base, monitor)
			monitorCancelers = append(monitorCancelers, cancel)
		}
	}
}

//
// Load plugins from the filesystem
//
func loadPlugins() {
	// Try to load plugins
	log.Println("🧩 Loading plugins...")

	err := filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		// Load all .so files in plugin dir
		if strings.HasSuffix(path, ".so") {
			// Open the shared object file
			plugin, err := plugin.Open(path)
			if err != nil {
				log.Printf("🔥 WARNING! Failed to load %s as a plugin, probably not a shared library file\n", path)
				return nil
			}

			// Important: Lookup the exported Run function symbol, all plugins *MUST* export this
			runner, err := plugin.Lookup("Run")
			if err != nil {
				log.Printf("🔥 WARNING! Failed to load %s as a plugin, exported Run() function was not found\n", path)
				return nil
			}

			// Important: Lookup the exported NewMonitor function symbol, all plugins *MUST* export this
			newConfig, err := plugin.Lookup("NewMonitor")
			if err != nil {
				log.Printf("🔥 WARNING! Failed to load %s as a plugin, exported NewMonitor() function was not found\n", path)
				return nil
			}
			// Get short plugin name
			name := path[8:]
			name = name[:len(name)-3]
			log.Printf("  - Loaded plugin: %s", name)

			// Store the exported symbol mapped to name for lookup and reuse
			plugins[name] = pluginSymbols{
				RunFunc:     runner,
				MonitorFunc: newConfig,
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("🔥 FATAL! Unable to read plugins directory %s\n", pluginDir)
	}
}
