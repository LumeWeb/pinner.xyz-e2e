package steps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/indexd/api/app"
)

// SiaAccountSteps holds step definitions for Sia account management
type SiaAccountSteps struct {
	accountResponse app.AccountResponse
}

// NewSiaAccountSteps creates a new SiaAccountSteps instance
func NewSiaAccountSteps() *SiaAccountSteps {
	return &SiaAccountSteps{}
}

// InitializeScenario registers all Sia account step definitions with godog
func (s *SiaAccountSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user has connected their Sia account$`, s.theUserIsAuthenticated)
	ctx.Step(`^the user views their Sia account$`, s.theUserRequestsAccountDetails)
	ctx.Step(`^the Sia account shows usage and quota information$`, s.theResponseContainsQuotaAndUsageInformation)
	ctx.Step(`^a new Sia user registration$`, s.aTestUserRegistrationFlow)
	ctx.Step(`^the Sia user signs in$`, s.theTestUserLogsIn)
	ctx.Step(`^the SIA account summary is displayed$`, s.theAccountSummaryIsReturned)
}

func (s *SiaAccountSteps) theUserIsAuthenticated(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.RegisterAndLoginTestUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to register and login test user: %w", err)
	}

	ctx, err = helpers.VerifyTestUserEmail(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify test user email: %w", err)
	}

	ctx, err = helpers.SiaSDKConnect(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to SIA: %w", err)
	}

	return ctx, nil
}

// theUserRequestsAccountDetails retrieves account details from the Sia indexer
func (s *SiaAccountSteps) theUserRequestsAccountDetails(ctx context.Context) (context.Context, error) {
	account, err := helpers.SiaGetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get SIA account details: %w", err)
	}

	s.accountResponse = account
	return ctx, nil
}

// theResponseContainsQuotaAndUsageInformation verifies the account response has quota and usage fields
func (s *SiaAccountSteps) theResponseContainsQuotaAndUsageInformation(ctx context.Context) (context.Context, error) {
	if s.accountResponse == (app.AccountResponse{}) {
		return ctx, fmt.Errorf("SIA account response is empty, expected quota and usage information")
	}

	return ctx, nil
}

func (s *SiaAccountSteps) aTestUserRegistrationFlow(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.RegisterAndLoginTestUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to register and login test user: %w", err)
	}

	ctx, err = helpers.VerifyTestUserEmail(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to verify test user email: %w", err)
	}

	ctx, err = helpers.SiaSDKConnect(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to connect to SIA during registration flow: %w", err)
	}

	return ctx, nil
}

// theTestUserLogsIn authenticates the test user and reconnects to Sia
func (s *SiaAccountSteps) theTestUserLogsIn(ctx context.Context) (context.Context, error) {
	ctx, err := helpers.LoginTestUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to login test user: %w", err)
	}

	// Reconnect to Sia using the stored app key if available
	if _, ok := helpers.GetSiaAppKey(ctx); ok {
		ctx, err = helpers.SiaSDKFromAppKey(ctx)
		if err != nil {
			return ctx, fmt.Errorf("failed to reconnect to SIA after login: %w", err)
		}
	}

	return ctx, nil
}

// theAccountSummaryIsReturned verifies that account details can be retrieved
func (s *SiaAccountSteps) theAccountSummaryIsReturned(ctx context.Context) (context.Context, error) {
	account, err := helpers.SiaGetAccount(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to get SIA account summary: %w", err)
	}

	s.accountResponse = account

	if s.accountResponse == (app.AccountResponse{}) {
		return ctx, fmt.Errorf("SIA account summary is empty")
	}

	return ctx, nil
}

// SiaConnectUISteps holds step definitions for Sia connect UI page interactions
type SiaConnectUISteps struct{}

// NewSiaConnectUISteps creates a new SiaConnectUISteps instance
func NewSiaConnectUISteps() *SiaConnectUISteps {
	return &SiaConnectUISteps{}
}

// InitializeScenario registers all Sia connect UI step definitions with godog
func (s *SiaConnectUISteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^an unauthenticated request is made to the SIA connect page$`, s.anUnauthenticatedRequestIsMadeToTheSIAConnectPage)
	ctx.Step(`^the response is a redirect to the login page with a return URL$`, s.theResponseIsARedirectToTheLoginPageWithAReturnURL)
	ctx.Step(`^an authenticated request is made to the SIA connect page$`, s.anAuthenticatedRequestIsMadeToTheSIAConnectPage)
	ctx.Step(`^the response is an HTML page with approval controls$`, s.theResponseIsAnHTMLPageWithApprovalControls)
}

