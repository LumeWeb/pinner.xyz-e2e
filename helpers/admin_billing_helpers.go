package helpers

import (
	"context"
	"fmt"
	"strings"
	"time"

	// Side effects import to ensure admin subpackage is vendored
	admin "go.lumeweb.com/portal-sdk/admin"
)

// Context keys for admin billing management
const (
	AdminCreditsKey              contextKey = "admin_credits"
	AdminCurrentCreditKey        contextKey = "admin_current_credit"
	AdminCurrentCreditIDKey      contextKey = "admin_current_credit_id"
	AdminUserBalanceKey          contextKey = "admin_user_balance"
	AdminPriceLinesKey           contextKey = "admin_price_lines"
	AdminCurrentPriceLineKey     contextKey = "admin_current_price_line"
	AdminCurrentPriceLineIDKey   contextKey = "admin_current_price_line_id"
	AdminPricingPlansKey         contextKey = "admin_pricing_plans"
	AdminCurrentPricingPlanKey   contextKey = "admin_current_pricing_plan"
	AdminCurrentPricingPlanIDKey contextKey = "admin_current_pricing_plan_id"
	AdminPricingPlanPeriodsKey   contextKey = "admin_pricing_plan_periods"
	AdminCurrentPeriodKey        contextKey = "admin_current_period"
	AdminCurrentPeriodIDKey      contextKey = "admin_current_period_id"
	AdminPurgedCreditsCountKey   contextKey = "admin_purged_credits_count"
	AdminSubscribersKey          contextKey = "admin_subscribers"
	AdminCurrentSubscriberKey    contextKey = "admin_current_subscriber"
	AdminCurrentSubscriberIDKey  contextKey = "admin_current_subscriber_id"
	AdminManagementResultKey     contextKey = "admin_management_result"
	AdminPlanChangeResultKey     contextKey = "admin_plan_change_result"
)

// BillingID is a type alias for billing ID values that can be stored/retrieved
// as either string or int without manual conversion everywhere.
type BillingID string

// AsInt returns the ID as int
func (id BillingID) AsInt() int {
	var result int
	fmt.Sscanf(string(id), "%d", &result)
	return result
}

// AsInt64 returns the ID as int64
func (id BillingID) AsInt64() int64 {
	var result int64
	fmt.Sscanf(string(id), "%d", &result)
	return result
}

// String returns the ID as string (implements fmt.Stringer)
func (id BillingID) String() string {
	return string(id)
}

// AsInterface returns the ID as interface{} for generic use
func (id BillingID) AsInterface() interface{} {
	return id.AsInt64()
}

// ToBillingID converts an int to BillingID
func ToBillingID(id int) BillingID {
	return BillingID(fmt.Sprint(id))
}

// ToBillingIDFromInt64 converts an int64 to BillingID
func ToBillingIDFromInt64(id int64) BillingID {
	return BillingID(fmt.Sprint(id))
}

// ToBillingIDFromUUID converts a UUID to BillingID
func ToBillingIDFromUUID(id any) BillingID {
	// Handle different UUID formats from generated SDK
	switch v := id.(type) {
	case string:
		return BillingID(v)
	case [16]byte:
		return BillingID(fmt.Sprintf("%x-%x-%x-%x-%x",
			v[0:4], v[4:6], v[6:8], v[8:10], v[10:16]))
	default:
		return BillingID(fmt.Sprint(id))
	}
}

// GetBillingIDFromContext retrieves a BillingID from context using the given key
func GetBillingIDFromContext(ctx context.Context, key contextKey) (BillingID, bool) {
	id, ok := GetContextValue[string](ctx, key)
	if !ok || id == "" {
		return "", false
	}
	return BillingID(id), true
}

// RequireBillingIDFromContext retrieves a BillingID from context or returns error
func RequireBillingIDFromContext(ctx context.Context, key contextKey, name string) (BillingID, error) {
	id, ok := GetBillingIDFromContext(ctx, key)
	if !ok {
		return "", fmt.Errorf("admin %s ID not available in context", name)
	}
	return id, nil
}

// Cleanup list keys
const (
	AdminCreditsCleanupKey            contextKey = "admin_credits_cleanup"
	AdminPriceLinesCleanupKey         contextKey = "admin_price_lines_cleanup"
	AdminPricingPlansCleanupKey       contextKey = "admin_pricing_plans_cleanup"
	AdminPricingPlanPeriodsCleanupKey contextKey = "admin_pricing_plan_periods_cleanup"
)

