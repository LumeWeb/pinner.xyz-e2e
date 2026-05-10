package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/indexd/slabs"
	"go.sia.tech/siastorage"
)

type siaCtxKey string

const (
	siaDownloadedContentKey   siaCtxKey = "sia_downloaded_content"
	siaObjectMetadataKey      siaCtxKey = "sia_object_metadata"
	siaExpectedMetadataKey    siaCtxKey = "sia_expected_metadata"
	siaRetrievedMetadataKey   siaCtxKey = "sia_retrieved_metadata"
)

// SiaObjectSteps holds step definitions for Sia object operations
type SiaObjectSteps struct {
	testPayload []byte
}

// NewSiaObjectSteps creates a new SiaObjectSteps instance
func NewSiaObjectSteps() *SiaObjectSteps {
	return &SiaObjectSteps{}
}

// InitializeScenario registers all Sia object step definitions with godog
func (s *SiaObjectSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user has a small Sia test file$`, s.theUserHasASmallTestPayload)
	ctx.Step(`^the user uploads the file to Sia$`, s.theUserUploadsTheTestPayloadToSIAViaPortal)
	ctx.Step(`^the SIA file is stored successfully$`, s.theObjectShouldBePinned)
	ctx.Step(`^the user has a stored Sia file$`, s.anObjectCIDIsPinned)
	ctx.Step(`^the user downloads the file from Sia$`, s.theUserDownloadsTheObjectViaPortal)
	ctx.Step(`^the downloaded Sia content matches the original$`, s.theDownloadedContentShouldMatchTheOriginal)
	ctx.Step(`^the user views the Sia file details$`, s.theUserRetrievesTheObjectByKey)
	ctx.Step(`^the SIA file information is displayed$`, s.theObjectMetadataShouldBePresent)
	ctx.Step(`^the user removes the file from Sia$`, s.theUserRemovesTheFileFromSia)
	ctx.Step(`^the SIA file is no longer listed$`, s.theFileIsNoLongerListed)
	ctx.Step(`^the user uploads a file with custom SIA metadata$`, s.theUserUploadsAFileWithCustomMetadata)
	ctx.Step(`^the SIA object metadata matches the stored metadata$`, s.theRetrievedMetadataMatchesTheStoredMetadata)
	ctx.Step(`^the user retrieves the SIA object metadata$`, s.theUserRetrievesTheObjectMetadata)
}

func (s *SiaObjectSteps) theUserHasASmallTestPayload(ctx context.Context) (context.Context, error) {
	s.testPayload = []byte(helpers.GenerateUniqueContent("sia-e2e-test-payload"))
	ctx = helpers.SetKnownContent(ctx, string(s.testPayload))
	return ctx, nil
}

func (s *SiaObjectSteps) theUserUploadsTheTestPayloadToSIAViaPortal(ctx context.Context) (context.Context, error) {
	if len(s.testPayload) == 0 {
		return ctx, fmt.Errorf("no test payload available; run 'the user has a small test payload' first")
	}

	var err error
	ctx, err = helpers.SiaUploadData(ctx, bytes.NewReader(s.testPayload), int64(len(s.testPayload)), 1, 1)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload test payload to SIA: %w", err)
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theObjectShouldBePinned(ctx context.Context) (context.Context, error) {
	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("expected SIA object key in context, but it was not found")
	}

	var zeroKey [32]byte
	if objectKey == zeroKey {
		return ctx, fmt.Errorf("SIA object key is zero value; object may not have been pinned")
	}

	return ctx, nil
}

func (s *SiaObjectSteps) anObjectCIDIsPinned(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = helpers.SiaUploadRandomData(ctx, 1024)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload random data to SIA: %w", err)
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theUserDownloadsTheObjectViaPortal(ctx context.Context) (context.Context, error) {
	var buf bytes.Buffer
	if err := helpers.SiaDownloadObject(ctx, &buf); err != nil {
		return ctx, fmt.Errorf("failed to download SIA object: %w", err)
	}

	ctx = context.WithValue(ctx, siaDownloadedContentKey, buf.String())
	return ctx, nil
}

func (s *SiaObjectSteps) theDownloadedContentShouldMatchTheOriginal(ctx context.Context) (context.Context, error) {
	downloaded, ok := ctx.Value(siaDownloadedContentKey).(string)
	if !ok {
		return ctx, fmt.Errorf("no downloaded content found in context")
	}

	original, ok := helpers.GetKnownContent(ctx)
	if !ok {
		return ctx, fmt.Errorf("no original content found in context")
	}

	if downloaded != original {
		return ctx, fmt.Errorf("downloaded content does not match original: got %d bytes, expected %d bytes", len(downloaded), len(original))
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theUserRetrievesTheObjectByKey(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context")
	}

	obj, err := sdk.Object(ctx, objectKey)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve SIA object by key: %w", err)
	}

	ctx = context.WithValue(ctx, siaObjectMetadataKey, obj)
	return ctx, nil
}

func (s *SiaObjectSteps) theUserRetrievesTheObjectDetailsByItsKey(ctx context.Context) (context.Context, error) {
	return s.theUserRetrievesTheObjectByKey(ctx)
}

func (s *SiaObjectSteps) theObjectMetadataShouldBePresent(ctx context.Context) (context.Context, error) {
	obj, ok := ctx.Value(siaObjectMetadataKey).(siastorage.Object)
	if !ok {
		return ctx, fmt.Errorf("no SIA object metadata found in context")
	}

	if len(obj.Slabs()) == 0 {
		return ctx, fmt.Errorf("SIA object has no slabs; metadata may be incomplete")
	}

	if obj.Size() == 0 {
		return ctx, fmt.Errorf("SIA object has zero size; metadata may be incomplete")
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theUserRemovesTheFileFromSia(ctx context.Context) (context.Context, error) {
	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context")
	}

	if err := helpers.SiaDeleteObject(ctx, objectKey); err != nil {
		return ctx, fmt.Errorf("failed to remove SIA file: %w", err)
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theFileIsNoLongerListed(ctx context.Context) (context.Context, error) {
	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context")
	}

	objects, err := helpers.SiaListObjectEvents(ctx, slabs.Cursor{}, 100)
	if err != nil {
		return ctx, fmt.Errorf("failed to list SIA objects: %w", err)
	}

	for _, obj := range objects {
		if obj.ID() == objectKey {
			return ctx, fmt.Errorf("SIA file is still listed after removal")
		}
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theUserUploadsAFileWithCustomMetadata(ctx context.Context) (context.Context, error) {
	s.testPayload = []byte(helpers.GenerateUniqueContent("sia-e2e-metadata"))
	ctx = helpers.SetKnownContent(ctx, string(s.testPayload))

	metadata, err := json.Marshal(map[string]string{
		"filename":   "e2e-test-file.txt",
		"department": "engineering",
		"version":    "1.0.0",
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to marshal test metadata: %w", err)
	}
	ctx = context.WithValue(ctx, siaExpectedMetadataKey, json.RawMessage(metadata))

	ctx, err = helpers.SiaUploadDataWithMetadata(ctx, bytes.NewReader(s.testPayload), int64(len(s.testPayload)), 1, 1, metadata)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload file with metadata to SIA: %w", err)
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theRetrievedMetadataMatchesTheStoredMetadata(ctx context.Context) (context.Context, error) {
	expected, ok := ctx.Value(siaExpectedMetadataKey).(json.RawMessage)
	if !ok {
		return ctx, fmt.Errorf("no expected metadata found in context")
	}

	retrieved, ok := ctx.Value(siaRetrievedMetadataKey).(json.RawMessage)
	if !ok {
		return ctx, fmt.Errorf("no retrieved metadata found in context")
	}

	var expectedMap, retrievedMap map[string]string
	if err := json.Unmarshal(expected, &expectedMap); err != nil {
		return ctx, fmt.Errorf("failed to unmarshal expected metadata: %w", err)
	}
	if err := json.Unmarshal(retrieved, &retrievedMap); err != nil {
		return ctx, fmt.Errorf("failed to unmarshal retrieved metadata: %w", err)
	}

	for k, v := range expectedMap {
		if retrievedMap[k] != v {
			return ctx, fmt.Errorf("metadata mismatch for key %q: expected %q, got %q", k, v, retrievedMap[k])
		}
	}

	return ctx, nil
}

func (s *SiaObjectSteps) theUserRetrievesTheObjectMetadata(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context")
	}

	obj, err := sdk.Object(ctx, objectKey)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve SIA object metadata: %w", err)
	}

	ctx = context.WithValue(ctx, siaRetrievedMetadataKey, obj.Metadata())
	return ctx, nil
}
