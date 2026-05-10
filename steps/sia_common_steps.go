package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/core/types"
)

// SiaCommonSteps holds shared step definitions for Sia operations
type SiaCommonSteps struct{}

// NewSiaCommonSteps creates a new SiaCommonSteps instance
func NewSiaCommonSteps() *SiaCommonSteps {
	return &SiaCommonSteps{}
}

// InitializeScenario registers all shared Sia step definitions with godog
func (s *SiaCommonSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Shared verification steps for Sia operations
	ctx.Step(`^the SIA account key is stored$`, s.theAppKeyIsStoredInTheTestContext)
	ctx.Step(`^the SIA connection is established$`, s.theConnectStatusBecomesApproved)
	ctx.Step(`^the SIA connection fails$`, s.theConnectFlowShouldFailWithAnError)
	ctx.Step(`^the SIA connection is established and an account key is created$`, s.theConnectFlowSucceedsAndAnAppKeyIsCreated)

	ctx.Step(`^the SIA account is ready$`, s.theSiaAccountIsReady)

	// Shared Sia cleanup hook
	ctx.After(s.afterScenarioCleanup)
}

// theAppKeyIsStoredInTheTestContext verifies that a Sia app key exists in context
func (s *SiaCommonSteps) theAppKeyIsStoredInTheTestContext(ctx context.Context) (context.Context, error) {
	appKey, ok := helpers.GetSiaAppKey(ctx)
	if !ok || appKey == "" {
		return ctx, fmt.Errorf("expected SIA app key in context, but it was not found")
	}
	return ctx, nil
}

func (s *SiaCommonSteps) theConnectStatusBecomesApproved(ctx context.Context) (context.Context, error) {
	if sdk, ok := helpers.GetSiaSDK(ctx); ok && sdk != nil {
		return ctx, nil
	}
	if !helpers.GetSiaConnectApproved(ctx) {
		return ctx, fmt.Errorf("SIA connect was not approved")
	}
	return ctx, nil
}

// theConnectFlowShouldFailWithAnError verifies that the Sia connect flow failed
// This checks that no valid SDK was stored in context, indicating the connect was rejected.
func (s *SiaCommonSteps) theConnectFlowShouldFailWithAnError(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if ok && sdk != nil {
		return ctx, fmt.Errorf("expected SIA connect flow to fail, but SDK was created in context")
	}
	return ctx, nil
}

// theConnectFlowSucceedsAndAnAppKeyIsCreated verifies the full connect flow succeeded
// by checking both the SDK and app key are present in context.
func (s *SiaCommonSteps) theConnectFlowSucceedsAndAnAppKeyIsCreated(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("SIA connect flow did not succeed: no SDK in context")
	}

	appKey, ok := helpers.GetSiaAppKey(ctx)
	if !ok || appKey == "" {
		return ctx, fmt.Errorf("SIA connect flow did not succeed: no app key in context")
	}

	return ctx, nil
}

func (s *SiaCommonSteps) theSiaAccountIsReady(ctx context.Context) (context.Context, error) {
	if err := helpers.WaitForSiaReady(ctx); err != nil {
		return ctx, fmt.Errorf("SIA account is not ready: %w", err)
	}
	return ctx, nil
}

// afterScenarioCleanup cleans up Sia objects created during the scenario
func (s *SiaCommonSteps) afterScenarioCleanup(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	// Clean up Sia objects created during the scenario
	objectKeys := helpers.GetSiaObjectsCleanup(ctx)
	for _, keyStr := range objectKeys {
		var key types.Hash256
		if _, parseErr := fmt.Sscanf(keyStr, "%x", &key); parseErr != nil {
			continue
		}
		if deleteErr := helpers.SiaDeleteObject(ctx, key); deleteErr != nil {
			helpers.NewLogger(sc.Name).Warn(ctx, "Failed to delete Sia object %s: %v", keyStr, deleteErr)
		}
	}

	return ctx, nil
}