// GetPlansFromPriceLine safely extracts plans from PriceLineDetailResponse
// Returns the plans slice or empty slice if price line is nil
func GetPlansFromPriceLine(priceLine *admin.PriceLineDetailResponse) []*admin.PricingPlanItem {
	if priceLine == nil {
		return []*admin.PricingPlanItem{}
	}
	return priceLine.Plans
}

// CountPlansInPriceLine returns the number of plans in a price line
func CountPlansInPriceLine(priceLine *admin.PriceLineDetailResponse) int {
	if priceLine == nil || priceLine.Plans == nil {
		return 0
	}
	return len(priceLine.Plans)
}

// PlanExistsInPriceLine checks if a plan ID exists in a price line
func PlanExistsInPriceLine(priceLine *admin.PriceLineDetailResponse, planID BillingID) bool {
	if priceLine == nil || priceLine.Plans == nil {
		return false
	}
	id := planID.AsInt()
	for _, plan := range priceLine.Plans {
		if plan.Id == id {
			return true
		}
	}
	return false
}

// GetPlanPositionInPriceLine returns the position of a plan in a price line, or -1 if not found
func GetPlanPositionInPriceLine(priceLine *admin.PriceLineDetailResponse, planID BillingID) int {
	if priceLine == nil || priceLine.Plans == nil {
		return -1
	}
	id := planID.AsInt()
	for _, plan := range priceLine.Plans {
		if plan.Id == id {
			return plan.Position
		}
	}
	return -1
}

// SetAdminCredits stores credits in context
func SetAdminCredits(ctx context.Context, credits []*admin.CreditItem) context.Context {
	return SetContextValue(ctx, AdminCreditsKey, credits)
}

// GetAdminCredits retrieves credits from context
func GetAdminCredits(ctx context.Context) ([]*admin.CreditItem, bool) {
	return GetContextValue[[]*admin.CreditItem](ctx, AdminCreditsKey)
}

// SetAdminCurrentCredit stores current credit in context
func SetAdminCurrentCredit(ctx context.Context, credit *admin.Credit) context.Context {
	return SetContextValue(ctx, AdminCurrentCreditKey, credit)
}

// GetAdminCurrentCredit retrieves current credit from context
func GetAdminCurrentCredit(ctx context.Context) (*admin.Credit, bool) {
	return GetContextValue[*admin.Credit](ctx, AdminCurrentCreditKey)
}

// RequireAdminCurrentCredit retrieves current credit or returns error
func RequireAdminCurrentCredit(ctx context.Context) (*admin.Credit, error) {
	credit, ok := GetAdminCurrentCredit(ctx)
	if !ok {
		return nil, fmt.Errorf("admin credit not available in context")
	}
	return credit, nil
}

// SetAdminCurrentCreditID stores current credit ID in context
func SetAdminCurrentCreditID(ctx context.Context, creditID BillingID) context.Context {
	return SetContextValue(ctx, AdminCurrentCreditIDKey, creditID.String())
}

// GetAdminCurrentCreditID retrieves current credit ID from context
func GetAdminCurrentCreditID(ctx context.Context) (BillingID, bool) {
	return GetBillingIDFromContext(ctx, AdminCurrentCreditIDKey)
}

// RequireAdminCurrentCreditID retrieves current credit ID or returns error
func RequireAdminCurrentCreditID(ctx context.Context) (BillingID, error) {
	return RequireBillingIDFromContext(ctx, AdminCurrentCreditIDKey, "credit")
}

// SetAdminUserBalance stores user balance in context
func SetAdminUserBalance(ctx context.Context, balance *admin.UserBalance) context.Context {
	return SetContextValue(ctx, AdminUserBalanceKey, balance)
}

// GetAdminUserBalance retrieves user balance from context
func GetAdminUserBalance(ctx context.Context) (*admin.UserBalance, bool) {
	return GetContextValue[*admin.UserBalance](ctx, AdminUserBalanceKey)
}

// RequireAdminUserBalance retrieves user balance or returns error
func RequireAdminUserBalance(ctx context.Context) (*admin.UserBalance, error) {
	balance, ok := GetAdminUserBalance(ctx)
	if !ok {
		return nil, fmt.Errorf("admin user balance not available in context")
	}
	return balance, nil
}

