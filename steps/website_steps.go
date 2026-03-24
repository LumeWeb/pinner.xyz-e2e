package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.lumeweb.com/ipfs-sdk"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// WebsiteDefaultTestDomain is the standard test domain used in website feature files
	WebsiteDefaultTestDomain = "test-website.example.com"
)

// WebsiteSteps provides step definitions for website CRUD operations
type WebsiteSteps struct{}

// NewWebsiteSteps creates a new WebsiteSteps instance
func NewWebsiteSteps() *WebsiteSteps {
	return &WebsiteSteps{}
}

// InitializeScenario registers website CRUD operation steps with godog
func (s *WebsiteSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// CRUD operations
	ctx.Step(`^the user creates a website with domain "([^"]*)"$`, s.theUserCreatesAWebsiteWithDomain)
	ctx.Step(`^the user creates a website with domain "([^"]*)" and target hash "([^"]*)"$`, s.theUserCreatesAWebsiteWithDomainAndTargetHash)
	ctx.Step(`^the user creates a website with DNS hosting enabled$`, s.theUserCreatesAWebsiteWithDnsHostingEnabled)
	ctx.Step(`^the user has a website$`, s.theUserHasAWebsite)
	ctx.Step(`^the user has an IPNS website$`, s.theUserHasAnIPNSWebsite)
	ctx.Step(`^we upload a website bundle$`, s.weUploadAWebsiteBundle)
	ctx.Step(`^the upload operation completes$`, s.theUploadOperationCompletes)
	ctx.Step(`^the user creates a pure IPFS website with domain "([^"]*)"$`, s.theUserCreatesPureIPFSWebsite)
	ctx.Step(`^the user creates a website with domain "([^"]*)" using the uploaded CID$`, s.theUserCreatesAWebsiteUsingTheUploadedCID)
	ctx.Step(`^the user creates a website with domain "([^"]*)" using the IPFS CID$`, s.theUserCreatesAWebsiteUsingTheIPFSCID)
	ctx.Step(`^the user creates a website with domain "([^"]*)" using the IPNS peer ID$`, s.theUserCreatesAWebsiteUsingTheIPNSPeerID)
	ctx.Step(`^the user has an IPNS key$`, s.theUserHasAnIPNSKey)
	ctx.Step(`^the user gets the website by ID$`, s.theUserGetsTheWebsiteByID)
	ctx.Step(`^the user lists websites$`, s.theUserListsWebsites)
	ctx.Step(`^the user updates the website$`, s.theUserUpdatesTheWebsite)
	ctx.Step(`^the user updates the website to use IPNS target "([^"]*)"$`, s.theUserUpdatesTheWebsiteToUseIPNSTarget)
	ctx.Step(`^the user deletes the website$`, s.theUserDeletesTheWebsite)
	ctx.Step(`^the user checks SSL status for domain "([^"]*)"$`, s.theUserChecksSSLStatusForDomain)
	
	// IPNS integration steps
	ctx.Step(`^the user creates an IPNS key$`, s.theUserCreatesAnIPNSKey)
	ctx.Step(`^the user publishes new content to the IPNS key$`, s.theUserPublishesNewContentToIPNSKey)
	ctx.Step(`^the user publishes the IPFS CID to the IPNS key$`, s.theUserPublishesIPFSCIDToIPNSKey)
	ctx.Step(`^the user publishes the uploaded CID to the IPNS key$`, s.theUserPublishesTheUploadedCIDToIPNSKey)
	ctx.Step(`^the user publishes the uploaded CID to the existing IPNS key$`, s.theUserPublishesTheUploadedCIDToIPNSKey)
	ctx.Step(`^the user publishes the website CID to the IPNS key$`, s.theUserPublishesWebsiteCIDToIPNSKey)
	ctx.Step(`^the user updates the website to use the IPNS target$`, s.theUserUpdatesWebsiteToUseIPNSTarget)
	ctx.Step(`^the website target hash is the IPNS name$`, s.theWebsiteTargetHashIsIPNSName)
	ctx.Step(`^the website target hash is updated$`, s.theWebsiteTargetHashUpdated)
	ctx.Step(`^the website target type is still "([^"]*)"$`, s.theWebsiteTargetTypeIsStill)
	ctx.Step(`^the website has an IPNS target$`, s.theWebsiteHasAnIPNSTarget)
}

