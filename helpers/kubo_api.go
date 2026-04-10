package helpers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	goCid "github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	caopts "github.com/ipfs/kubo/core/coreiface/options"
)

// Global RPC client instance for local Kubo
var (
	kuboClient *rpc.HttpApi
	kuboOnce   sync.Once
	kuboErr    error
)

// kuboClientWithPanicHandling wraps getKuboClient with panic handling
// Returns the Kubo client or an error if initialization fails
func kuboClientWithPanicHandling(funcName string, ctx context.Context) (*rpc.HttpApi, error) {
	panicHandler := NewPanicHandler(funcName).WithName(funcName)
	defer panicHandler.RecoverFromPanic(nil)
	return getKuboClient()
}

// requireKuboNameAPI ensures the Name API client is available
// Returns an error if the Name API is nil
func requireKuboNameAPI(client *rpc.HttpApi) error {
	if client.Name() == nil {
		return fmt.Errorf("Kubo client Name API is nil")
	}
	return nil
}

// requireKuboKeyAPI ensures the Key API client is available
// Returns an error if the Key API is nil
func requireKuboKeyAPI(client *rpc.HttpApi) error {
	if client.Key() == nil {
		return fmt.Errorf("Kubo client Key API is nil")
	}
	return nil
}

// createIPFSPath creates an IPFS path from a CID string
// Returns the path or an error if the CID is invalid
func createIPFSPath(cid string) (path.Path, error) {
	if cid == "" {
		return nil, fmt.Errorf("CID cannot be empty")
	}
	return path.NewPath("/ipfs/" + cid)
}

// ExtractIPNSName removes the /ipns/ prefix from an IPNS path
// Returns just the IPNS name (e.g., k51qz...)
func ExtractIPNSName(ipnsPath string) string {
	if strings.HasPrefix(ipnsPath, "/ipns/") {
		return ipnsPath[6:]
	}
	return ipnsPath
}

// getKuboClient returns the Kubo RPC client instance
func getKuboClient() (*rpc.HttpApi, error) {
	kuboOnce.Do(func() {
		ipfsAPIEndpoint := os.Getenv("IPFS_API_ENDPOINT")
		if ipfsAPIEndpoint == "" {
			ipfsAPIEndpoint = "http://127.0.0.1:5001"
		}
		kuboClient, kuboErr = rpc.NewURLApiWithClient(ipfsAPIEndpoint, http.DefaultClient)
		if kuboErr != nil {
			kuboErr = fmt.Errorf("failed to create Kubo client: %w", kuboErr)
		}
	})
	return kuboClient, kuboErr
}

// KuboWaitForSwarmPeers waits until Kubo has at least minimum swarm peers connected
// This ensures peer connections are established before performing network operations
// Polls every 500ms with a default timeout of 2 minutes
func KuboWaitForSwarmPeers(ctx context.Context, minSwarmPeers int) error {
	const (
		pollInterval = 500 * time.Millisecond
		timeout      = 2 * time.Minute
	)

	client, err := getKuboClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubo client: %w", err)
	}

	if client.Swarm() == nil {
		return fmt.Errorf("Kubo client Swarm API is nil")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			peers, _ := client.Swarm().Peers(ctx)
			return fmt.Errorf("timeout waiting for %d swarm peers (currently have %d)", minSwarmPeers, len(peers))
		case <-ticker.C:
			peers, err := client.Swarm().Peers(ctx)
			if err != nil {
				return fmt.Errorf("failed to retrieve swarm peers: %w", err)
			}

			if len(peers) >= minSwarmPeers {
				return nil
			}
		}
	}
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
	// Ensure Kubo has at least 1 swarm peer before attempting operations
	if err := KuboWaitForSwarmPeers(ctx, 1); err != nil {
		return "", fmt.Errorf("swarm peer check failed: %w", err)
	}

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