// SetAdminPriceLines stores price lines in context
func SetAdminPriceLines(ctx context.Context, priceLines []*admin.PriceLine) context.Context {
	return SetContextValue(ctx, AdminPriceLinesKey, priceLines)
}

// GetAdminPriceLines retrieves price lines from context
func GetAdminPriceLines(ctx context.Context) ([]*admin.PriceLine, bool) {
	return GetContextValue[[]*admin.PriceLine](ctx, AdminPriceLinesKey)
}

// SetAdminCurrentPriceLine stores current price line in context
func SetAdminCurrentPriceLine(ctx context.Context, priceLine any) context.Context {
	switch v := priceLine.(type) {
	case *admin.PriceLine:
		return SetContextValue[*admin.PriceLine](ctx, AdminCurrentPriceLineKey, v)
	case *admin.PriceLineDetailResponse:
		return SetContextValue[*admin.PriceLineDetailResponse](ctx, AdminCurrentPriceLineKey, v)
	default:
		return SetContextValue[*admin.PriceLine](ctx, AdminCurrentPriceLineKey, nil)
	}
}

// GetAdminCurrentPriceLine retrieves current price line from context
func GetAdminCurrentPriceLine(ctx context.Context) (any, bool) {
	return GetContextValue[any](ctx, AdminCurrentPriceLineKey)
}

// RequireAdminCurrentPriceLine retrieves current price line or returns error
func RequireAdminCurrentPriceLine(ctx context.Context) (any, error) {
	priceLine, ok := GetAdminCurrentPriceLine(ctx)
	if !ok {
		return nil, fmt.Errorf("admin price line not available in context")
	}
	return priceLine, nil
}

