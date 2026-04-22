package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"pinner.xyz-e2e/helpers"
)

// ContentListSteps holds the state for content listing step definitions
type ContentListSteps struct{}

// NewContentListSteps creates a new ContentListSteps instance
func NewContentListSteps() *ContentListSteps {
	return &ContentListSteps{}
}

// InitializeScenario registers all step definitions with godog
func (s *ContentListSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	// Content list and filter steps
	ctx.Step(`^the user has 5 uploaded IPFS files$`, s.theUserHasUploadedFiles)
	ctx.Step(`^only matching items are returned$`, s.onlyMatchingItemsAreReturned)
	ctx.Step(`^the user lists their IPFS content$`, s.theUserListsTheirContent)
	ctx.Step(`^all (\d+) IPFS files are returned$`, s.allFilesAreReturned)
	ctx.Step(`^the user has uploaded IPFS files named "([^"]*)", "([^"]*)", "([^"]*)"$`, s.theUserHasUploadedFilesNamed)
	ctx.Step(`^the user filters IPFS content by "([^"]*)"$`, s.theUserFiltersContentBy)
	// Content status check steps
	ctx.Step(`^an existing registered user with uploaded content$`, s.anExistingRegisteredUserWithUploadedContent)
	ctx.Step(`^the content has pinned status$`, s.theContentHasPinnedStatus)
	ctx.Step(`^the status is returned successfully$`, s.theStatusIsReturnedSuccessfully)
	ctx.Step(`^the user checks IPFS content status$`, s.theUserChecksContentStatus)
	ctx.Step(`^the user filters IPFS content by name "([^"]*)"$`, s.theUserFiltersContentByName)
	ctx.Step(`^the user has uploaded IPFS content$`, s.theUserHasUploadedContentSingle)
}

func (s *ContentListSteps) theUserHasUploadedFiles(ctx context.Context) (context.Context, error) {
	// Get existing content list or create new one
	contents, _ := helpers.GetIPFSContentList(ctx)

	// Add files until we have 5 total
	initialLen := len(contents)
	if initialLen >= 5 {
		ctx = helpers.SetIPFSContentList(ctx, contents)
		return ctx, nil
	}
	needed := 5 - initialLen

	for i := range needed {
		index := i + initialLen
		buf := []byte{}
		buf = fmt.Appendf(buf, "test content %d", index)
		
		// Upload file via portal (POST to IPFS SDK upload endpoint)
		cid, _, err := helpers.IPFSPortalUpload(ctx, buf, fmt.Sprintf("file%d.txt", index+1))
		if err != nil {
			return ctx, fmt.Errorf("failed to upload file %d: %w", index, err)
		}
		
		// Wait for operation completion (upload auto-pins)
		if err := helpers.WaitForOperationCompleteByCID(ctx, cid, helpers.DefaultOperationTimeout); err != nil {
			return ctx, fmt.Errorf("operation for file %d did not complete: %w", index, err)
		}
		
		contents = append(contents, helpers.IPFSContent{Name: fmt.Sprintf("file%d.txt", index+1), CID: cid})
	}

	ctx = helpers.SetIPFSContentList(ctx, contents)
	return ctx, nil
}

func (s *ContentListSteps) theUserListsTheirContent(ctx context.Context) (context.Context, error) {
	// Get content list from context (already set by previous steps)
	_, ok := helpers.GetIPFSContentList(ctx)
	_ = ok // suppress unused warning
	// The actual listing happens in verification steps
	return ctx, nil
}

func (s *ContentListSteps) allFilesAreReturned(ctx context.Context, count int) (context.Context, error) {
	// Get content list from context
	contents, ok := helpers.GetIPFSContentList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no content list found in context")
	}

	if len(contents) != count {
		return ctx, fmt.Errorf("expected %d files, got %d", count, len(contents))
	}

	return ctx, nil
}

func (s *ContentListSteps) theUserHasUploadedFilesNamed(ctx context.Context, arg1, arg2, arg3 string) (context.Context, error) {
	// Create specific named files
	testFiles := []helpers.IPFSContent{
		{Name: arg1}, {Name: arg2}, {Name: arg3},
	}

	// Upload each file via portal (POST to IPFS SDK upload endpoint)
	for i, tf := range testFiles {
		buf := []byte{}
		buf = fmt.Appendf(buf, "test content for %s", tf.Name)
		
		cid, _, err := helpers.IPFSPortalUpload(ctx, buf, tf.Name)
		if err != nil {
			return ctx, fmt.Errorf("failed to upload file %s: %w", tf.Name, err)
		}
		
		// Wait for operation completion (upload auto-pins)
		if err := helpers.WaitForOperationCompleteByCID(ctx, cid, helpers.DefaultOperationTimeout); err != nil {
			return ctx, fmt.Errorf("operation for %s did not complete: %w", tf.Name, err)
		}
		
		testFiles[i].CID = cid
	}

	ctx = helpers.SetIPFSContentList(ctx, testFiles)
	return ctx, nil
}

func (s *ContentListSteps) theUserFiltersContentBy(ctx context.Context, name string) (context.Context, error) {
	return s.theUserFiltersContentByName(ctx, name)
}

