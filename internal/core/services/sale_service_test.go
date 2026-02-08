package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// --- Mock SaleRepository ---

type mockSaleRepository struct {
	sales     []*domain.Sale
	saleItems []*domain.SaleItem
}

func (m *mockSaleRepository) CreateSale(_ context.Context, sale *domain.Sale) error {
	m.sales = append(m.sales, sale)
	return nil
}

func (m *mockSaleRepository) CreateSaleItem(_ context.Context, item *domain.SaleItem) error {
	m.saleItems = append(m.saleItems, item)
	return nil
}

// --- Mock TransactionManager for SaleService tests ---

type mockSaleTxManager struct {
	productRepo  *mockProductRepository
	categoryRepo *mockCategoryRepository
	auditRepo    *mockAuditLogRepository
	saleRepo     *mockSaleRepository
}

func (m *mockSaleTxManager) WithTx(_ context.Context, fn func(tx ports.Ports) error) error {
	txPorts := ports.Ports{
		ProductRepo:  m.productRepo,
		CategoryRepo: m.categoryRepo,
		AuditRepo:    m.auditRepo,
		SaleRepo:     m.saleRepo,
	}
	return fn(txPorts)
}

// --- SaleService Tests ---

func TestProcessSale_Success(t *testing.T) {
	productRepo := &mockProductRepository{
		products: []*domain.Product{
			{ID: "p1", Name: "Widget", SKU: "SKU-001", BasePrice: 10.00, CostPrice: 5.00, Quantity: 100},
			{ID: "p2", Name: "Gadget", SKU: "SKU-002", BasePrice: 20.00, CostPrice: 8.00, Quantity: 50},
		},
	}
	saleRepo := &mockSaleRepository{}
	auditRepo := &mockAuditLogRepository{}
	txManager := &mockSaleTxManager{
		productRepo:  productRepo,
		categoryRepo: &mockCategoryRepository{categories: make(map[string]*domain.Category)},
		auditRepo:    auditRepo,
		saleRepo:     saleRepo,
	}

	svc := NewSaleService(txManager)

	sale, err := svc.ProcessSale(context.Background(), []ports.SaleItemRequest{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 1},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sale == nil {
		t.Fatal("expected sale, got nil")
	}

	// Total = (10.00 * 2) + (20.00 * 1) = 40.00
	if sale.TotalAmount != 40.00 {
		t.Fatalf("expected total 40.00, got %f", sale.TotalAmount)
	}

	// Verify stock decremented
	if productRepo.products[0].Quantity != 98 {
		t.Fatalf("expected product p1 quantity 98, got %d", productRepo.products[0].Quantity)
	}
	if productRepo.products[1].Quantity != 49 {
		t.Fatalf("expected product p2 quantity 49, got %d", productRepo.products[1].Quantity)
	}

	// Verify sale items created with price snapshots
	if len(saleRepo.saleItems) != 2 {
		t.Fatalf("expected 2 sale items, got %d", len(saleRepo.saleItems))
	}
	if saleRepo.saleItems[0].UnitPrice != 10.00 {
		t.Fatalf("expected unit price 10.00, got %f", saleRepo.saleItems[0].UnitPrice)
	}
	if saleRepo.saleItems[0].CostPrice != 5.00 {
		t.Fatalf("expected cost price 5.00, got %f", saleRepo.saleItems[0].CostPrice)
	}

	// Verify audit log created
	if len(auditRepo.logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(auditRepo.logs))
	}
	if auditRepo.logs[0].Action != "SALE_PROCESSED" {
		t.Fatalf("expected SALE_PROCESSED action, got %s", auditRepo.logs[0].Action)
	}
}

func TestProcessSale_InsufficientStock(t *testing.T) {
	productRepo := &mockProductRepository{
		products: []*domain.Product{
			{ID: "p1", Name: "Widget", SKU: "SKU-001", BasePrice: 10.00, Quantity: 5},
		},
	}
	saleRepo := &mockSaleRepository{}
	auditRepo := &mockAuditLogRepository{}
	txManager := &mockSaleTxManager{
		productRepo:  productRepo,
		categoryRepo: &mockCategoryRepository{categories: make(map[string]*domain.Category)},
		auditRepo:    auditRepo,
		saleRepo:     saleRepo,
	}

	svc := NewSaleService(txManager)

	_, err := svc.ProcessSale(context.Background(), []ports.SaleItemRequest{
		{ProductID: "p1", Quantity: 10},
	})

	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got: %v", err)
	}
}

func TestProcessSale_ProductNotFound(t *testing.T) {
	productRepo := &mockProductRepository{}
	saleRepo := &mockSaleRepository{}
	auditRepo := &mockAuditLogRepository{}
	txManager := &mockSaleTxManager{
		productRepo:  productRepo,
		categoryRepo: &mockCategoryRepository{categories: make(map[string]*domain.Category)},
		auditRepo:    auditRepo,
		saleRepo:     saleRepo,
	}

	svc := NewSaleService(txManager)

	_, err := svc.ProcessSale(context.Background(), []ports.SaleItemRequest{
		{ProductID: "nonexistent", Quantity: 1},
	})

	if err == nil {
		t.Fatal("expected error for nonexistent product")
	}
}

func TestProcessSale_EmptyItems(t *testing.T) {
	txManager := &mockSaleTxManager{
		productRepo:  &mockProductRepository{},
		categoryRepo: &mockCategoryRepository{categories: make(map[string]*domain.Category)},
		auditRepo:    &mockAuditLogRepository{},
		saleRepo:     &mockSaleRepository{},
	}

	svc := NewSaleService(txManager)

	_, err := svc.ProcessSale(context.Background(), []ports.SaleItemRequest{})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestImportProducts_WithQuantityAndCostPrice(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	csv := "name,sku,base_price,quantity,cost_price\nWidget A,SKU-001,9.99,100,5.00\nWidget B,SKU-002,19.99,50,10.00\n"
	count, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	// Verify quantity and cost_price were set
	if productRepo.products[0].Quantity != 100 {
		t.Fatalf("expected quantity 100, got %d", productRepo.products[0].Quantity)
	}
	if productRepo.products[0].CostPrice != 5.00 {
		t.Fatalf("expected cost_price 5.00, got %f", productRepo.products[0].CostPrice)
	}
	if productRepo.products[1].Quantity != 50 {
		t.Fatalf("expected quantity 50, got %d", productRepo.products[1].Quantity)
	}
	if productRepo.products[1].CostPrice != 10.00 {
		t.Fatalf("expected cost_price 10.00, got %f", productRepo.products[1].CostPrice)
	}
}

func TestImportProducts_WithoutQuantityAndCostPrice(t *testing.T) {
	productRepo := &mockProductRepository{}
	categoryRepo := &mockCategoryRepository{categories: make(map[string]*domain.Category)}
	auditRepo := &mockAuditLogRepository{}
	auditSvc := NewAuditService(auditRepo)
	txManager := &mockTransactionManager{productRepo: productRepo, categoryRepo: categoryRepo, auditRepo: auditRepo}

	svc := NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// CSV without quantity/cost_price columns - should default to 0
	csv := "name,sku,base_price\nWidget A,SKU-001,9.99\n"
	count, err := svc.ImportProducts(context.Background(), "", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 imported, got %d", count)
	}

	if productRepo.products[0].Quantity != 0 {
		t.Fatalf("expected quantity 0 (default), got %d", productRepo.products[0].Quantity)
	}
	if productRepo.products[0].CostPrice != 0.0 {
		t.Fatalf("expected cost_price 0.0 (default), got %f", productRepo.products[0].CostPrice)
	}
}
