package e2e_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
	"pinner.xyz-e2e/steps"
)

var opts = godog.Options{
	Output: os.Stdout,
	Format: "pretty",
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	opts.Paths = pflag.Args()
	if len(opts.Paths) == 0 {
		opts.Paths = []string{"features"}
	}

	status := godog.TestSuite{
		Name:                 "e2e",
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()

	// Optional: Run `testing` package's logic besides godog.
	if st := m.Run(); st > status {
		status = st
	}

	if status != 0 {
		fmt.Fprintln(os.Stderr, "godog tests failed")
	}

	os.Exit(status)
}

// InitializeScenario initializes the scenario context with step definitions
func InitializeScenario(ctx *godog.ScenarioContext) {
	// Initialize auth steps
	authSteps := steps.NewAuthSteps()
	authSteps.InitializeScenario(ctx)

	// Initialize password profile steps
	passwordProfileSteps := steps.NewPasswordProfileSteps()
	passwordProfileSteps.InitializeScenario(ctx)

	// Initialize account steps
	accountSteps := steps.NewAccountSteps()
	accountSteps.RegisterHooks(ctx)
	accountSteps.InitializeScenario(ctx)

	// Initialize password reset steps
	passwordResetSteps := steps.NewPasswordResetSteps()
	passwordResetSteps.InitializeScenario(ctx)
}
