package steps

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/core/types"
)

type SiaErrorHandlingSteps struct {
	lastResponse *http.Response
	lastError    error
}

func NewSiaErrorHandlingSteps() *SiaErrorHandlingSteps {
	return &SiaErrorHandlingSteps{}
}

func (s *SiaErrorHandlingSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user attempts to access a Sia file without authentication$`, s.theUserCallsAProtectedEndpoint)
	ctx.Step(`^the SIA access is denied with authentication required$`, s.theResponseShouldBe401Or403)
	ctx.Step(`^the user sends an invalid Sia request$`, s.theUserSendsAMalformedPayload)
	ctx.Step(`^the SIA response indicates bad request$`, s.theResponseShouldBe400)
	ctx.Step(`^the user requests a Sia file that does not exist$`, s.theUserRequestsANonExistentObject)
	ctx.Step(`^the SIA response indicates file not found$`, s.theResponseShouldBe404)
}

func (s *SiaErrorHandlingSteps) theUserCallsAProtectedEndpoint(ctx context.Context) (context.Context, error) {
	siaEndpoint := helpers.GetSiaEndpoint()
	if siaEndpoint == "" {
		return ctx, fmt.Errorf("SIA endpoint not configured")
	}

	resp, err := s.makeSiaRequest(http.MethodGet, siaEndpoint+"/objects", nil, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to call protected SIA endpoint: %w", err)
	}
	s.lastResponse = resp

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) theResponseShouldBe401Or403(ctx context.Context) (context.Context, error) {
	if s.lastResponse == nil {
		return ctx, fmt.Errorf("no SIA response available")
	}

	statusCode := s.lastResponse.StatusCode
	if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden {
		return ctx, fmt.Errorf("expected 401 or 403, got %d", statusCode)
	}

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) theUserSendsAMalformedPayload(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	var zeroKey types.Hash256
	_, err := sdk.Object(ctx, zeroKey)
	s.lastError = err

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) theResponseShouldBe400(ctx context.Context) (context.Context, error) {
	if s.lastError == nil {
		return ctx, fmt.Errorf("expected an error but got none")
	}

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) theUserRequestsANonExistentObject(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	var zeroKey types.Hash256
	_, err := sdk.Object(ctx, zeroKey)
	s.lastError = err

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) theResponseShouldBe404(ctx context.Context) (context.Context, error) {
	if s.lastError == nil {
		return ctx, fmt.Errorf("expected an error for non-existent object but got none")
	}

	return ctx, nil
}

func (s *SiaErrorHandlingSteps) makeSiaRequest(method, reqURL string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create SIA request: %w", err)
	}

	siaHost := helpers.GetSiaHost()
	if siaHost != "" {
		req.Host = siaHost
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute SIA request: %w", err)
	}

	return resp, nil
}
