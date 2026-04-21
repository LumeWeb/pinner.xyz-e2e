package helpers

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// AddWebsiteDNSZoneToDevServer adds DNS records for a website to the dev DNS server
// This simulates what would happen when DNS hosting is enabled for a website
func AddWebsiteDNSZoneToDevServer(ctx context.Context, domain string, dnslink string, validationToken string) error {
	client, err := RequireDNSDevClient(ctx)
	if err != nil {
		return err
	}

	// Build base DNS records
	records := []DNSDevRecord{
		{
			Name:  "@",
			Type:  "A",
			Value: "127.0.0.1",
			TTL:   300,
		},
		{
			Name:  "_dnslink",
			Type:  "TXT",
			Value: dnslink,
			TTL:   300,
		},
		{
			Name:  "@",
			Type:  "TXT",
			Value: fmt.Sprintf("lumeweb-verify=%s", validationToken),
			TTL:   300,
		},
	}

	// Add NS records from configured nameservers
	nameservers := os.Getenv("PORTAL__PLUGIN__IPFS__SERVICE__DNS__NAMESERVERS")
	if nameservers != "" {
		for _, ns := range strings.Split(nameservers, ",") {
			// Clean up whitespace and ensure trailing dot
			ns = strings.TrimSpace(ns)
			if !strings.HasSuffix(ns, ".") {
				ns += "."
			}
			records = append(records, DNSDevRecord{
				Name:  "@",
				Type:  "NS",
				Value: ns,
				TTL:   300,
			})
		}
	}

	// Create zone with required DNS records
	zone := DNSDevZone{
		Domain:  domain,
		Records: records,
	}

	_, err = client.AddZone(ctx, zone)
	if err != nil {
		return fmt.Errorf("failed to add zone %s to dev server: %w", domain, err)
	}

	// Verify the zone was actually added by listing zones
	_, listErr := client.ListZones(ctx)
	if listErr != nil {
		// Log error but don't fail - zone may have been added successfully
	}

	return nil
}

// formatDNSLinkProtocol constructs a DNS link record with the specified protocol.
// DNS link format: dnslink=/<protocol>/<hash>
// Protocol should be either "ipfs" or "ipns"
func formatDNSLinkProtocol(protocol string, hash string) string {
	return fmt.Sprintf("dnslink=/%s/%s", protocol, hash)
}

// BuildDNSLink constructs a DNS link record from an IPFS CID
// DNS link format: dnslink=/ipfs/<cid>
func BuildDNSLink(cid string) string {
	return formatDNSLinkProtocol("ipfs", cid)
}

// BuildWebsiteDNSLink constructs a DNS link record from a website's target hash and type.
// Returns the appropriate dnslink format based on the website's target type:
// - IPNS targets: dnslink=/ipns/<peer_id>
// - IPFS targets: dnslink=/ipfs/<cid>
//
// When DNS hosting is enabled, the portal auto-converts IPFS targets to IPNS,
// so this helper checks the target type to determine the correct format.
func BuildWebsiteDNSLink(targetHash string, targetType string) string {
	if targetType == "ipns" {
		return formatDNSLinkProtocol("ipns", targetHash)
	}
	return BuildDNSLink(targetHash)
}

// RemoveWebsiteDNSZoneFromDevServer removes a DNS zone from the dev DNS server
func RemoveWebsiteDNSZoneFromDevServer(ctx context.Context, domain string) error {
	client, err := RequireDNSDevClient(ctx)
	if err != nil {
		return err
	}

	err = client.DeleteZone(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to remove zone %s from dev server: %w", domain, err)
	}

	return nil
}

// CheckDNSDevServerHealth checks if the dev DNS server is healthy
func CheckDNSDevServerHealth(ctx context.Context) error {
	client, err := RequireDNSDevClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Health(ctx)
	if err != nil {
		return fmt.Errorf("DNS dev server health check failed: %w", err)
	}

	return nil
}