// theWebsiteHasAnIPNSTarget verifies that the website has a valid IPNS target hash
// theWebsiteHasAnIPNSTarget verifies the website has a valid IPNS peer ID
// Uses libp2p peer library to validate peer IDs properly
func (s *WebsiteSteps) theWebsiteHasAnIPNSTarget(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	// Get the website from API
	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	// Verify target hash is a valid libp2p peer ID
	targetHash := website.TargetHash
	_, err = peer.Decode(targetHash)
	if err != nil {
		return fmt.Errorf("target hash '%s' is not a valid libp2p peer ID: %w", targetHash, err)
	}

	return nil
}

// theUserCreatesAnIPNSKey creates a new IPNS key for website operations
// Action step - user creates an IPNS key through the API
func (s *WebsiteSteps) theUserCreatesAnIPNSKey(ctx context.Context) (context.Context, error) {
	keyName := "website-test-key"
	
	ctx, ipnsKey, err := helpers.CreateIPNSKey(ctx, keyName)
	if err != nil {
		return ctx, fmt.Errorf("failed to create IPNS key: %w", err)
	}
	
	
	// Store IPNS key details for cleanup and verification
	ctx = helpers.SetIPNSKeyID(ctx, ipnsKey.Id)
	ctx = helpers.SetIPNSKeyName(ctx, keyName)
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsKey.Name)
	
	return ctx, nil
}

// theUserPublishesNewContentToIPNSKey publishes new content to the IPNS key
// This simulates a user updating their website content without changing the website configuration
func (s *WebsiteSteps) theUserPublishesNewContentToIPNSKey(ctx context.Context) (context.Context, error) {
	
	// Get the IPNS key ID from context
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	// Upload new content to get a new CID
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	newCID, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload new content: %w", err)
	}

	// Store upload result in context
	ctx = helpers.SetUploadResult(ctx, &helpers.UploadResult{CID: newCID})
	
	// Wait for upload to complete
	err = helpers.WaitForOperation(ctx, newCID)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}

	// Publish the new CID to the IPNS key
	ipnsResponse, err := helpers.PublishToIPNS(ctx, keyID, newCID)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish content to IPNS key: %w", err)
	}

	
	// Store the new IPNS name for verification
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsResponse.Name)
	
	// Wait for IPNS record to be available before continuing
	if err := helpers.WaitForIPNSPublish(ctx, keyID, newCID); err != nil {
		return ctx, fmt.Errorf("failed to wait for IPNS record: %w", err)
	}
	
	return ctx, nil
}

// theUserPublishesIPFSCIDToIPNSKey publishes an IPFS CID to an IPNS key
// Used when converting a website from IPFS to IPNS target
func (s *WebsiteSteps) theUserPublishesIPFSCIDToIPNSKey(ctx context.Context) (context.Context, error) {
	
	// Get the IPNS key ID from context
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	// Get the website's current target hash (which is an IPFS CID)
	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target hash found in context")
	}

	// Publish the CID to the IPNS key
	ipnsResponse, err := helpers.PublishToIPNS(ctx, keyID, targetHash)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish CID to IPNS key: %w", err)
	}

	
	// Store the IPNS name for verification
	// Note: Don't overwrite website target hash here - it should remain the CID
	// until the website is explicitly updated in a later step
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsResponse.Name)
	
	// Wait for IPNS record to be available before continuing
	if err := helpers.WaitForIPNSPublish(ctx, keyID, targetHash); err != nil {
		return ctx, fmt.Errorf("failed to wait for IPNS record: %w", err)
	}
	
	return ctx, nil
}

// theUserPublishesTheUploadedCIDToIPNSKey publishes the uploaded CID to an existing IPNS key
// Gets CID from the upload result in context
// Uses existing IPNS key ID from context (requires context to be set up first)
func (s *WebsiteSteps) theUserPublishesTheUploadedCIDToIPNSKey(ctx context.Context) (context.Context, error) {
	
	// Get the IPNS key ID from context (must exist from prior setup)
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no existing IPNS key ID found - ensure IPNS key setup step was called first")
	}

	// Get the uploaded CID from context
	uploadResult, ok := helpers.GetUploadResult(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload result found in context")
	}
	cid := uploadResult.CID

	// Publish the CID to the IPNS key
	ipnsResponse, err := helpers.PublishToIPNS(ctx, keyID, cid)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish CID to IPNS key: %w", err)
	}

	
	// Update the IPNS name in context
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsResponse.Name)
	
	// Wait for IPNS record to be available before continuing
	if err := helpers.WaitForIPNSPublish(ctx, keyID, cid); err != nil {
		return ctx, fmt.Errorf("failed to wait for IPNS record: %w", err)
	}
	
	return ctx, nil
}

