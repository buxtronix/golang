// This file contains functionality to do real-time PID control.
package pid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	// Parameters that cannot be changed.
	roParams = []string{"temperature"}
)

// RealTime holds state for the realtime execution.
type RealTime struct {
	// Name is the name of the system to start.
	Name   string
	system System
	mu     sync.Mutex
}

// Begin begins the realtime controller.
func (r *RealTime) Begin(name string) error {
	readOnlyValues = roParams
	if name == "" {
		return errors.New("Must specify a system.")
	}
	r.Name = name
	r.system = GenerateSystem(r.Name)
	http.Handle("/", &indexHandler{})
	http.HandleFunc("/config", r.configHandler)
	http.HandleFunc("/graph", r.graphHandler)
	fmt.Printf("Realtime ready to serve.\n")
	go r.Run()
	http.ListenAndServe(":8080", nil)
	return nil
}

// Run loops through the realtime processing.
func (r *RealTime) Run() {
	for {
		select {
		case <-time.After(time.Duration(r.system.interval) * time.Second):
			fmt.Printf("Processing...\n")
			r.mu.Lock()
			r.system.ProcessInterval()
			r.system.time += r.system.interval
			r.mu.Unlock()
		}
	}
}

// ServeHTTP returns the graph for the supplied parameters.
func (r *RealTime) graphHandler(w http.ResponseWriter, rq *http.Request) {
	if err := rq.ParseForm(); err != nil {
		fmt.Printf("/graph error: %v\n", err)
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.system.SetFormParameters(rq.Form)
	w.Header().Set("Content-Type", "image/png")
	r.system.PngWriter().WriteTo(w)
}

// configHandler returns the current paramters and settings.
func (r *RealTime) configHandler(w http.ResponseWriter, rq *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "text/plain")
	enc.Encode(r.system.AllParameters())
}