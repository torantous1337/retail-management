package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/torantous1337/retail-management/internal/core/domain"
	"github.com/torantous1337/retail-management/internal/core/ports"
)

// ProductService implements the product business logic.
type ProductService struct {
	productRepo  ports.ProductRepository
	categoryRepo ports.CategoryRepository
	auditSvc     ports.AuditService
	txManager    ports.TransactionManager
}

// NewProductService creates a new product service instance.
func NewProductService(productRepo ports.ProductRepository, categoryRepo ports.CategoryRepository, auditSvc ports.AuditService, txManager ports.TransactionManager) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		auditSvc:     auditSvc,
		txManager:    txManager,
	}
}

// validateProductProperties fetches the category and validates product properties.
func (s *ProductService) validateProductProperties(ctx context.Context, product *domain.Product) error {
	if product.CategoryID == "" {
		return nil
	}

	category, err := s.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return err
	}

	return ValidateProperties(category, product.Properties)
}

// CreateProduct creates a new product and logs the action.
func (s *ProductService) CreateProduct(ctx context.Context, product *domain.Product) error {
	// Validate properties against category blueprint
	if err := s.validateProductProperties(ctx, product); err != nil {
		return err
	}

	err := s.productRepo.Create(ctx, product)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product creation but should be logged)
	payload := map[string]interface{}{
		"product_id": product.ID,
		"action":     "create_product",
		"sku":        product.SKU,
		"name":       product.Name,
	}
	if err := s.auditSvc.LogAction(ctx, "CREATE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

// GetProductBySKU retrieves a product by SKU.
func (s *ProductService) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

// ListProducts retrieves all products with pagination.
func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	return s.productRepo.List(ctx, limit, offset)
}

// SearchProducts retrieves products matching the given filter options.
// Property filter keys are validated against the category blueprint when a
// CategoryID is provided, preventing SQL injection via JSON paths.
func (s *ProductService) SearchProducts(ctx context.Context, opts domain.FilterOptions) ([]*domain.Product, error) {
	var allowedKeys []string
	if opts.CategoryID != "" {
		category, err := s.categoryRepo.GetByID(ctx, opts.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("category lookup: %w", err)
		}
		for _, attr := range category.AttributeDefinitions {
			allowedKeys = append(allowedKeys, attr.Key)
		}
	}
	return s.productRepo.Search(ctx, opts, allowedKeys)
}

// UpdateProduct updates a product and logs the action.
func (s *ProductService) UpdateProduct(ctx context.Context, product *domain.Product) error {
	// Validate properties against category blueprint
	if err := s.validateProductProperties(ctx, product); err != nil {
		return err
	}

	err := s.productRepo.Update(ctx, product)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product update but should be logged)
	payload := map[string]interface{}{
		"product_id": product.ID,
		"action":     "update_product",
		"sku":        product.SKU,
	}
	if err := s.auditSvc.LogAction(ctx, "UPDATE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}

// DeleteProduct deletes a product and logs the action.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	err := s.productRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Log the action (audit failures should not prevent product deletion but should be logged)
	payload := map[string]interface{}{
		"product_id": id,
		"action":     "delete_product",
	}
	if err := s.auditSvc.LogAction(ctx, "DELETE_PRODUCT", "system", payload); err != nil {
		// TODO: Log this error to a monitoring system
		// For now, we don't fail the operation but the error should be tracked
	}

	return nil
}

// ImportProducts imports products from a CSV reader within a single transaction.
// CSV must have a header row. The columns "name", "sku", and "base_price" are required.
// Additional columns are mapped to product properties using the header as the key.
func (s *ProductService) ImportProducts(ctx context.Context, categoryID string, csvReader io.Reader) (int, error) {
	// Fetch category once (outside the transaction) for validation.
	var category *domain.Category
	if categoryID != "" {
		var err error
		category, err = s.categoryRepo.GetByID(ctx, categoryID)
		if err != nil {
			return 0, fmt.Errorf("category lookup: %w", err)
		}
	}

	reader := csv.NewReader(csvReader)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read CSV header: %w", err)
	}

	// Build column index
	colIndex := make(map[string]int, len(headers))
	for i, h := range headers {
		colIndex[h] = i
	}

	// Require mandatory columns
	for _, required := range []string{"name", "sku", "base_price"} {
		if _, ok := colIndex[required]; !ok {
			return 0, fmt.Errorf("missing required CSV column: %s", required)
		}
	}

	// Pre-read all rows to avoid holding a transaction open during I/O.
	var allRows [][]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("read CSV row: %w", err)
		}
		allRows = append(allRows, row)
	}

	if len(allRows) == 0 {
		return 0, nil
	}

	var imported int

	err = s.txManager.WithTx(ctx, func(tx ports.Ports) error {
		// Track prev_hash for sequential audit chain within the transaction
		lastLog, err := tx.AuditRepo.GetLastLog(ctx)
		prevHash := ""
		if err == nil && lastLog != nil {
			prevHash = lastLog.CurrentHash
		}

		for lineNum, row := range allRows {
			if len(row) != len(headers) {
				return fmt.Errorf("CSV line %d: expected %d columns, got %d", lineNum+2, len(headers), len(row))
			}

			name := row[colIndex["name"]]
			sku := row[colIndex["sku"]]
			basePriceStr := row[colIndex["base_price"]]

			if name == "" || sku == "" {
				return fmt.Errorf("CSV line %d: name and sku are required", lineNum+2)
			}

			basePrice, err := strconv.ParseFloat(basePriceStr, 64)
			if err != nil {
				return fmt.Errorf("CSV line %d: invalid base_price %q: %w", lineNum+2, basePriceStr, err)
			}

			// Build properties from extra columns
			properties := make(map[string]interface{})
			for header, idx := range colIndex {
				if header == "name" || header == "sku" || header == "base_price" {
					continue
				}
				properties[header] = row[idx]
			}

			// Validate against category blueprint
			if category != nil {
				if err := ValidateProperties(category, properties); err != nil {
					return fmt.Errorf("CSV line %d: %w", lineNum+2, err)
				}
			}

			now := time.Now()
			product := &domain.Product{
				ID:         uuid.New().String(),
				Name:       name,
				SKU:        sku,
				CategoryID: categoryID,
				BasePrice:  basePrice,
				Properties: properties,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := tx.ProductRepo.Create(ctx, product); err != nil {
				return fmt.Errorf("CSV line %d: insert product: %w", lineNum+2, err)
			}

			// Create audit log inside the same transaction
			txAuditSvc := NewAuditService(tx.AuditRepo)
			txAuditSvc.SetPrevHash(prevHash)
			if err := txAuditSvc.LogAction(ctx, "CREATE_PRODUCT", "system", map[string]interface{}{
				"product_id": product.ID,
				"action":     "import_product",
				"sku":        product.SKU,
				"name":       product.Name,
			}); err != nil {
				return fmt.Errorf("CSV line %d: audit log: %w", lineNum+2, err)
			}
			prevHash = txAuditSvc.LastHash()

			imported++
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return imported, nil
}
