package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/torantous1337/retail-management/internal/adapters/handler"
	"github.com/torantous1337/retail-management/internal/adapters/storage"
	"github.com/torantous1337/retail-management/internal/core/services"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "retail.db"
	}

	// Initialize database
	db, err := storage.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized at: %s", dbPath)

	// Initialize repositories
	productRepo := storage.NewProductRepository(db)
	auditRepo := storage.NewAuditLogRepository(db)
	categoryRepo := storage.NewCategoryRepository(db)
	txManager := storage.NewSQLTransactionManager(db)

	// Initialize services (Clean Architecture: Services depend on Repository interfaces)
	auditSvc := services.NewAuditService(auditRepo)
	categorySvc := services.NewCategoryService(categoryRepo)
	productSvc := services.NewProductService(productRepo, categoryRepo, auditSvc, txManager)

	// Initialize HTTP handlers
	productHandler := handler.NewProductHandler(productSvc)
	auditHandler := handler.NewAuditHandler(auditSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Retail Management System v1.0",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"system": "Retail Management System",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Product routes
	products := api.Group("/products")
	products.Post("/", productHandler.CreateProduct)
	products.Post("/import", productHandler.ImportProducts)
	products.Get("/", productHandler.ListProducts)
	products.Get("/:id", productHandler.GetProduct)
	products.Get("/sku/:sku", productHandler.GetProductBySKU)
	products.Put("/:id", productHandler.UpdateProduct)
	products.Delete("/:id", productHandler.DeleteProduct)

	// Category routes
	categories := api.Group("/categories")
	categories.Post("/", categoryHandler.CreateCategory)
	categories.Get("/", categoryHandler.ListCategories)

	// Audit log routes
	audit := api.Group("/audit-logs")
	audit.Get("/", auditHandler.ListAuditLogs)
	audit.Get("/verify", auditHandler.VerifyAuditChain)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