// theUserPublishesWebsiteCIDToIPNSKey publishes the website's current CID to an IPNS key
// Used when converting a website from IPFS to IPNS target
// Gets the website's current target hash (which is an IPFS CID) and publishes it to IPNS
func (s *WebsiteSteps) theUserPublishesWebsiteCIDToIPNSKey(ctx context.Context) (context.Context, error) {
	
	// Get the IPNS key ID from context
	keyID, ok := helpers.GetIPNSKeyID(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS key ID found in context")
	}

	// Get the website's current target hash (which is an IPFS CID)
	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target hash found in context")
	}

	// Publish the CID to the IPNS key
	ipnsResponse, err := helpers.PublishToIPNS(ctx, keyID, targetHash)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish CID to IPNS key: %w", err)
	}

	
	// Store the IPNS name for verification
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsResponse.Name)
	
	// Wait for IPNS record to be available before continuing
	// This ensures the janitor can find the record when it validates
	if err := helpers.WaitForIPNSPublish(ctx, keyID, targetHash); err != nil {
		return ctx, fmt.Errorf("failed to wait for IPNS record: %w", err)
	}
	
	return ctx, nil
}

// theWebsiteTargetHashIsIPNSName verifies the website target hash is an IPNS name
// This checks observable API response, not internal context
func (s *WebsiteSteps) theWebsiteTargetHashIsIPNSName(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	// Get IPNS name from context (we stored it there during publish)
	expectedIPNSName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return fmt.Errorf("no IPNS name found in context")
	}

	// Verify the website's target hash matches the IPNS name
	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	if website.TargetHash != expectedIPNSName {
		return fmt.Errorf("expected target hash '%s', got '%s'", expectedIPNSName, website.TargetHash)
	}

	return nil
}

// theWebsiteTargetHashUpdated verifies the website target hash was updated
// This checks the API response shows a different hash than before
func (s *WebsiteSteps) theWebsiteTargetHashUpdated(ctx context.Context) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	// Get updated website from API
	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	// Get the new IPNS name from context
	newIPNSName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return fmt.Errorf("no new IPNS name found in context")
	}

	// Verify the target hash is the new IPNS name
	if website.TargetHash != newIPNSName {
		return fmt.Errorf("expected target hash to be updated to '%s', got '%s'", newIPNSName, website.TargetHash)
	}

	return nil
}

// theWebsiteTargetTypeIsStill verifies the website target type remains unchanged
// This checks observable API response
func (s *WebsiteSteps) theWebsiteTargetTypeIsStill(ctx context.Context, expectedType string) error {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return err
	}

	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return fmt.Errorf("website is nil")
	}

	if website.TargetType != expectedType {
		return fmt.Errorf("expected target type '%s', got '%s'", expectedType, website.TargetType)
	}

	return nil
}

// theUserCreatesAWebsiteWithDomain creates a website with the specified domain
// Uses IPFS target type by default
func (s *WebsiteSteps) theUserCreatesAWebsiteWithDomain(ctx context.Context, domain string) (context.Context, error) {
	// Upload the sample website directory to IPFS using IPFS SDK upload API
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}
	
	// Wait for the upload operation to complete
	err = helpers.WaitForOperation(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}
	
	// Create the website
	return s.createWebsiteInternal(ctx, domain, cid, "ipfs", false)
}

// theUserCreatesAWebsiteWithDomainAndTargetHash creates a website with specified domain and target hash
func (s *WebsiteSteps) theUserCreatesAWebsiteWithDomainAndTargetHash(ctx context.Context, domain, targetHash string) (context.Context, error) {
	// For default test domain, use a unique domain
	if domain == WebsiteDefaultTestDomain {
		domain = helpers.GenerateTestWebsiteDomain()
	}
	
	// Create the website with IPFS target type
	return s.createWebsiteInternal(ctx, domain, targetHash, "ipfs", false)
}

