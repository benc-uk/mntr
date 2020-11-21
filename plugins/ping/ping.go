package main

import (
	"fmt"
	"log"
	"time"

	"github.com/benc-uk/mntr/collector/results"
	"github.com/benc-uk/mntr/collector/types"
	"github.com/go-ping/ping"
)

// Holds specific data struct and params for this plugin encapsulating the generic fields too
type monitor struct {
	// Generic & common fields here
	types.Monitor
	// Fields specific to this plugin
	Params struct {
		Host       string  `validate:"required"`
		Count      int     `validate:"min=1" default:"5"`
		Timeout    int     `validate:"min=1" default:"5000"`
		PacketLoss float64 `yaml:"packetLoss" validate:"max=100.0" default:"0.0"`
	}
}

// NewMonitor returns this plugin's config - MANDATORY function for all mntr plugins
func NewMonitor() interface{} {
	return &monitor{}
}

// Run this plugin monitor - MANDATORY function for all mntr plugins
func Run(rawMonitor interface{}) results.Result {
	result := results.NewResult()

	// All monitors require this step to assert the raw config into typed struct
	mon, ok := rawMonitor.(*monitor)
	if !ok {
		// Should never get here
		log.Fatalln("FATAL! Unable to convert monitor struct, that's really bad!")
	}

	pinger, err := ping.NewPinger(mon.Params.Host)
	pinger.Timeout = time.Duration(mon.Params.Timeout) * time.Millisecond
	if err != nil {
		result.Fail(err.Error())
		return result
	}
	pinger.SetPrivileged(true)
	pinger.Count = mon.Params.Count

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		result.Fail(err.Error())
		return result
	}

	stats := pinger.Statistics()
	result.AddFloat("packet_loss", stats.PacketLoss)
	result.AddFloat("round_trip_time", float64(stats.AvgRtt))

	// Check packet loss percentage
	if stats.PacketLoss > mon.Params.PacketLoss {
		result.Fail(fmt.Sprintf("Packet loss was too high: %f%%", stats.PacketLoss))
	}

	return result
}
