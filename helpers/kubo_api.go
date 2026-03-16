package helpers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	goCid "github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
)

// Global RPC client instance for local Kubo
var kuboClient *rpc.HttpApi

// getKuboClient returns the Kubo RPC client instance
func getKuboClient() (*rpc.HttpApi, error) {
	if kuboClient != nil {
		return kuboClient, nil
	}

	// Get IPFS API endpoint from environment
	ipfsAPIEndpoint := os.Getenv("IPFS_API_ENDPOINT")
	if ipfsAPIEndpoint == "" {
		ipfsAPIEndpoint = "http://127.0.0.1:5001"
	}

	// Create new HTTP API client with default HTTP client
	var err error
	kuboClient, err = rpc.NewURLApiWithClient(ipfsAPIEndpoint, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubo client: %w", err)
	}

	return kuboClient, nil
}

// KuboAdd adds the given content to local Kubo and returns the CID
// This is ONLY for generating test CIDs that exist in the IPFS network
func KuboAdd(ctx context.Context, content []byte) (string, error) {
	panicHandler := NewPanicHandler("").WithName("KuboAdd")
	defer panicHandler.RecoverFromPanic(nil)

	client, err := getKuboClient()
	if err != nil {
		return "", err
	}

	if client == nil {
		return "", fmt.Errorf("Kubo client is nil after initialization")
	}

	f := files.NewBytesFile(content)
	if f == nil {
		return "", fmt.Errorf("failed to create file node from bytes")
	}

	if client.Unixfs() == nil {
		return "", fmt.Errorf("Kubo client Unixfs API is nil")
	}

	addedPath, err := client.Unixfs().Add(ctx, f)
	if err != nil {
		return "", fmt.Errorf("failed to add file to Kubo: %w", err)
	}

	// Pin the content to prevent garbage collection
	if client.Pin() == nil {
		return "", fmt.Errorf("Kubo client Pin API is nil")
	}

	if err := client.Pin().Add(ctx, addedPath); err != nil {
		return "", fmt.Errorf("failed to pin content in Kubo: %w", err)
	}

	// Return the CID as a string (without /ipfs/ prefix)
	pathStr := addedPath.String()
	if strings.HasPrefix(pathStr, "/ipfs/") {
		return pathStr[6:], nil
	}
	return pathStr, nil
}

// KuboCat retrieves the content of a CID from local Kubo
func KuboCat(ctx context.Context, cidString string) (string, error) {
	client, err := getKuboClient()
	if err != nil {
		return "", err
	}

	decodedCid, err := goCid.Decode(cidString)
	if err != nil {
		return "", fmt.Errorf("invalid CID: %w", err)
	}

	p := path.FromCid(decodedCid)

	node, err := client.Unixfs().Get(ctx, p)
	if err != nil {
		return "", fmt.Errorf("failed to get file from Kubo: %w", err)
	}
	defer node.Close()

	file, ok := node.(files.File)
	if !ok {
		return "", fmt.Errorf("node is not a file")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return string(content), nil
}