// theUserCreatesAWebsiteWithDnsHostingEnabled creates a website with DNS hosting enabled
func (s *WebsiteSteps) theUserCreatesAWebsiteWithDnsHostingEnabled(ctx context.Context) (context.Context, error) {
	domain := helpers.GenerateTestWebsiteDomain()
	
	// Upload the sample website directory to IPFS using IPFS SDK upload API
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}
	
	// Wait for the upload operation to complete
	err = helpers.WaitForOperation(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}
	
	// Create the website with DNS hosting enabled
	return s.createWebsiteInternal(ctx, domain, cid, "ipfs", true)
}

// theUserHasAWebsite creates a website as a precondition for scenarios
// Uses default IPFS target type
func (s *WebsiteSteps) theUserHasAWebsite(ctx context.Context) (context.Context, error) {
	domain := helpers.GenerateTestWebsiteDomain()
	
	// Upload the sample website directory to IPFS using IPFS SDK upload API
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}
	
	// Wait for the upload operation to complete
	err = helpers.WaitForOperation(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}
	
	return s.createWebsiteInternal(ctx, domain, cid, "ipfs", false)
}

// theUserHasAnIPNSWebsite creates a website with IPNS target as a precondition
// Can operate in two modes:
// 1. Existing IPNS data in context (from prior publish operations) - creates website immediately
// 2. No IPNS data in context - performs full setup (upload → IPNS key → publish → website)
func (s *WebsiteSteps) theUserHasAnIPNSWebsite(ctx context.Context) (context.Context, error) {
	
	// Check if we already have IPNS setup in context
	ipnsName, hasIPNS := helpers.GetIPNSIPNSName(ctx)
	_, hasKey := helpers.GetIPNSKeyID(ctx)
	keyName, hasName := helpers.GetIPNSKeyName(ctx)
	
	if hasIPNS && hasKey {
		// Mode 1: Reuse existing IPNS setup from context
		
		domain := helpers.GenerateTestWebsiteDomain()
		
		// Create the website with IPNS target type
		
		result, err := s.createWebsiteInternal(ctx, domain, ipnsName, "ipns", false)
		if err != nil {
			return result, err
		}
		
		return result, nil
	}
	
	// Mode 2: Full setup - upload → key → publish → website
	domain := helpers.GenerateTestWebsiteDomain()
	
	// Create IPNS key
	actualKeyName := "website-test-key"
	if hasName {
		actualKeyName = keyName
	}
	ipnsKeyCtx, ipnsKey, err := helpers.CreateIPNSKey(ctx, actualKeyName)
	if err != nil {
		return ctx, fmt.Errorf("failed to create IPNS key: %w", err)
	}
	ctx = ipnsKeyCtx
	
	// Upload the sample website directory to IPFS
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}
	
	// Store upload result for operation complete step
	ctx = helpers.SetUploadResult(ctx, &helpers.UploadResult{CID: cid})
	
	// Publish CID to the IPNS key
	ipnsResponse, err := helpers.PublishToIPNS(ctx, ipnsKey.Id, cid)
	if err != nil {
		return ctx, fmt.Errorf("failed to publish IPNS content: %w", err)
	}
	
	// Store IPNS key details for cleanup
	ctx = helpers.SetIPNSKeyID(ctx, ipnsKey.Id)
	ctx = helpers.SetIPNSKeyName(ctx, actualKeyName)
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsResponse.Name)
	
	// Create the website with IPNS target type
	
	result, err := s.createWebsiteInternal(ctx, domain, ipnsResponse.Name, "ipns", false)
	if err != nil {
		return result, err
	}
	
	return result, nil
}

// theUserGetsTheWebsiteByID retrieves a specific website by ID
func (s *WebsiteSteps) theUserGetsTheWebsiteByID(ctx context.Context) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, fmt.Errorf("no website ID in context: %w", err)
	}

	idStr := fmt.Sprintf("%d", id)
	website, err := websiteService.Get(ctx, idStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to get website: %w", err)
	}

	if website == nil {
		return ctx, fmt.Errorf("website is nil")
	}

	// Store website details in context
	domain := website.Domain
	ctx = helpers.StoreWebsiteID(ctx, website.Id)
	ctx = helpers.SetWebsiteDomain(ctx, domain)
	ctx = helpers.SetWebsiteTargetHash(ctx, website.TargetHash)
	ctx = helpers.SetWebsiteTargetType(ctx, website.TargetType)
	ctx = helpers.SetWebsiteStatus(ctx, website.Status)

	return ctx, nil
}