func (s *SiaConnectUISteps) anUnauthenticatedRequestIsMadeToTheSIAConnectPage(ctx context.Context) (context.Context, error) {
	connectURL, _ := helpers.GetSiaConnectRequestID(ctx)
	requestID, err := helpers.ParseRequestIDFromConnectURL(connectURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse SIA connect request ID: %w", err)
	}

	pageURL := fmt.Sprintf("%s/auth/connect/%s", helpers.GetSiaEndpoint(), requestID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to create request: %w", err)
	}
	req.Host = helpers.GetSiaHost()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send request: %w", err)
	}

	ctx = helpers.SetSiaConnectUIResponse(ctx, resp)
	return ctx, nil
}

func (s *SiaConnectUISteps) theResponseIsARedirectToTheLoginPageWithAReturnURL(ctx context.Context) (context.Context, error) {
	resp, ok := helpers.GetSiaConnectUIResponse(ctx)
	if !ok || resp == nil {
		return ctx, fmt.Errorf("no SIA connect UI response found in context")
	}

	if resp.StatusCode != http.StatusFound {
		return ctx, fmt.Errorf("expected status 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return ctx, fmt.Errorf("expected Location header in redirect response")
	}

	if !strings.Contains(location, "account.") {
		return ctx, fmt.Errorf("expected Location to contain 'account.' subdomain, got %s", location)
	}

	if !strings.Contains(location, "return=") {
		return ctx, fmt.Errorf("expected Location to contain 'return=' parameter, got %s", location)
	}

	return ctx, nil
}

func (s *SiaConnectUISteps) anAuthenticatedRequestIsMadeToTheSIAConnectPage(ctx context.Context) (context.Context, error) {
	connectURL, _ := helpers.GetSiaConnectRequestID(ctx)
	requestID, err := helpers.ParseRequestIDFromConnectURL(connectURL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse SIA connect request ID: %w", err)
	}

	token, ok := helpers.GetJWTToken(ctx)
	if !ok || token == "" {
		return ctx, fmt.Errorf("no JWT token available for authenticated SIA connect page request")
	}

	pageURL := fmt.Sprintf("%s/auth/connect/%s", helpers.GetSiaEndpoint(), requestID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to create request: %w", err)
	}
	req.Host = helpers.GetSiaHost()
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send request: %w", err)
	}

	ctx = helpers.SetSiaConnectUIResponse(ctx, resp)
	return ctx, nil
}

func (s *SiaConnectUISteps) theResponseIsAnHTMLPageWithApprovalControls(ctx context.Context) (context.Context, error) {
	resp, ok := helpers.GetSiaConnectUIResponse(ctx)
	if !ok || resp == nil {
		return ctx, fmt.Errorf("no SIA connect UI response found in context")
	}

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return ctx, fmt.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read response body: %w", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, `data-action="approve"`) {
		return ctx, fmt.Errorf("expected HTML to contain approval controls (data-action=\"approve\")")
	}

	if !strings.Contains(bodyStr, `data-action="reject"`) {
		return ctx, fmt.Errorf("expected HTML to contain rejection controls (data-action=\"reject\")")
	}

	return ctx, nil
}