// KuboResolveIPNS resolves an IPNS name via Kubo's Name API
// Returns the resolved CID or an error
func KuboResolveIPNS(ctx context.Context, ipnsName string) (string, error) {
	client, err := kuboClientWithPanicHandling("KuboResolveIPNS", ctx)
	if err != nil {
		return "", err
	}

	if client.Name() == nil {
		return "", fmt.Errorf("Kubo client Name API is nil")
	}

	// Use Name API to resolve IPNS name
	pathValue, err := client.Name().Resolve(ctx, ipnsName)
	if err != nil {
		return "", fmt.Errorf("failed to resolve IPNS name %s via Kubo Name API: %w", ipnsName, err)
	}

	if pathValue == nil {
		return "", fmt.Errorf("Kubo Name API returned nil path for %s", ipnsName)
	}

	// Return the resolved CID (without /ipfs/ prefix)
	pathStr := pathValue.String()
	if strings.HasPrefix(pathStr, "/ipfs/") {
		return pathStr[6:], nil
	}
	return pathStr, nil
}

// KuboIPNSPullFromPortal pulls an IPNS record from the Portal API and stores it in Kubo
// This cross-node operation ensures Kubo can resolve IPNS names managed by the Portal

// KuboIPNSPublishPath publishes a CID to an IPNS key in Kubo
// Uses Kubo's Name.Publish API to create an IPNS record
// Returns the IPNS name (e.g., k51qz...) or an error
func KuboIPNSPublishPath(ctx context.Context, keyName string, cid string) (string, error) {
	const funcName = "KuboIPNSPublishPath"
	
	client, err := kuboClientWithPanicHandling(funcName, ctx)
	if err != nil {
		return "", err
	}

	if err := requireKuboNameAPI(client); err != nil {
		return "", err
	}

	if keyName == "" {
		return "", fmt.Errorf("Key name cannot be empty")
	}

	ipfsPath, err := createIPFSPath(cid)
	if err != nil {
		return "", err
	}

	ipnsName, err := client.Name().Publish(ctx, ipfsPath, caopts.Name.Key(keyName))
	if err != nil {
		return "", fmt.Errorf("failed to publish CID %s to IPNS key %s via Kubo: %w", cid, keyName, err)
	}

	return ipnsName.String(), nil
}

// KuboIPNSListKeys lists all IPNS keys in Kubo
// Returns a list of key names or an error
func KuboIPNSListKeys(ctx context.Context) ([]string, error) {
	const funcName = "KuboIPNSListKeys"
	
	client, err := kuboClientWithPanicHandling(funcName, ctx)
	if err != nil {
		return nil, err
	}

	if err := requireKuboKeyAPI(client); err != nil {
		return nil, err
	}

	keyList, err := client.Key().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPNS keys in Kubo: %w", err)
	}

	keys := make([]string, len(keyList))
	for i, key := range keyList {
		keys[i] = key.Name()
	}

	return keys, nil
}

// KuboIPNSCreateKey creates a new IPNS key in Kubo
// Returns the key name or an error
func KuboIPNSCreateKey(ctx context.Context, name string) (string, error) {
	const funcName = "KuboIPNSCreateKey"
	
	client, err := kuboClientWithPanicHandling(funcName, ctx)
	if err != nil {
		return "", err
	}

	if err := requireKuboKeyAPI(client); err != nil {
		return "", err
	}

	key, err := client.Key().Generate(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to create IPNS key %s in Kubo: %w", name, err)
	}

	return key.Name(), nil
}

// KuboFetchViaIPFSNetwork fetches a CID via IPFS network by reading (catting) it
// This triggers bitswap to download the content from the IPFS network if not local
// Using cat instead of pin because pin is async and can cause race conditions with test cleanup
// Portal tracks this download against the user's download quota
func KuboFetchViaIPFSNetwork(ctx context.Context, cidString string) error {
	// Note: KuboCat internally calls KuboWaitForSwarmPeers, ensuring sufficient
	// swarm peers are available before triggering bitswap download
	//
	// This avoids race conditions where test cleanup deletes the quota plan before
	// portal service can check download quota
	_, err := KuboCat(ctx, cidString)
	if err != nil {
		return fmt.Errorf("failed to fetch content via IPFS network: %w", err)
	}

	return nil
}

