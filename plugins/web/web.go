package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/benc-uk/mntr/collector/results"
	"github.com/benc-uk/mntr/collector/types"
	"github.com/stretchr/stew/slice"
)

// Holds specific data struct and params for this plugin encapsulating the generic fields too
type monitor struct {
	// Generic & common fields here
	types.Monitor
	// Fields specific to this plugin
	Params struct {
		URL      string `validate:"url"`
		Method   string `validate:"oneof=GET POST PUT DELETE OPTIONS HEAD" default:"GET"`
		Statuses []int  `validate:"required" default:"[200]"`
		Expect   string
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

	// Various timers
	var start, connStart, dnsStart, tlsStart time.Time

	// A HTTP request is a good place to start!
	req, err := http.NewRequest(mon.Params.Method, mon.Params.URL, nil)

	// Set up callbacks for HTTP tracing
	trace := &httptrace.ClientTrace{
		// DNSStart: func(dsi httptrace.DNSStartInfo) {
		// 	fmt.Println("DNS START")
		// 	dnsStart = time.Now()
		// },
		// DNSDone: func(ddi httptrace.DNSDoneInfo) {
		// 	fmt.Printf("DNS Done: %v\n", time.Since(dnsStart))
		// },

		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			result.AddFloat("tls_time", float64(time.Since(tlsStart).Milliseconds()))
		},

		ConnectStart: func(network, addr string) { connStart = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			result.AddFloat("connect_time", float64(time.Since(connStart).Milliseconds()))
		},

		GotFirstResponseByte: func() {
			result.AddFloat("first_byte_time", float64(time.Since(start).Milliseconds()))
		},
	}

	// Custom resover and client for finer control
	resolver := &net.Resolver{
		PreferGo: true,
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 1 * time.Second,
				Resolver:  resolver,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// Start the clocks
	start = time.Now()
	dnsStart = time.Now()

	// Custom DNS lookup as DNS gooks in httptrace simply is never called!
	url, err := url.Parse(mon.Params.URL)
	if err == nil {
		_, err := resolver.LookupHost(context.TODO(), url.Hostname())
		if err == nil {
			result.AddFloat("dns_time", float64(time.Since(dnsStart).Milliseconds()))
		}
	}

	// Start request with httptrace RoundTrip
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := client.Transport.RoundTrip(req)
	if err != nil {
		result.AddError(err.Error())
		result.Status = results.StatusFailed
		return result
	}

	// Read the body for a complete response time
	contentBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// Stop the clock
	result.AddFloat("total_time", float64(time.Since(start).Milliseconds()))

	// Check status codes
	if !slice.ContainsInt(mon.Params.Statuses, resp.StatusCode) {
		result.AddError(fmt.Sprintf("Status code was %d, not one of: %v", resp.StatusCode, mon.Params.Statuses))
		result.Status = results.StatusFailed
	}

	// Basic content check
	// TODO: Add regex support
	contentString := string(contentBytes)
	if mon.Params.Expect != "" {
		if !strings.Contains(contentString, mon.Params.Expect) {
			result.AddError(fmt.Sprintf("Expected string '%s' not found in HTTP content", mon.Params.Expect))
			result.Status = results.StatusFailed
		}
	}

	return result
}
