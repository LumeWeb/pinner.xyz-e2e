package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"pinner.xyz-e2e/helpers"
	"go.sia.tech/indexd/slabs"
	"go.sia.tech/siastorage"
)

type siaSlabCtxKey string

const (
	siaSlabDetailsKey    siaSlabCtxKey = "sia_slab_details"
	siaSlabListBeforeKey siaSlabCtxKey = "sia_slab_list_before"
)

// SiaSlabSteps holds step definitions for Sia slab operations
type SiaSlabSteps struct{}

// NewSiaSlabSteps creates a new SiaSlabSteps instance
func NewSiaSlabSteps() *SiaSlabSteps {
	return &SiaSlabSteps{}
}

// InitializeScenario registers all Sia slab step definitions with godog
func (s *SiaSlabSteps) InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^the user has uploaded a file to Sia$`, s.aTestCIDIsAvailable)
	ctx.Step(`^the user pins the SIA file$`, s.theUserPinsTheSlabViaPortal)
	ctx.Step(`^the SIA file reaches pinned status$`, s.theSlabReachesStatusPinned)
	ctx.Step(`^the user has a pinned Sia file$`, s.aPinnedSlabExists)
	ctx.Step(`^the user unpins the SIA file$`, s.theUserUnpinsTheSlab)
	ctx.Step(`^the user prunes unpinned Sia files$`, s.theUserPrunesSlabs)
	ctx.Step(`^the SIA file list reflects the changes$`, s.theSlabListReflectsTheChanges)
	ctx.Step(`^the user views the Sia file metadata$`, s.theUserRequestsSlabDetails)
	ctx.Step(`^the SIA file metadata is complete$`, s.slabMetadataIsPresent)
}

func (s *SiaSlabSteps) aTestCIDIsAvailable(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = helpers.SiaUploadRandomData(ctx, 1024)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload random data to SIA: %w", err)
	}
	return ctx, nil
}

func (s *SiaSlabSteps) theUserPinsTheSlabViaPortal(ctx context.Context) (context.Context, error) {
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
		return ctx, fmt.Errorf("failed to retrieve SIA object for slab pinning: %w", err)
	}

	if err := sdk.PinObject(ctx, obj); err != nil {
		return ctx, fmt.Errorf("failed to pin SIA slab via portal: %w", err)
	}

	return ctx, nil
}

func (s *SiaSlabSteps) theSlabReachesStatusPinned(ctx context.Context) (context.Context, error) {
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
		return ctx, fmt.Errorf("failed to verify SIA slab pinned status: %w", err)
	}

	if len(obj.Slabs()) == 0 {
		return ctx, fmt.Errorf("SIA object has no slabs; slab may not be pinned")
	}

	if obj.Size() == 0 {
		return ctx, fmt.Errorf("SIA object has zero size; slab may not be pinned")
	}

	return ctx, nil
}

func (s *SiaSlabSteps) aPinnedSlabExists(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = helpers.SiaUploadRandomData(ctx, 1024)
	if err != nil {
		return ctx, fmt.Errorf("failed to upload random data to SIA for pinned slab: %w", err)
	}

	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context after upload")
	}

	obj, err := sdk.Object(ctx, objectKey)
	if err != nil {
		return ctx, fmt.Errorf("failed to retrieve SIA object for slab setup: %w", err)
	}

	if len(obj.Slabs()) > 0 {
		slabID := obj.Slabs()[0].Digest()
		ctx = helpers.SetSiaSlabID(ctx, slabID.String())
		ctx = helpers.AddSiaSlabCleanup(ctx, slabID.String())
	}

	objects, err := helpers.SiaListObjectEvents(ctx, slabs.Cursor{}, 100)
	if err == nil && len(objects) > 0 {
		ctx = context.WithValue(ctx, siaSlabListBeforeKey, len(objects))
	}

	return ctx, nil
}

func (s *SiaSlabSteps) theUserUnpinsTheSlab(ctx context.Context) (context.Context, error) {
	sdk, ok := helpers.GetSiaSDK(ctx)
	if !ok || sdk == nil {
		return ctx, fmt.Errorf("no SIA SDK available in context")
	}

	objectKey, ok := helpers.GetSiaObjectKey(ctx)
	if !ok {
		return ctx, fmt.Errorf("no SIA object key found in context")
	}

	if err := sdk.DeleteObject(ctx, objectKey); err != nil {
		return ctx, fmt.Errorf("failed to unpin SIA slab (delete object): %w", err)
	}

	return ctx, nil
}

func (s *SiaSlabSteps) theUserPrunesSlabs(ctx context.Context) (context.Context, error) {
	if err := helpers.SiaPruneSlabs(ctx); err != nil {
		return ctx, fmt.Errorf("failed to prune SIA slabs: %w", err)
	}
	return ctx, nil
}

func (s *SiaSlabSteps) theSlabListReflectsTheChanges(ctx context.Context) (context.Context, error) {
	objects, err := helpers.SiaListObjectEvents(ctx, slabs.Cursor{}, 100)
	if err != nil {
		return ctx, fmt.Errorf("failed to list SIA objects after changes: %w", err)
	}

	beforeCount, hadBefore := ctx.Value(siaSlabListBeforeKey).(int)
	if hadBefore && len(objects) >= beforeCount {
		return ctx, fmt.Errorf("expected object count to decrease after unpin/prune: before=%d, after=%d", beforeCount, len(objects))
	}

	return ctx, nil
}

func (s *SiaSlabSteps) aSlabIsPinned(ctx context.Context) (context.Context, error) {
	return s.aPinnedSlabExists(ctx)
}

func (s *SiaSlabSteps) theUserRequestsSlabDetails(ctx context.Context) (context.Context, error) {
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
		return ctx, fmt.Errorf("failed to retrieve SIA object for slab details: %w", err)
	}

	ctx = context.WithValue(ctx, siaSlabDetailsKey, obj)
	return ctx, nil
}

func (s *SiaSlabSteps) slabMetadataIsPresent(ctx context.Context) (context.Context, error) {
	obj, ok := ctx.Value(siaSlabDetailsKey).(siastorage.Object)
	if !ok {
		return ctx, fmt.Errorf("no SIA slab details found in context")
	}

	slabSlices := obj.Slabs()
	if len(slabSlices) == 0 {
		return ctx, fmt.Errorf("SIA slab has no slab slices; metadata may be incomplete")
	}

	for i, ss := range slabSlices {
		var zeroEncKey [32]byte
		if ss.EncryptionKey == zeroEncKey {
			return ctx, fmt.Errorf("slab slice %d has zero encryption key; metadata may be incomplete", i)
		}
		if ss.MinShards == 0 {
			return ctx, fmt.Errorf("slab slice %d has zero min shards; metadata may be incomplete", i)
		}
		if len(ss.Sectors) == 0 {
			return ctx, fmt.Errorf("slab slice %d has no sectors; metadata may be incomplete", i)
		}
	}

	return ctx, nil
}