// PriceLineFields represents the common fields shared by PriceLine and PriceLineDetailResponse
type PriceLineFields struct {
	Id          int
	Name        string
	Description string
	IsActive    bool
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ExtractPriceLineFields extracts common fields from either PriceLine or PriceLineDetailResponse
func ExtractPriceLineFields(priceLine any) (*PriceLineFields, error) {
	switch pl := priceLine.(type) {
	case *admin.PriceLine:
		return &PriceLineFields{
			Id:          pl.Id,
			Name:        pl.Name,
			Description: pl.Description,
			IsActive:    pl.IsActive,
			IsDefault:   pl.IsDefault,
			CreatedAt:   pl.CreatedAt,
			UpdatedAt:   pl.UpdatedAt,
		}, nil
	case *admin.PriceLineDetailResponse:
		return &PriceLineFields{
			Id:          pl.Id,
			Name:        pl.Name,
			Description: pl.Description,
			IsActive:    pl.IsActive,
			IsDefault:   pl.IsDefault,
			CreatedAt:   pl.CreatedAt,
			UpdatedAt:   pl.UpdatedAt,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected price line type: %T", priceLine)
	}
}

// SetAdminCurrentPriceLineID stores current price line ID in context
func SetAdminCurrentPriceLineID(ctx context.Context, priceLineID BillingID) context.Context {
	return SetContextValue(ctx, AdminCurrentPriceLineIDKey, priceLineID.String())
}

// GetAdminCurrentPriceLineID retrieves current price line ID from context
func GetAdminCurrentPriceLineID(ctx context.Context) (BillingID, bool) {
	id, ok := GetContextValue[string](ctx, AdminCurrentPriceLineIDKey)
	if !ok || id == "" {
		return "", false
	}
	return BillingID(id), true
}

// RequireAdminCurrentPriceLineID retrieves current price line ID or returns error
func RequireAdminCurrentPriceLineID(ctx context.Context) (BillingID, error) {
	priceLineID, ok := GetAdminCurrentPriceLineID(ctx)
	if !ok {
		return "", fmt.Errorf("admin price line ID not available in context")
	}
	return priceLineID, nil
}

// isNotFoundError checks if error indicates resource not found
func isNotFoundError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found")
}

// SetAdminPricingPlans stores pricing plans in context
func SetAdminPricingPlans(ctx context.Context, plans []*admin.PricingPlanItem) context.Context {
	return SetContextValue(ctx, AdminPricingPlansKey, plans)
}

// GetAdminPricingPlans retrieves pricing plans from context
func GetAdminPricingPlans(ctx context.Context) ([]*admin.PricingPlanItem, bool) {
	return GetContextValue[[]*admin.PricingPlanItem](ctx, AdminPricingPlansKey)
}

// SetAdminCurrentPricingPlan stores current pricing plan in context
func SetAdminCurrentPricingPlan(ctx context.Context, plan *admin.PricingPlan) context.Context {
	return SetContextValue(ctx, AdminCurrentPricingPlanKey, plan)
}

// GetAdminCurrentPricingPlan retrieves current pricing plan from context
func GetAdminCurrentPricingPlan(ctx context.Context) (*admin.PricingPlan, bool) {
	return GetContextValue[*admin.PricingPlan](ctx, AdminCurrentPricingPlanKey)
}

// RequireAdminCurrentPricingPlan retrieves current pricing plan or returns error
func RequireAdminCurrentPricingPlan(ctx context.Context) (*admin.PricingPlan, error) {
	plan, ok := GetAdminCurrentPricingPlan(ctx)
	if !ok {
		return nil, fmt.Errorf("admin pricing plan not available in context")
	}
	return plan, nil
}

// SetAdminCurrentPricingPlanID stores current pricing plan ID in context
func SetAdminCurrentPricingPlanID(ctx context.Context, pricingPlanID BillingID) context.Context {
	return SetContextValue(ctx, AdminCurrentPricingPlanIDKey, pricingPlanID.String())
}

// GetAdminCurrentPricingPlanID retrieves current pricing plan ID from context
func GetAdminCurrentPricingPlanID(ctx context.Context) (BillingID, bool) {
	id, ok := GetContextValue[string](ctx, AdminCurrentPricingPlanIDKey)
	if !ok || id == "" {
		return "", false
	}
	return BillingID(id), true
}

// RequireAdminCurrentPricingPlanID retrieves current pricing plan ID or returns error
func RequireAdminCurrentPricingPlanID(ctx context.Context) (BillingID, error) {
	pricingPlanID, ok := GetAdminCurrentPricingPlanID(ctx)
	if !ok {
		return "", fmt.Errorf("admin pricing plan ID not available in context")
	}
	return pricingPlanID, nil
}

// SetAdminPricingPlanPeriods stores pricing plan periods in context
func SetAdminPricingPlanPeriods(ctx context.Context, periods []*admin.PricingPlanPeriod) context.Context {
	return SetContextValue(ctx, AdminPricingPlanPeriodsKey, periods)
}

// GetAdminPricingPlanPeriods retrieves pricing plan periods from context
func GetAdminPricingPlanPeriods(ctx context.Context) ([]*admin.PricingPlanPeriod, bool) {
	return GetContextValue[[]*admin.PricingPlanPeriod](ctx, AdminPricingPlanPeriodsKey)
}

// SetAdminCurrentPeriod stores current pricing plan period in context
func SetAdminCurrentPeriod(ctx context.Context, period *admin.PricingPlanPeriod) context.Context {
	return SetContextValue(ctx, AdminCurrentPeriodKey, period)
}

// GetAdminCurrentPeriod retrieves current pricing plan period from context
func GetAdminCurrentPeriod(ctx context.Context) (*admin.PricingPlanPeriod, bool) {
	return GetContextValue[*admin.PricingPlanPeriod](ctx, AdminCurrentPeriodKey)
}

// RequireAdminCurrentPeriod retrieves current period or returns error
func RequireAdminCurrentPeriod(ctx context.Context) (*admin.PricingPlanPeriod, error) {
	period, ok := GetAdminCurrentPeriod(ctx)
	if !ok {
		return nil, fmt.Errorf("admin pricing plan period not available in context")
	}
	return period, nil
}

// SetAdminCurrentPeriodID stores current pricing plan period ID in context
func SetAdminCurrentPeriodID(ctx context.Context, periodID BillingID) context.Context {
	return SetContextValue(ctx, AdminCurrentPeriodIDKey, periodID.String())
}

// GetAdminCurrentPeriodID retrieves current pricing plan period ID from context
func GetAdminCurrentPeriodID(ctx context.Context) (BillingID, bool) {
	id, ok := GetContextValue[string](ctx, AdminCurrentPeriodIDKey)
	if !ok {
		return "", false
	}
	return BillingID(id), true
}

// RequireAdminCurrentPeriodID retrieves current pricing plan period ID or returns error
func RequireAdminCurrentPeriodID(ctx context.Context) (BillingID, error) {
	periodID, ok := GetAdminCurrentPeriodID(ctx)
	if !ok {
		return "", fmt.Errorf("admin pricing plan period ID not available in context")
	}
	return periodID, nil
}

// SetAdminPurgedCreditsCount stores purged credits count in context
func SetAdminPurgedCreditsCount(ctx context.Context, count int) context.Context {
	return SetContextValue(ctx, AdminPurgedCreditsCountKey, count)
}

// GetAdminPurgedCreditsCount retrieves purged credits count from context
func GetAdminPurgedCreditsCount(ctx context.Context) (int, bool) {
	return GetContextValue[int](ctx, AdminPurgedCreditsCountKey)
}

// SetAdminSubscribers stores subscribers in context
func SetAdminSubscribers(ctx context.Context, subscribers []*admin.Subscriber) context.Context {
	return SetContextValue(ctx, AdminSubscribersKey, subscribers)
}

// GetAdminSubscribers retrieves subscribers from context
func GetAdminSubscribers(ctx context.Context) ([]*admin.Subscriber, bool) {
	return GetContextValue[[]*admin.Subscriber](ctx, AdminSubscribersKey)
}

// SetAdminCurrentSubscriber stores current subscriber in context
func SetAdminCurrentSubscriber(ctx context.Context, subscriber *admin.Subscriber) context.Context {
	return SetContextValue(ctx, AdminCurrentSubscriberKey, subscriber)
}

// GetAdminCurrentSubscriber retrieves current subscriber from context
func GetAdminCurrentSubscriber(ctx context.Context) (*admin.Subscriber, bool) {
	return GetContextValue[*admin.Subscriber](ctx, AdminCurrentSubscriberKey)
}

// RequireAdminCurrentSubscriber retrieves current subscriber or returns error
func RequireAdminCurrentSubscriber(ctx context.Context) (*admin.Subscriber, error) {
	subscriber, ok := GetAdminCurrentSubscriber(ctx)
	if !ok {
		return nil, fmt.Errorf("admin subscriber not available in context")
	}
	return subscriber, nil
}

// SetAdminCurrentSubscriberID stores current subscriber ID in context
func SetAdminCurrentSubscriberID(ctx context.Context, subscriberID BillingID) context.Context {
	return SetContextValue(ctx, AdminCurrentSubscriberIDKey, subscriberID.String())
}

// GetAdminCurrentSubscriberID retrieves current subscriber ID from context
func GetAdminCurrentSubscriberID(ctx context.Context) (BillingID, bool) {
	id, ok := GetContextValue[string](ctx, AdminCurrentSubscriberIDKey)
	if !ok {
		return "", false
	}
	return BillingID(id), true
}

// RequireAdminCurrentSubscriberID retrieves current subscriber ID or returns error
func RequireAdminCurrentSubscriberID(ctx context.Context) (BillingID, error) {
	subscriberID, ok := GetAdminCurrentSubscriberID(ctx)
	if !ok {
		return "", fmt.Errorf("admin subscriber ID not available in context")
	}
	return subscriberID, nil
}

// SetAdminManagementResult stores management result in context
func SetAdminManagementResult(ctx context.Context, result *admin.ManagementResult) context.Context {
	return SetContextValue(ctx, AdminManagementResultKey, result)
}

// GetAdminManagementResult retrieves management result from context
func GetAdminManagementResult(ctx context.Context) (*admin.ManagementResult, bool) {
	return GetContextValue[*admin.ManagementResult](ctx, AdminManagementResultKey)
}

// RequireAdminManagementResult retrieves management result or returns error
func RequireAdminManagementResult(ctx context.Context) (*admin.ManagementResult, error) {
	result, ok := GetAdminManagementResult(ctx)
	if !ok {
		return nil, fmt.Errorf("admin management result not available in context")
	}
	return result, nil
}

// SetAdminPlanChangeResult stores plan change result in context
func SetAdminPlanChangeResult(ctx context.Context, result *admin.PlanChangeResult) context.Context {
	return SetContextValue(ctx, AdminPlanChangeResultKey, result)
}

// GetAdminPlanChangeResult retrieves plan change result from context
func GetAdminPlanChangeResult(ctx context.Context) (*admin.PlanChangeResult, bool) {
	return GetContextValue[*admin.PlanChangeResult](ctx, AdminPlanChangeResultKey)
}

// RequireAdminPlanChangeResult retrieves plan change result or returns error
func RequireAdminPlanChangeResult(ctx context.Context) (*admin.PlanChangeResult, error) {
	result, ok := GetAdminPlanChangeResult(ctx)
	if !ok {
		return nil, fmt.Errorf("admin plan change result not available in context")
	}
	return result, nil
}

// AddAdminCreditCleanup adds credit ID to cleanup list
func AddAdminCreditCleanup(ctx context.Context, creditID string) context.Context {
	return AddToCleanupList(ctx, AdminCreditsCleanupKey, creditID)
}

// GetAdminCreditsCleanup retrieves credits cleanup list
func GetAdminCreditsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, AdminCreditsCleanupKey)
}