// theUserListsWebsites lists all websites for the authenticated user
func (s *WebsiteSteps) theUserListsWebsites(ctx context.Context) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	websites, err := websiteService.List(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to list websites: %w", err)
	}

	if websites == nil {
		return ctx, fmt.Errorf("website list is nil")
	}

	// Store website count in context for verification
	websiteCount := len(websites)
	ctx = helpers.SetListCount(ctx, helpers.WebsiteListKey, websiteCount)

	return ctx, nil



}

// theUserUpdatesTheWebsite updates an existing website
func (s *WebsiteSteps) theUserUpdatesTheWebsite(ctx context.Context) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	targetHash, ok := helpers.GetWebsiteTargetHash(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target hash in context")
	}

	targetType, ok := helpers.GetWebsiteTargetType(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website target type in context")
	}

	idStr := fmt.Sprintf("%d", id)
	updatedWebsite, err := websiteService.Update(ctx, idStr, domain, targetHash, targetType)
	if err != nil {
		return ctx, fmt.Errorf("failed to update website: %w", err)
	}

	// Store updated website details in context
	ctx = helpers.SetWebsiteDomain(ctx, updatedWebsite.Domain)
	ctx = helpers.SetWebsiteTargetHash(ctx, updatedWebsite.TargetHash)
	ctx = helpers.SetWebsiteTargetType(ctx, updatedWebsite.TargetType)
	ctx = helpers.SetWebsiteStatus(ctx, updatedWebsite.Status)

	return ctx, nil
}

// theUserUpdatesTheWebsiteToUseIPNSTarget updates a website to use an IPNS target
func (s *WebsiteSteps) theUserUpdatesTheWebsiteToUseIPNSTarget(ctx context.Context, ipnsName string) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	domain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	idStr := fmt.Sprintf("%d", id)
	updatedWebsite, err := websiteService.Update(ctx, idStr, domain, ipnsName, "ipns")
	if err != nil {
		return ctx, fmt.Errorf("failed to update website to IPNS target: %w", err)
	}

	// Store updated details in context
	ctx = helpers.SetWebsiteTargetHash(ctx, updatedWebsite.TargetHash)
	ctx = helpers.SetWebsiteTargetType(ctx, updatedWebsite.TargetType)

	return ctx, nil
}

// theUserUpdatesWebsiteToUseIPNSTarget updates a website to use IPNS target
// Gets the IPNS name from context (set during IPNS publish operations)
func (s *WebsiteSteps) theUserUpdatesWebsiteToUseIPNSTarget(ctx context.Context) (context.Context, error) {
	
	// Get IPNS name from context (published during previous step)
	ipnsName, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS name found in context")
	}

	// Call the parameterized version with the IPNS name
	return s.theUserUpdatesTheWebsiteToUseIPNSTarget(ctx, ipnsName)
}

// theUserDeletesTheWebsite deletes a website
func (s *WebsiteSteps) theUserDeletesTheWebsite(ctx context.Context) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	id, err := helpers.RequireWebsiteID(ctx)
	if err != nil {
		return ctx, err
	}

	idStr := fmt.Sprintf("%d", id)
	err = websiteService.Delete(ctx, idStr)
	if err != nil {
		return ctx, fmt.Errorf("failed to delete website: %w", err)
	}

	return ctx, nil
}

// theUserChecksSSLStatusForDomain checks the SSL status for a domain
// Note: The domain parameter is ignored in favor of using the actual website domain from context
// This is because test websites use randomized domains to avoid conflicts
func (s *WebsiteSteps) theUserChecksSSLStatusForDomain(ctx context.Context, _ string) (context.Context, error) {
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	// Get the actual domain from context (created during website setup)
	actualDomain, ok := helpers.GetWebsiteDomain(ctx)
	if !ok {
		return ctx, fmt.Errorf("no website domain in context")
	}

	sslStatus, err := websiteService.GetSSLStatus(ctx, actualDomain)
	if err != nil {
		return ctx, fmt.Errorf("failed to get SSL status: %w", err)
	}

	if sslStatus.Ssl != nil {
		ctx = helpers.SetWebsiteSslStatus(ctx, sslStatus.Ssl.Status)
	}

	return ctx, nil
}

