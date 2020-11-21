package main

import (
	"log"

	"github.com/benc-uk/mntr/collector/results"
	"github.com/benc-uk/mntr/collector/types"
)

// Monitor is specific data struct and params for this plugin encapsulating the generic fields too
type Monitor struct {
	types.Monitor
	Params struct {
		Something string `validate:"required"`
		SomeElse  string `validate:"required" default:"GET"`
		SomeInts  []int  `validate:"required" default:"[1,2,3]"`
	}
}

func NewMonitor() interface{} {
	return &Monitor{}
}

// Run this monitor
func Run(rawMonitor interface{}) results.Result {
	result := results.NewResult()

	mon, ok := rawMonitor.(*Monitor)
	if !ok {
		// Should never get here
		log.Fatalln("FATAL! Unable to convert monitor struct, that's really bad!")
	}

	// Add any metrics
	result.AddFloat("SomeMetric", 1.836)

	// Conditionally work based on params
	if mon.Params.Something != "" {
		if 5 != 9 {
			// And set errors
			result.AddError("This is an example of an error")
			result.Status = results.StatusFailed
		}
	}

	return result
}