// AddAdminPriceLineCleanup adds price line ID to cleanup list
func AddAdminPriceLineCleanup(ctx context.Context, priceLineID string) context.Context {
	return AddToCleanupList(ctx, AdminPriceLinesCleanupKey, priceLineID)
}

// GetAdminPriceLinesCleanup retrieves price lines cleanup list
func GetAdminPriceLinesCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, AdminPriceLinesCleanupKey)
}

// AddAdminPricingPlanCleanup adds pricing plan ID to cleanup list
func AddAdminPricingPlanCleanup(ctx context.Context, planID string) context.Context {
	return AddToCleanupList(ctx, AdminPricingPlansCleanupKey, planID)
}

// GetAdminPricingPlansCleanup retrieves pricing plans cleanup list
func GetAdminPricingPlansCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, AdminPricingPlansCleanupKey)
}

// AddAdminPricingPlanPeriodCleanup adds period ID to cleanup list
func AddAdminPricingPlanPeriodCleanup(ctx context.Context, periodID string) context.Context {
	return AddToCleanupList(ctx, AdminPricingPlanPeriodsCleanupKey, periodID)
}

// GetAdminPricingPlanPeriodsCleanup retrieves periods cleanup list
func GetAdminPricingPlanPeriodsCleanup(ctx context.Context) []string {
	return GetCleanupList[string](ctx, AdminPricingPlanPeriodsCleanupKey)
}