// createWebsiteInternal is an internal helper for creating websites
// It handles the common website creation logic and stores context values
func (s *WebsiteSteps) createWebsiteInternal(ctx context.Context, domain, targetHash, targetType string, dnsHostingEnabled bool) (context.Context, error) {
	
	websiteService, err := helpers.RequireWebsiteService(ctx)
	if err != nil {
		return ctx, err
	}

	website, err := websiteService.CreateWithOptions(ctx, ipfs.WebsiteRequest{
		Domain:            domain,
		TargetHash:       targetHash,
		TargetType:       targetType,
		DnsHostingEnabled: &dnsHostingEnabled,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to create website: %w", err)
	}

	// Store website details in context for verification and cleanup
	ctx = helpers.StoreWebsiteID(ctx, website.Id)
	ctx = helpers.SetWebsiteDomain(ctx, website.Domain)
	ctx = helpers.SetWebsiteTargetHash(ctx, website.TargetHash)
	ctx = helpers.SetWebsiteTargetType(ctx, website.TargetType)
	ctx = helpers.SetWebsiteStatus(ctx, website.Status)
	ctx = helpers.SetWebsiteDnsHostingEnabled(ctx, website.DnsHostingEnabled)

	// Add DNS records to dev server for DNS validation
	// DNS validation is required regardless of DNS hosting status to prove domain ownership.
	// The dnsHostingEnabled flag only controls whether the portal manages DNS, not whether
	// validation is required.
	dnslink := helpers.BuildWebsiteDNSLink(website.TargetHash, website.TargetType)
	err = helpers.AddWebsiteDNSZoneToDevServer(ctx, website.Domain, dnslink, website.ValidationToken)
	if err != nil {
		return ctx, fmt.Errorf("failed to add DNS zone to dev server: %w", err)
	}

	// Add to cleanup list
	ctx = helpers.AddWebsiteIDCleanup(ctx, website.Id)

	return ctx, nil
}

// theUserCreatesAWebsiteUsingTheUploadedCID creates a website using a CID from a prior upload
// This tests the scenario where the user wants to use a specific uploaded CID as website target
func (s *WebsiteSteps) theUserCreatesAWebsiteUsingTheUploadedCID(ctx context.Context, intendedDomain string) (context.Context, error) {
	
	// Generate a unique domain for API calls to avoid duplicates
	actualDomain := helpers.GenerateTestWebsiteDomain()
	
	// Store both intended domain (for verification) and actual domain (used in API)
	ctx = helpers.SetWebsiteIntendedDomain(ctx, intendedDomain)
	ctx = helpers.SetWebsiteActualDomain(ctx, actualDomain)
	
	// Get the uploaded CID from context (from the Given step)
	uploadResult, ok := helpers.GetUploadResult(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload result found in context")
	}
	
	cid := uploadResult.CID
	
	// Store the intended target hash (uploaded CID) for verification
	ctx = helpers.SetWebsiteIntendedTargetHash(ctx, cid)
	
	
	// Create the website with the actual domain (randomized)
	// API will convert the CID to an IPNS peer ID
	return s.createWebsiteInternal(ctx, actualDomain, cid, "ipfs", false)
}

// weUploadAWebsiteBundle uploads the sample website fixture to IPFS
// This initiates the upload but does not wait for completion
// Completion should be verified with "the upload operation completes" step
func (s *WebsiteSteps) weUploadAWebsiteBundle(ctx context.Context) (context.Context, error) {
	
	// Upload the sample website directory to IPFS using IPFS SDK upload API
	fixturePath := helpers.GetSampleWebsiteFixturePath()
	cid, err := helpers.IPFSPortalUploadDirFromFS(ctx, fixturePath)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload sample website: %w", err)
	}

	// Store uploaded CID for verification in Then step
	ctx = helpers.SetUploadResult(ctx, &helpers.UploadResult{CID: cid})

	return ctx, nil
}

// theUploadOperationCompletes waits for the upload operation to complete
func (s *WebsiteSteps) theUploadOperationCompletes(ctx context.Context) (context.Context, error) {
	uploadResult, ok := helpers.GetUploadResult(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload result found in context")
	}

	cid := uploadResult.CID
	err := helpers.WaitForOperation(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("operation failed to complete: %w", err)
	}

	ctx = helpers.SetContextValue(ctx, helpers.UploadOperationCompletedKey, true)
	return ctx, nil
}

