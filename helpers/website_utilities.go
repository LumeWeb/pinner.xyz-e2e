package helpers

import (
	"context"
	"fmt"
	"strings"
)

// CleanupWebsites removes all websites tracked in the cleanup list
// Returns an error only for actual failures, not "website not found" (which indicates successful prior deletion)
func CleanupWebsites(ctx context.Context) error {
	websiteIDs := GetWebsitesCleanup(ctx)
	if len(websiteIDs) == 0 {
		return nil
	}

	websiteService, err := GetWebsiteService(ctx)
	if err != nil {
		return err
	}

	var cleanupErrors []error
	for _, id := range websiteIDs {
		idStr := fmt.Sprintf("%d", id)
		err := websiteService.Delete(ctx, idStr)
		// "website not found" errors are not actual failures - the website was already deleted
		// Check for variants of this error message in the error string
		if err != nil {
			errStr := err.Error()
			// Check if error message contains "website not found" (handling API prefixes)
			if isWebsiteNotFoundError(errStr) {
				// Website was already deleted - this is successful cleanup, continue to next
				continue
			}
			// Other errors are actual failures - collect and continue
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to delete website %d: %w", id, err))
		}
	}
	if len(cleanupErrors) > 0 {
		return fmt.Errorf("website cleanup completed with %d errors: %v", len(cleanupErrors), cleanupErrors)
	}

	return nil
}

// isWebsiteNotFoundError checks if the error message indicates a "website not found" error
// This handles various API response formats and prefixes that may be added
func isWebsiteNotFoundError(errMsg string) bool {
	// Check if error message contains "website not found" (handles API prefixes like "Failed to process the file.: website not found")
	return strings.Contains(errMsg, "website not found")
}
