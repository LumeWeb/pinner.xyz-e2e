package steps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
)

type SiaSharingSteps struct {
	lastResponse *http.Response
	lastBody     string
}

func NewSiaSharingSteps() *SiaSharingSteps {
	return &SiaSharingSteps{}
}

func (s *SiaSharingSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user has uploaded a Sia file$`, s.theUserHasUploadedASiaFile)
	ctx.Step(`^the user shares a Sia file with a valid time-limited link$`, s.theUserSharesASiaFileWithAValidTimeLimitedLink)
	ctx.Step(`^the user shares a Sia file with an expired time-limited link$`, s.theUserSharesASiaFileWithAnExpiredTimeLimitedLink)
	ctx.Step(`^the SIA shared file is accessible$`, s.theSIASharedFileIsAccessible)
	ctx.Step(`^the SIA shared file access is denied$`, s.theSIASharedFileAccessIsDenied)
}

func (s *SiaSharingSteps) theUserHasUploadedASiaFile(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = helpers.SiaUploadRandomData(ctx, 1024)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload random data to SIA: %w", err)
	}
	return ctx, nil
}

func (s *SiaSharingSteps) theUserSharesASiaFileWithAValidTimeLimitedLink(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no sia SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no sia object key found in context")
	}

	sharedURL, err := sdk.CreateSharedObjectURL(ctx, objectKey, time.Now().Add(24*time.Hour))
	if err != nil {
		return ctx, fmt.Errorf("failed to create shared object URL: %w", err)
	}

	resp, err := s.makeSharedURLRequest(sharedURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to request shared URL: %w", err)
	}
	s.lastResponse = resp

	return ctx, nil
}

func (s *SiaSharingSteps) theUserSharesASiaFileWithAnExpiredTimeLimitedLink(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no sia SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no sia object key found in context")
	}

	sharedURL, err := sdk.CreateSharedObjectURL(ctx, objectKey, time.Now().Add(-1*time.Hour))
	if err != nil {
		return ctx, fmt.Errorf("failed to create expired shared object URL: %w", err)
	}

	resp, err := s.makeSharedURLRequest(sharedURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to request expired shared URL: %w", err)
	}
	s.lastResponse = resp

	return ctx, nil
}

func (s *SiaSharingSteps) theSIASharedFileIsAccessible(ctx context.Context) (context.Context, error) {
	if s.lastResponse == nil {
		return ctx, fmt.Errorf("no SIA shared file response available")
	}

	statusCode := s.lastResponse.StatusCode
	if statusCode < 200 || statusCode >= 300 {
		return ctx, fmt.Errorf("expected 2xx status code, got %d", statusCode)
	}

	return ctx, nil
}

func (s *SiaSharingSteps) theSIASharedFileAccessIsDenied(ctx context.Context) (context.Context, error) {
	if s.lastResponse == nil {
		return ctx, fmt.Errorf("no SIA shared file response available")
	}

	statusCode := s.lastResponse.StatusCode
	if statusCode < 400 || statusCode >= 600 {
		return ctx, fmt.Errorf("expected 4xx/5xx status code, got %d", statusCode)
	}

	return ctx, nil
}

func (s *SiaSharingSteps) makeSharedURLRequest(sharedURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, sharedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	siaHost := helpers.GetSiaHost()
	if siaHost != "" {
		req.Host = siaHost
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()

	s.lastBody = strings.TrimSpace(string(bodyBytes))

	return resp, nil
}