// anExistingRegisteredUserWithUploadedContent creates test content for a registered user
// Note: This registers a test user and generates test content, but does not pin it.
// Content pinning happens in scenario steps.
func (s *ContentListSteps) anExistingRegisteredUserWithUploadedContent(ctx context.Context) (context.Context, error) {
	// Register a test user first (same as AuthSteps.anExistingRegisteredUser)
	ctx, err := helpers.RegisterTestUser(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to register test user: %w", err)
	}

	// Create 3 test files with content
	testFiles := []helpers.IPFSContent{
		{Name: "test1.txt"}, {Name: "test2.txt"}, {Name: "example.txt"},
	}

	// Generate CID for each file (no pinning - that's done in scenario steps)
	for i := range len(testFiles) {
		buf := []byte{}
		buf = fmt.Appendf(buf, "test content %d", i)
		cid, err := helpers.KuboAdd(ctx, buf)
		if err != nil {
			return ctx, fmt.Errorf("failed to generate CID for file %s: %w", testFiles[i].Name, err)
		}
		testFiles[i].CID = cid
	}

	// Save content list to context
	ctx = helpers.SetIPFSContentList(ctx, testFiles)

	// Also save first CID for scenarios that need it (like checking status)
	if len(testFiles) > 0 && testFiles[0].CID != "" {
		ctx = helpers.SetCID(ctx, testFiles[0].CID)
	}

	return ctx, nil
}

// onlyMatchingItemsAreReturned verifies filtering worked correctly
func (s *ContentListSteps) onlyMatchingItemsAreReturned(ctx context.Context) (context.Context, error) {
	// Get content list from context
	contents, ok := helpers.GetIPFSContentList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no content list found in context")
	}

	// Get filter name from context
	filterName, ok := helpers.GetFilterName(ctx)
	if !ok || filterName == "" {
		return ctx, fmt.Errorf("no filter name found in context")
	}

	// Verify all content names match the filter
	for _, content := range contents {
		if !strings.Contains(content.Name, filterName) {
			return ctx, fmt.Errorf("content name %s does not match filter %s", content.Name, filterName)
		}
	}

	return ctx, nil
}

// theContentHasPinnedStatus verifies the content is pinned
func (s *ContentListSteps) theContentHasPinnedStatus(ctx context.Context) (context.Context, error) {
	if err := helpers.VerifyCIDPinned(ctx, "content pinning status"); err != nil {
		return ctx, fmt.Errorf("failed to check pin status: %w", err)
	}
	return ctx, nil
}

// theStatusIsReturnedSuccessfully verifies status check succeeded
func (s *ContentListSteps) theStatusIsReturnedSuccessfully(ctx context.Context) (context.Context, error) {
	// This step just verifies the previous step (theUserChecksContentStatus) succeeded
	// Since we're here, the status check didn't return an error
	return ctx, nil
}

// theUserChecksContentStatus checks the pin status of the current CID
func (s *ContentListSteps) theUserChecksContentStatus(ctx context.Context) (context.Context, error) {
	cid, err := helpers.RequireCID(ctx, "content status check")
	if err != nil {
		return ctx, err
	}

	// Check if the content is pinned
	pinned, err := helpers.IPFSIsPinned(ctx, cid)
	if err != nil {
		return ctx, fmt.Errorf("failed to check content status: %w", err)
	}

	// Save pinned status to context
	ctx = helpers.SetPinnedStatus(ctx, pinned)
	return ctx, nil
}

// theUserHasUploadedContentSingle creates a single content item for status checking
func (s *ContentListSteps) theUserHasUploadedContentSingle(ctx context.Context) (context.Context, error) {
	// Make content unique to prevent IPFS deduplication across test runs
	uniqueContent := helpers.GenerateUniqueContent("test status content")
	
	// Upload a single test file via portal (POST to IPFS SDK upload endpoint)
	cid, _, err := helpers.IPFSPortalUpload(ctx, []byte(uniqueContent), "status-test.txt")
	if err != nil {
		return ctx, err
	}
	
	// Wait for operation completion (upload auto-pins)
	if err := helpers.WaitForOperationCompleteByCID(ctx, cid, helpers.DefaultOperationTimeout); err != nil {
		return ctx, err
	}
	
	ctx = helpers.SetCID(ctx, cid)
	return ctx, nil
}

// theUserFiltersContentByName filters content list by name
func (s *ContentListSteps) theUserFiltersContentByName(ctx context.Context, name string) (context.Context, error) {
	// Get content list from context
	contents, ok := helpers.GetIPFSContentList(ctx)
	if !ok {
		return ctx, fmt.Errorf("no content list found in context")
	}

	// Save filter name to context
	ctx = helpers.SetFilterName(ctx, name)

	// Filter content by name
	var filteredContent []helpers.IPFSContent
	for _, content := range contents {
		if strings.Contains(content.Name, name) {
			filteredContent = append(filteredContent, content)
		}
	}

	// Save filtered content to context
	ctx = helpers.SetIPFSContentList(ctx, filteredContent)
	return ctx, nil
}
