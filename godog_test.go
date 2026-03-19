package e2e_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
	"pinner.xyz-e2e/helpers"
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
		Name:                "e2e",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	// Print scenario timing summary after all tests complete
	helpers.PrintScenarioTimings()

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
	// Register common hooks before initializing any step definitions
	// This ensures cleanup tracking for all scenarios
	helpers.RegisterCommonHooks(ctx)

	// Initialize auth steps
	authSteps := steps.NewAuthSteps()
	authSteps.InitializeScenario(ctx)

	// Initialize password profile steps
	passwordProfileSteps := steps.NewPasswordProfileSteps()
	passwordProfileSteps.InitializeScenario(ctx)

	// Initialize account steps
	accountSteps := steps.NewAccountSteps()
	accountSteps.InitializeScenario(ctx)

	// Initialize password reset steps
	passwordResetSteps := steps.NewPasswordResetSteps()
	passwordResetSteps.InitializeScenario(ctx)

	// Initialize IPFS common steps (shared wait/verification steps)
	// Must be registered before service-specific IPFS steps
	ipfsCommonSteps := steps.NewIPFSCommonSteps()
	ipfsCommonSteps.InitializeScenario(ctx)

	// Initialize IPFS upload steps
	uploadSteps := steps.NewIPFSUploadSteps()
	uploadSteps.InitializeScenario(ctx)

	// Initialize IPFS pinning steps
	pinningSteps := steps.NewPinningSteps()
	pinningSteps.InitializeScenario(ctx)

	// Initialize IPFS content list steps
	contentListSteps := steps.NewContentListSteps()
	contentListSteps.InitializeScenario(ctx)
}
