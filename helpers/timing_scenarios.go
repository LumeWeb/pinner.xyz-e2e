package helpers

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

// ScenarioTiming holds timing data for a single scenario
type ScenarioTiming struct {
	Name     string
	Duration time.Duration
}

var (
	// scenarioTimings stores collected scenario timings
	scenarioTimings []ScenarioTiming
	// timingsMutex protects concurrent access to scenarioTimings
	timingsMutex sync.Mutex
)

// RecordScenarioTiming records timing data for a completed scenario
func RecordScenarioTiming(name string, duration time.Duration) {
	timingsMutex.Lock()
	defer timingsMutex.Unlock()

	scenarioTimings = append(scenarioTimings, ScenarioTiming{
		Name:     name,
		Duration: duration,
	})
}

// PrintScenarioTimings prints all collected scenario timings to stdout
func PrintScenarioTimings() {
	timingsMutex.Lock()
	defer timingsMutex.Unlock()

	if len(scenarioTimings) == 0 {
		return
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "=== Scenario Timing Summary ===")
	fmt.Fprintln(os.Stdout, "")

	// Sort by name
	sort.Slice(scenarioTimings, func(i, j int) bool {
		return scenarioTimings[i].Name < scenarioTimings[j].Name
	})

	// Print all timings
	for _, timing := range scenarioTimings {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("  %s: %v", timing.Name, timing.Duration))
	}

	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, fmt.Sprintf("Total scenarios: %d", len(scenarioTimings)))
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "=============================")
	fmt.Fprintln(os.Stdout, "")
}
