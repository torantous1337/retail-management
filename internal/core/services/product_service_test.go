package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// --- Mock implementations for testing ---

type mockProductRepository struct {
	products []*domain.Product
}

func (m *mockProductRepository) Create(_ context.Context, product *domain.Product) error {
	for _, p := range m.products {
		if p.SKU == product.SKU {
			return errors.New("duplicate sku")
		}
	}
	m.products = append(m.products, product)
	return nil
}
func (m *mockProductRepository) GetByID(_ context.Context, id string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("product not found")
}
func (m *mockProductRepository) GetBySKU(_ context.Context, sku string) (*domain.Product, error) {
	for _, p := range m.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return nil, errors.New("product not found")
}
func (m *mockProductRepository) List(_ context.Context, _, _ int) ([]*domain.Product, error) {
	return m.products, nil
}
func (m *mockProductRepository) Update(_ context.Context, _ *domain.Product) error { return nil }
func (m *mockProductRepository) Delete(_ context.Context, _ string) error          { return nil }

type mockCategoryRepository struct {
	categories map[string]*domain.Category
}

func (m *mockCategoryRepository) Create(_ context.Context, c *domain.Category) error {
	m.categories[c.ID] = c
	return nil
}
func (m *mockCategoryRepository) GetByID(_ context.Context, id string) (*domain.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	return c, nil
}
func (m *mockCategoryRepository) List(_ context.Context, _, _ int) ([]*domain.Category, error) {
	var out []*domain.Category
	for _, c := range m.categories {
		out = append(out, c)
	}
	return out, nil
}

type mockAuditLogRepository struct {
	logs []*domain.AuditLog
}

func (m *mockAuditLogRepository) Create(_ context.Context, log *domain.AuditLog) error {
	log.ID = int64(len(m.logs) + 1)
	m.logs = append(m.logs, log)
	return nil
}
func (m *mockAuditLogRepository) GetLastLog(_ context.Context) (*domain.AuditLog, error) {
	if len(m.logs) == 0 {
		return nil, nil
	}
	return m.logs[len(m.logs)-1], nil
}
func (m *mockAuditLogRepository) List(_ context.Context, _, _ int) ([]*domain.AuditLog, error) {
	return m.logs, nil
}
func (m *mockAuditLogRepository) VerifyChain(_ context.Context) (bool, error) {
	return true, nil
}

type mockTransactionManager struct {
	productRepo  *mockProductRepository
	categoryRepo *mockCategoryRepository
	auditRepo    *mockAuditLogRepository
}

func (m *mockTransactionManager) WithTx(_ context.Context, fn func(tx ports.Ports) error) error {
	txPorts := ports.Ports{
		ProductRepo:  m.productRepo,
		CategoryRepo: m.categoryRepo,
		AuditRepo:    m.auditRepo,
	}
	return fn(txPorts)
}

// --- Tests ---

func TestImportProducts_BasicCSV(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	csv := "name,sku,base_price\nWidget A,SKU-001,9.99\nWidget B,SKU-002,19.99\n"
	count, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}
	if len(productRepo.products) != 2 {
		t.Fatalf("expected 2 products in repo, got %d", len(productRepo.products))
	}
	if len(auditRepo.logs) != 2 {
		t.Fatalf("expected 2 audit logs, got %d", len(auditRepo.logs))
	}
}

func TestImportProducts_WithCategoryValidation(t *testing.T) {
	productRepo := &mockProductRepository{}
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
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// CSV includes the voltage column
	csv := "name,sku,base_price,voltage\nWire,SKU-100,5.00,220V\n"
	count, err := svc.ImportProducts(context.Background(), "cat-1", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 imported, got %d", count)
	}
}

func TestImportProducts_CategoryValidationFails(t *testing.T) {
	productRepo := &mockProductRepository{}
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
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// CSV missing required "voltage" column
	csv := "name,sku,base_price\nWire,SKU-100,5.00\n"
	_, err := svc.ImportProducts(context.Background(), "cat-1", strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing required property")
	}
	if !strings.Contains(err.Error(), "voltage") {
		t.Fatalf("expected error to mention 'voltage', got: %v", err)
	}
}

func TestImportProducts_MissingRequiredColumn(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// Missing "sku" column
	csv := "name,base_price\nWidget,9.99\n"
	_, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing sku column")
	}
	if !strings.Contains(err.Error(), "missing required CSV column") {
		t.Fatalf("expected missing column error, got: %v", err)
	}
}

func TestImportProducts_InvalidBasePrice(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	csv := "name,sku,base_price\nWidget,SKU-001,not_a_number\n"
	_, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for invalid base_price")
	}
	if !strings.Contains(err.Error(), "invalid base_price") {
		t.Fatalf("expected base_price error, got: %v", err)
	}
}

func TestImportProducts_EmptyCSV(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// Header only, no data rows
	csv := "name,sku,base_price\n"
	count, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 imported, got %d", count)
	}
}

func TestImportProducts_AuditChainSequential(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	csv := "name,sku,base_price\nA,SKU-A,1.00\nB,SKU-B,2.00\nC,SKU-C,3.00\n"
	count, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 imported, got %d", count)
	}

	// Verify audit chain: each log's PrevHash should equal the previous log's CurrentHash
	for i := 1; i < len(auditRepo.logs); i++ {
		if auditRepo.logs[i].PrevHash != auditRepo.logs[i-1].CurrentHash {
			t.Fatalf("audit chain broken at index %d: PrevHash=%q, expected=%q",
				i, auditRepo.logs[i].PrevHash, auditRepo.logs[i-1].CurrentHash)
		}
	}
}

func TestImportProducts_CategoryNotFound(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	csv := "name,sku,base_price\nWidget,SKU-001,9.99\n"
	_, err := svc.ImportProducts(context.Background(), "nonexistent-cat", strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for nonexistent category")
	}
	if !strings.Contains(err.Error(), "category") {
		t.Fatalf("expected category error, got: %v", err)
	}
}
