package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/siastorage"
)

// SiaConnectSteps holds step definitions for Sia connect flow
type SiaConnectSteps struct {
	builder *siastorage.Builder
}

// NewSiaConnectSteps creates a new SiaConnectSteps instance
func NewSiaConnectSteps() *SiaConnectSteps {
	return &SiaConnectSteps{}
}

// InitializeScenario registers all Sia connect step definitions with godog
func (s *SiaConnectSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user initiates a SIA account connection$`, s.theUserInitiatesASIAConnectViaThePortal)
	ctx.Step(`^the SIA connection is pending approval$`, s.thePortalReturnsAResponseURLStatusURLAndRegisterURL)
	ctx.Step(`^the user approves the SIA connection$`, s.theUserApprovesTheConnectViaThePortal)
	ctx.Step(`^the user saves their SIA recovery phrase$`, s.theUserRegistersTheAppUsingAMnemonic)
	ctx.Step(`^the user rejects the SIA connection$`, s.theUserRejectsTheConnectFromThePortal)
	ctx.Step(`^the user retries the SIA connection$`, s.theUserRetriesTheConnectFlow)
}

// theUserInitiatesASIAConnectViaThePortal starts the Sia connect flow by requesting a connection
func (s *SiaConnectSteps) theUserInitiatesASIAConnectViaThePortal(ctx context.Context) (context.Context, error) {
	indexerURL := helpers.GetSiaEndpoint()
	if indexerURL == "" {
		return ctx, fmt.Errorf("SIA endpoint not configured")
	}

	metadata := helpers.SiaTestAppMetadata()

	builder := siastorage.NewBuilder(indexerURL, metadata)
	s.builder = builder

	responseURL, err := builder.RequestConnection(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to request SIA connection: %w", err)
	}

	ctx = helpers.SetSiaConnectRequestID(ctx, responseURL)
	ctx = helpers.SetSiaAppID(ctx, metadata.ID.String())

	return ctx, nil
}

// thePortalReturnsAResponseURLStatusURLAndRegisterURL verifies the connect request returned valid URLs
func (s *SiaConnectSteps) thePortalReturnsAResponseURLStatusURLAndRegisterURL(ctx context.Context) (context.Context, error) {
	requestID, ok := helpers.GetSiaConnectRequestID(ctx)
	if !ok || requestID == "" {
		return ctx, fmt.Errorf("SIA connect request did not return a response URL")
	}

	if s.builder == nil {
		return ctx, fmt.Errorf("SIA connect builder not initialized")
	}

	return ctx, nil
}

// theUserApprovesTheConnectViaThePortal waits for the connect request to be approved
// and registers the app with a generated mnemonic
func (s *SiaConnectSteps) theUserApprovesTheConnectViaThePortal(ctx context.Context) (context.Context, error) {
	if s.builder == nil {
		return ctx, fmt.Errorf("SIA connect builder not initialized")
	}

	connectURL, _ := helpers.GetSiaConnectRequestID(ctx)
	requestID, err := helpers.ParseRequestIDFromConnectURL(connectURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse SIA connect request ID: %w", err)
	}

	if err := helpers.ApproveSiaConnect(ctx, requestID); err != nil {
		return ctx, fmt.Errorf("failed to approve SIA connection: %w", err)
	}

	approved, err := s.builder.WaitForApproval(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to wait for SIA approval: %w", err)
	}
	if !approved {
		return ctx, fmt.Errorf("SIA connection request was denied")
	}

	ctx = helpers.SetSiaConnectApproved(ctx)
	return ctx, nil
}

// theUserRegistersTheAppUsingAMnemonic registers the Sia app with a generated mnemonic
func (s *SiaConnectSteps) theUserRegistersTheAppUsingAMnemonic(ctx context.Context) (context.Context, error) {
	if s.builder == nil {
		return ctx, fmt.Errorf("SIA connect builder not initialized")
	}

	mnemonic := siastorage.NewSeedPhrase()

	sdk, err := s.builder.Register(ctx, mnemonic)
	if err != nil {
		return ctx, fmt.Errorf("failed to register SIA app: %w", err)
	}

	appKey := sdk.AppKey()
	appKeyStr := fmt.Sprintf("%x", appKey[:])

	ctx = helpers.SetSiaSDK(ctx, sdk)
	ctx = helpers.SetSiaMnemonic(ctx, mnemonic)
	ctx = helpers.SetSiaAppKey(ctx, appKeyStr)

	return ctx, nil
}

func (s *SiaConnectSteps) theUserRejectsTheConnectFromThePortal(ctx context.Context) (context.Context, error) {
	connectURL, _ := helpers.GetSiaConnectRequestID(ctx)
	requestID, err := helpers.ParseRequestIDFromConnectURL(connectURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse SIA connect request ID: %w", err)
	}

	if err := helpers.RejectSiaConnect(ctx, requestID); err != nil {
		return ctx, fmt.Errorf("failed to reject SIA connection: %w", err)
	}

	return ctx, nil
}

// theUserRetriesTheConnectFlow performs the full Sia connect flow from scratch
func (s *SiaConnectSteps) theUserRetriesTheConnectFlow(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = helpers.SiaSDKConnect(ctx)
	if err != nil {
		return ctx, fmt.Errorf("SIA connect retry failed: %w", err)
	}
	return ctx, nil
}