// theUserCreatesAWebsiteUsingTheIPFSCID creates a website using an IPFS CID from a prior upload
// This is the scenario where we want DNS hosting enabled (auto-converts to IPNS)
func (s *WebsiteSteps) theUserCreatesAWebsiteUsingTheIPFSCID(ctx context.Context, intendedDomain string) (context.Context, error) {
	// Generate a unique domain for API calls to avoid duplicates
	actualDomain := helpers.GenerateTestWebsiteDomain()

	// Store both intended domain (for verification) and actual domain (used in API)
	ctx = helpers.SetWebsiteIntendedDomain(ctx, intendedDomain)
	ctx = helpers.SetWebsiteActualDomain(ctx, actualDomain)

	// Get the uploaded CID from context (from the Given step)
	uploadResult, ok := helpers.GetUploadResult(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload result found in context")
	}

	cid := uploadResult.CID

	// Store the intended target hash (uploaded CID) for verification
	ctx = helpers.SetWebsiteIntendedTargetHash(ctx, cid)

	// Create the website with the actual domain (randomized)
	// DNS hosting enabled = auto-converts CID to IPNS peer ID
	return s.createWebsiteInternal(ctx, actualDomain, cid, "ipfs", true)
}

// theUserCreatesPureIPFSWebsite creates a website with DNS hosting disabled (pure IPFS mode)
func (s *WebsiteSteps) theUserCreatesPureIPFSWebsite(ctx context.Context, intendedDomain string) (context.Context, error) {
	// Generate a unique domain for API calls to avoid duplicates
	actualDomain := helpers.GenerateTestWebsiteDomain()

	// Store both intended domain (for verification) and actual domain (used in API)
	ctx = helpers.SetWebsiteIntendedDomain(ctx, intendedDomain)
	ctx = helpers.SetWebsiteActualDomain(ctx, actualDomain)

	// Get the uploaded CID from context (from the Given step)
	uploadResult, ok := helpers.GetUploadResult(ctx)
	if !ok {
		return ctx, fmt.Errorf("no upload result found in context")
	}

	cid := uploadResult.CID

	// Store the intended target hash (uploaded CID) for verification
	ctx = helpers.SetWebsiteIntendedTargetHash(ctx, cid)

	// Create the website with the actual domain (randomized)
	// DNS hosting disabled = stays as pure IPFS (no auto-conversion)
	return s.createWebsiteInternal(ctx, actualDomain, cid, "ipfs", false)
}

// theUserCreatesAWebsiteUsingTheIPNSPeerID creates a website using an IPNS peer ID
func (s *WebsiteSteps) theUserCreatesAWebsiteUsingTheIPNSPeerID(ctx context.Context, intendedDomain string) (context.Context, error) {
	// Generate a unique domain for API calls to avoid duplicates
	actualDomain := helpers.GenerateTestWebsiteDomain()

	// Store both intended domain (for verification) and actual domain (used in API)
	ctx = helpers.SetWebsiteIntendedDomain(ctx, intendedDomain)
	ctx = helpers.SetWebsiteActualDomain(ctx, actualDomain)

	// Get the IPNS peer ID from context (from the Given step)
	ipnsPeerID, ok := helpers.GetIPNSIPNSName(ctx)
	if !ok {
		return ctx, fmt.Errorf("no IPNS peer ID found in context")
	}

	// Store the intended target hash (IPNS peer ID) for verification
	ctx = helpers.SetWebsiteIntendedTargetHash(ctx, ipnsPeerID)

	// Create the website with the actual domain (randomized)
	// Use IPNS target type
	return s.createWebsiteInternal(ctx, actualDomain, ipnsPeerID, "ipns", false)
}

// theUserHasAnIPNSKey creates an IPNS key as a precondition for scenarios
func (s *WebsiteSteps) theUserHasAnIPNSKey(ctx context.Context) (context.Context, error) {
	keyName := "website-test-key"
	ipnsKeyCtx, ipnsKey, err := helpers.CreateIPNSKey(ctx, keyName)
	if err != nil {
		return ctx, fmt.Errorf("failed to create IPNS key: %w", err)
	}

	// Store IPNS key details for cleanup and verification
	ctx = helpers.SetIPNSKeyID(ctx, ipnsKey.Id)
	ctx = helpers.SetIPNSKeyName(ctx, keyName)
	ctx = helpers.SetIPNSIPNSName(ctx, ipnsKey.Name)

	return ipnsKeyCtx, nil
}
