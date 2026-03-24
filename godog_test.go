package e2e_test

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"pinner.xyz-e2e/helpers"
	"pinner.xyz-e2e/steps"
)

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "pretty",
	Concurrency: 4,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestFeatures(t *testing.T) {
	status := godog.TestSuite{
		Name:                "e2e",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	// Print scenario timing summary after all tests complete
	helpers.PrintScenarioTimings()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
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

	// Initialize IPNS common steps (shared wait/verification steps)
	// Must be registered before service-specific IPNS steps
	ipnsCommonSteps := steps.NewIPNSCommonSteps()
	ipnsCommonSteps.InitializeScenario(ctx)

	// Initialize IPNS key management steps
	ipnsKeyManagementSteps := steps.NewIPNSKeyManagementSteps()
	ipnsKeyManagementSteps.InitializeScenario(ctx)

	// Initialize IPNS publishing and resolution steps
	ipnsPublishingResolutionSteps := steps.NewIPNSPublishingResolutionSteps()
	ipnsPublishingResolutionSteps.InitializeScenario(ctx)

	// Initialize DNS background steps (shared authentication/IPFS steps)
	// Must be registered before any other DNS steps
	dnsBackgroundSteps := steps.NewDNSBackgroundSteps()
	dnsBackgroundSteps.InitializeScenario(ctx)

	// Initialize DNS common steps (shared verification steps)
	// Must be registered before service-specific DNS steps
	dnsCommonSteps := steps.NewDNSCommonSteps()
	dnsCommonSteps.InitializeScenario(ctx)

	// Initialize DNS zone management steps
	dnsZoneSteps := steps.NewDNSZoneSteps()
	dnsZoneSteps.InitializeScenario(ctx)

	// Initialize DNS record management steps
	dnsRecordSteps := steps.NewDNSRecordSteps()
	dnsRecordSteps.InitializeScenario(ctx)

	// Initialize website common steps (shared verification steps)
	// Must be registered before service-specific website steps
	websiteCommonSteps := steps.NewWebsiteCommonSteps()
	websiteCommonSteps.InitializeScenario(ctx)

	// Initialize website management steps
	websiteSteps := steps.NewWebsiteSteps()
	websiteSteps.InitializeScenario(ctx)

	// Initialize website DNS validation steps
	websiteDNSSteps := steps.NewWebsiteDNSSteps()
	websiteDNSSteps.InitializeScenario(ctx)

	// Initialize website IPNS verification steps
	websiteIPNSSteps := steps.NewWebsiteIPNSSteps()
	websiteIPNSSteps.InitializeScenario(ctx)

	// Initialize website SSL verification steps
	websiteSSLSteps := steps.NewWebsiteSSLSteps()
	websiteSSLSteps.InitializeScenario(ctx)
}
