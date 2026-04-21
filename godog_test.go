package e2e_test

import (
	"flag"
	"os"
	"testing"

	"context"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"pinner.xyz-e2e/helpers"
	"pinner.xyz-e2e/steps"
)

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "pretty",
	Concurrency: 1,
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

	// Initialize common steps BEFORE service-specific steps
	// Godog uses first-registered step handler, so common steps must initialize first
	commonAuthSteps := steps.NewCommonAuthSteps()
	commonAuthSteps.InitializeScenario(ctx)

	commonPasswordSteps := steps.NewCommonPasswordSteps()
	commonPasswordSteps.InitializeScenario(ctx)

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

	// Initialize quota management steps (user-side)
	quotaSteps := steps.NewQuotaSteps()
	quotaSteps.InitializeScenario(ctx)

	// Initialize admin quota management steps
	adminQuotaSteps := steps.NewAdminQuotaSteps()
	adminQuotaSteps.InitializeScenario(ctx)

	// Initialize admin billing management steps
	adminBillingSteps := steps.NewAdminBillingSteps()
	adminBillingSteps.InitializeScenario(ctx)

	// Initialize admin subscription management steps (gateway-agnostic)
	adminSubscriptionSteps := steps.NewAdminSubscriptionSteps()
	adminSubscriptionSteps.InitializeScenario(ctx)

	// Initialize gateway configuration steps (MUST be first for subscription tests)
	gatewaySteps := steps.NewGatewaySteps()
	gatewaySteps.InitializeScenario(ctx)

	// Initialize generic subscription common steps (gateway-agnostic)
	subscriptionCommonSteps := steps.NewSubscriptionCommonSteps()
	subscriptionCommonSteps.InitializeScenario(ctx)

	// Initialize user subscription steps (gateway-agnostic)
	userSubscriptionSteps := steps.NewUserSubscriptionSteps()
	userSubscriptionSteps.InitializeScenario(ctx)

	// Initialize Stripe-specific subscription steps
	stripeSubscriptionSteps := steps.NewStripeSubscriptionSteps()
	stripeSubscriptionSteps.InitializeScenario(ctx)

	// Initialize Stripe manual control steps
	stripeManualControlSteps := steps.NewStripeManualControlSteps()
	stripeManualControlSteps.InitializeScenario(ctx)

	// Initialize Atlos-specific subscription steps
	atlosSubscriptionSteps := steps.NewAtlosSubscriptionSteps()
	atlosSubscriptionSteps.InitializeScenario(ctx)

	// Initialize Atlos manual control steps
	atlosManualControlSteps := steps.NewAtlosManualControlSteps()
	atlosManualControlSteps.InitializeScenario(ctx)

	// Initialize quota enforcement steps
	quotaEnforcementSteps := steps.NewQuotaEnforcementSteps()
	quotaEnforcementSteps.InitializeScenario(ctx)

	// Register scenario-level hooks
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Set default gateway at the start of each scenario
		ctx = helpers.SetActiveGateway(ctx, helpers.GatewayDefault)

		// Reset stripe-mock before each scenario to ensure clean state
		// This clears all data: customers, subscriptions, products, prices, webhooks
		if err := helpers.ResetStripeMock(ctx); err != nil {
			helpers.NewLogger(sc.Name).Error(ctx, "Failed to reset stripe-mock: %v", err)
		}
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Reset payment gateway mocks after each scenario
		if helpers.GatewayIsStripe(ctx) {
			if err := helpers.ResetStripeMock(ctx); err != nil {
				helpers.NewLogger(sc.Name).Error(ctx, "Failed to reset stripe-mock after scenario: %v", err)
			}
		}
		// Cleanup billing infrastructure if created
		if _, _, ok := helpers.GetBillingInfrastructure(ctx); ok {
			if err := helpers.CleanupBillingInfrastructure(ctx); err != nil {
				helpers.NewLogger(sc.Name).Error(ctx, "Failed to cleanup billing infrastructure: %v", err)
			}
		}
		return ctx, nil
	})
}