// deleteWithIgnoreNotFound deletes a resource, ignoring "not found" errors
func deleteWithIgnoreNotFound(ctx context.Context, billingService *admin.BillingService, deleteFunc func(ctx context.Context, id string) error, id string) {
	err := deleteFunc(ctx, id)
	if err != nil && isNotFoundError(err) {
		return
	}
	// Other errors are logged but ignored during cleanup
}

// CleanupAdminBilling cleans up billing resources created during testing
func CleanupAdminBilling(ctx context.Context) error {
	// Get the cleanup lists - they will be empty for non-billing scenarios
	credits := GetAdminCreditsCleanup(ctx)
	priceLines := GetAdminPriceLinesCleanup(ctx)
	plans := GetAdminPricingPlansCleanup(ctx)
	periods := GetAdminPricingPlanPeriodsCleanup(ctx)

	// If nothing to clean up, return early
	if len(credits) == 0 && len(priceLines) == 0 && len(plans) == 0 && len(periods) == 0 {
		return nil
	}

	// Only try to use admin client if we have something to clean up
	adminClient, err := RequireAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("admin client not available for cleanup: %w", err)
	}

	billingService := adminClient.Billing()
	if billingService == nil {
		return fmt.Errorf("admin client's Billing() returned nil")
	}

	// Clean up credits (purge deleted ones first, then clean up remaining)
	for _, creditID := range credits {
		deleteWithIgnoreNotFound(ctx, billingService, billingService.DeleteCredit, creditID)
	}

	// Clean up price lines
	for _, priceLineID := range priceLines {
		deleteWithIgnoreNotFound(ctx, billingService, billingService.DeletePriceLine, priceLineID)
	}

	// Clean up pricing plan periods first (they must be deleted before plans)
	for _, periodID := range periods {
		deleteWithIgnoreNotFound(ctx, billingService, billingService.DeletePricingPlanPeriod, periodID)
	}

	// Clean up pricing plans
	for _, planID := range plans {
		deleteWithIgnoreNotFound(ctx, billingService, billingService.DeletePricingPlan, planID)
	}

	return nil
}


