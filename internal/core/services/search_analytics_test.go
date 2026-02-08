package services

import (
	"context"
	"testing"

	"github.com/torantous1337/retail-management/internal/core/domain"
)

// --- Enhanced mock with search/analytics support ---

type searchMockProductRepository struct {
	products []*domain.Product
}

func (m *searchMockProductRepository) Create(_ context.Context, product *domain.Product) error {
	m.products = append(m.products, product)
	return nil
}
func (m *searchMockProductRepository) GetByID(_ context.Context, id string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}
func (m *searchMockProductRepository) GetBySKU(_ context.Context, sku string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return nil, nil
}
func (m *searchMockProductRepository) List(_ context.Context, _, _ int) ([]*domain.Product, error) {
	return m.products, nil
}
func (m *searchMockProductRepository) Update(_ context.Context, _ *domain.Product) error { return nil }
func (m *searchMockProductRepository) Delete(_ context.Context, _ string) error          { return nil }

func (m *searchMockProductRepository) Search(_ context.Context, opts domain.FilterOptions, allowedKeys []string) ([]*domain.Product, error) {
	var result []*domain.Product
	for _, p := range m.products {
		// Simulate category filter
		if opts.CategoryID != "" && p.CategoryID != opts.CategoryID {
			continue
		}
		// Simulate price range
		if opts.MinPrice != nil && p.BasePrice < *opts.MinPrice {
			continue
		}
		if opts.MaxPrice != nil && p.BasePrice > *opts.MaxPrice {
			continue
		}
		// Simulate property filter (only allowed keys)
		match := true
		allowed := make(map[string]bool)
		for _, k := range allowedKeys {
			allowed[k] = true
		}
		for key, val := range opts.Properties {
			if !allowed[key] {
				continue
			}
			if pv, ok := p.Properties[key]; !ok || pv != val {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *searchMockProductRepository) GetInventorySummary(_ context.Context) (*domain.InventorySummary, error) {
	summary := &domain.InventorySummary{}
	catMap := make(map[string]*domain.CategoryBreakdown)
	for _, p := range m.products {
		catID := p.CategoryID
		if _, ok := catMap[catID]; !ok {
			catMap[catID] = &domain.CategoryBreakdown{CategoryID: catID}
		}
		catMap[catID].Count++
		catMap[catID].TotalValue += p.BasePrice
		summary.TotalItems++
		summary.TotalValue += p.BasePrice
	}
	for _, bd := range catMap {
		summary.CategoryBreakdown = append(summary.CategoryBreakdown, *bd)
	}
	return summary, nil
}

// --- Tests ---

func TestSearchProducts_NoFilters(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "Widget A", SKU: "SKU-001", BasePrice: 10},
			{ID: "2", Name: "Widget B", SKU: "SKU-002", BasePrice: 20},
		},
	}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: &mockProductRepository{}, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	results, err := svc.SearchProducts(context.Background(), domain.FilterOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 products, got %d", len(results))
	}
}

func TestSearchProducts_CategoryFilter(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "Widget A", SKU: "SKU-001", CategoryID: "cat-1", BasePrice: 10},
			{ID: "2", Name: "Widget B", SKU: "SKU-002", CategoryID: "cat-2", BasePrice: 20},
		},
	}
	categoryRepo := &mockCategoryRepository{categories: map[string]*domain.Category{
		"cat-1": {ID: "cat-1", Name: "Electrical", AttributeDefinitions: []domain.AttributeDefinition{}},
	}}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: &mockProductRepository{}, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	results, err := svc.SearchProducts(context.Background(), domain.FilterOptions{CategoryID: "cat-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 product, got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Fatalf("expected product ID 1, got %s", results[0].ID)
	}
}

func TestSearchProducts_PriceRange(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "Cheap", SKU: "SKU-001", BasePrice: 5},
			{ID: "2", Name: "Medium", SKU: "SKU-002", BasePrice: 15},
			{ID: "3", Name: "Expensive", SKU: "SKU-003", BasePrice: 50},
		},
	}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: &mockProductRepository{}, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	min := 10.0
	max := 20.0
	results, err := svc.SearchProducts(context.Background(), domain.FilterOptions{
		MinPrice: &min,
		MaxPrice: &max,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 product, got %d", len(results))
	}
	if results[0].ID != "2" {
		t.Fatalf("expected product ID 2, got %s", results[0].ID)
	}
}

func TestSearchProducts_PropertyFilter(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "Wire 220", SKU: "SKU-001", CategoryID: "cat-1", BasePrice: 10, Properties: map[string]interface{}{"voltage": "220V"}},
			{ID: "2", Name: "Wire 110", SKU: "SKU-002", CategoryID: "cat-1", BasePrice: 10, Properties: map[string]interface{}{"voltage": "110V"}},
		},
	}
	categoryRepo := &mockCategoryRepository{categories: map[string]*domain.Category{
		"cat-1": {
			ID:   "cat-1",
			Name: "Electrical",
			AttributeDefinitions: []domain.AttributeDefinition{
				{Key: "voltage", Type: "string", Required: true},
			},
		},
	}}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: &mockProductRepository{}, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	results, err := svc.SearchProducts(context.Background(), domain.FilterOptions{
		CategoryID: "cat-1",
		Properties: map[string]string{"voltage": "220V"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 product, got %d", len(results))
	}
	if results[0].ID != "1" {
		t.Fatalf("expected product ID 1, got %s", results[0].ID)
	}
}

func TestSearchProducts_PropertyFilterIgnoredWithoutCategory(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "Widget", SKU: "SKU-001", BasePrice: 10, Properties: map[string]interface{}{"voltage": "220V"}},
		},
	}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: &mockProductRepository{}, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// Without category_id, no allowedKeys => property filter keys are ignored
	results, err := svc.SearchProducts(context.Background(), domain.FilterOptions{
		Properties: map[string]string{"voltage": "220V"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All products returned because no keys are allowed without category
	if len(results) != 1 {
		t.Fatalf("expected 1 product, got %d", len(results))
	}
}

func TestGetInventorySummary(t *testing.T) {
	productRepo := &searchMockProductRepository{
		products: []*domain.Product{
			{ID: "1", Name: "A", SKU: "SKU-A", CategoryID: "cat-1", BasePrice: 10},
			{ID: "2", Name: "B", SKU: "SKU-B", CategoryID: "cat-1", BasePrice: 20},
			{ID: "3", Name: "C", SKU: "SKU-C", CategoryID: "cat-2", BasePrice: 30},
		},
	}

	svc := NewAnalyticsService(productRepo)

	summary, err := svc.GetInventorySummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalItems != 3 {
		t.Fatalf("expected 3 total items, got %d", summary.TotalItems)
	}
	if summary.TotalValue != 60.0 {
		t.Fatalf("expected total value 60.0, got %f", summary.TotalValue)
	}
	if len(summary.CategoryBreakdown) != 2 {
		t.Fatalf("expected 2 category breakdowns, got %d", len(summary.CategoryBreakdown))
	}
}

func TestGetInventorySummary_Empty(t *testing.T) {
	productRepo := &searchMockProductRepository{}

	svc := NewAnalyticsService(productRepo)

	summary, err := svc.GetInventorySummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalItems != 0 {
		t.Fatalf("expected 0 total items, got %d", summary.TotalItems)
	}
	if summary.TotalValue != 0.0 {
		t.Fatalf("expected total value 0.0, got %f", summary.TotalValue)
	}
}
