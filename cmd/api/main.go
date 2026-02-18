package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cheoscafe/backend/internal/config"
	"github.com/cheoscafe/backend/internal/database"
	"github.com/cheoscafe/backend/internal/handlers"
	"github.com/cheoscafe/backend/internal/middleware"
	"github.com/cheoscafe/backend/internal/repository"
	"github.com/cheoscafe/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.GoEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Firebase connection
	firebaseClient, err := database.NewFirebaseConnection(cfg)
	if err != nil {
		logger.Fatalf("Failed to connect to Firebase: %v", err)
	}
	defer firebaseClient.Close()

	// Initialize Redis connection
	var redisClient database.RedisClient
	redisClient, err = database.NewRedisConnection(cfg)
	if err != nil {
		logger.Warnf("Failed to connect to Redis: %v (continuing without cache)", err)
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("Redis connected successfully")
	}

	// Test Firebase connection
	if firebaseClient.Firestore == nil {
		logger.Fatalf("Failed to initialize Firestore client")
	}
	logger.Info("Firebase connected successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(firebaseClient)
	productRepo := repository.NewProductRepository(firebaseClient)
	orderRepo := repository.NewOrderRepository(firebaseClient)
	discountRepo := repository.NewDiscountRepository(firebaseClient)
	reviewRepo := repository.NewReviewRepository(firebaseClient)
	locationRepo := repository.NewLocationRepository(firebaseClient)
	galleryRepo := repository.NewGalleryRepository(firebaseClient)
	siteConfigRepo := repository.NewSiteConfigRepository(firebaseClient)
	cartRepo := repository.NewCartRepository(firebaseClient)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	productService := services.NewProductService(productRepo)
	cartService := services.NewCartService(cartRepo, productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo, cartRepo)
	discountService := services.NewDiscountService(discountRepo)
	reviewService := services.NewReviewService(reviewRepo, productRepo)
	locationService := services.NewLocationService(locationRepo)
	galleryService := services.NewGalleryService(galleryRepo)
	siteConfigService := services.NewSiteConfigService(siteConfigRepo)

	// Initialize upload service (Cloudinary)
	uploadService, err := services.NewUploadService(cfg)
	if err != nil {
		logger.Warnf("Failed to initialize Cloudinary: %v (image upload will not be available)", err)
		uploadService = nil
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)
	discountHandler := handlers.NewDiscountHandler(discountService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	locationHandler := handlers.NewLocationHandler(locationService)
	galleryHandler := handlers.NewGalleryHandler(galleryService, uploadService)
	siteConfigHandler := handlers.NewSiteConfigHandler(siteConfigService)
	cartHandler := handlers.NewCartHandler(cartService)

	// Initialize Gin router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORS(cfg))

	// Apply rate limiting
	rateLimitDuration, _ := time.ParseDuration(cfg.RateLimitDuration)
	router.Use(middleware.RateLimiter(cfg.RateLimitRequests, rateLimitDuration))

	// Setup routes
	setupRoutes(router, cfg, firebaseClient, redisClient, authHandler, productHandler, orderHandler, discountHandler, reviewHandler, locationHandler, galleryHandler, siteConfigHandler, cartHandler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Starting server on port %s", cfg.Port)
		logger.Infof("Environment: %s", cfg.GoEnv)
		logger.Infof("API Version: %s", cfg.APIVersion)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited successfully")
}

func setupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	firebaseClient *database.FirebaseClient,
	redis database.RedisClient,
	authHandler *handlers.AuthHandler,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
	discountHandler *handlers.DiscountHandler,
	reviewHandler *handlers.ReviewHandler,
	locationHandler *handlers.LocationHandler,
	galleryHandler *handlers.GalleryHandler,
	siteConfigHandler *handlers.SiteConfigHandler,
	cartHandler *handlers.CartHandler,
	logger *logrus.Logger,
) {
	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Cheos Café Backend API is running",
			"version": cfg.APIVersion,
		})
	})

	router.GET("/health/firebase", func(c *gin.Context) {
		if firebaseClient.Firestore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Firebase connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Firebase is healthy",
		})
	})

	router.GET("/health/redis", func(c *gin.Context) {
		if redis == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Redis not configured",
			})
			return
		}
		if err := redis.Ping(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Redis connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Redis is healthy",
		})
	})

	// API v1 routes
	v1 := router.Group(fmt.Sprintf("/api/%s", cfg.APIVersion))
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", middleware.LoginRateLimiter(), authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// User profile routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		{
			users.GET("/me", authHandler.GetProfile)
			users.PUT("/me", authHandler.UpdateProfile)

			// Admin only routes
			adminUsers := users.Group("")
			adminUsers.Use(middleware.RequireAdmin())
			{
				adminUsers.GET("", authHandler.GetAllUsers)
				adminUsers.PUT("/:id", authHandler.UpdateUserByID)
				adminUsers.DELETE("/:id", authHandler.DeleteUser)
			}
		}

		// Product routes (public read, admin write)
		products := v1.Group("/products")
		{
			// Public routes
			products.GET("", productHandler.GetAllProducts)
			products.GET("/featured", productHandler.GetFeaturedProducts)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/:id", productHandler.GetProduct)

			// Admin only routes
			adminProducts := products.Group("")
			adminProducts.Use(middleware.AuthMiddleware(cfg))
			adminProducts.Use(middleware.RequireAdmin())
			{
				adminProducts.POST("", productHandler.CreateProduct)
				adminProducts.PUT("/:id", productHandler.UpdateProduct)
				adminProducts.DELETE("/:id", productHandler.DeleteProduct)
				adminProducts.PATCH("/:id/stock", productHandler.UpdateStock)
			}
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			// Public routes (guest checkout)
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/number/:number", orderHandler.GetOrderByNumber)

			// Authenticated user routes
			userOrders := orders.Group("")
			userOrders.Use(middleware.AuthMiddleware(cfg))
			{
				userOrders.GET("/me", orderHandler.GetUserOrders)
				userOrders.GET("/:id", orderHandler.GetOrder)
			}

			// Admin only routes
			adminOrders := orders.Group("")
			adminOrders.Use(middleware.AuthMiddleware(cfg))
			adminOrders.Use(middleware.RequireAdmin())
			{
				adminOrders.GET("", orderHandler.GetAllOrders)
				adminOrders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
				adminOrders.PATCH("/:id/payment", orderHandler.UpdatePaymentStatus)
			}
		}

		// Discount Code routes
		discounts := v1.Group("/discounts")
		{
			// Public route (validate discount code)
			discounts.POST("/validate", discountHandler.ValidateDiscountCode)

			// Admin only routes
			adminDiscounts := discounts.Group("")
			adminDiscounts.Use(middleware.AuthMiddleware(cfg))
			adminDiscounts.Use(middleware.RequireAdmin())
			{
				adminDiscounts.GET("", discountHandler.GetAllDiscountCodes)
				adminDiscounts.POST("", discountHandler.CreateDiscountCode)
				adminDiscounts.GET("/:id", discountHandler.GetDiscountCode)
				adminDiscounts.PUT("/:id", discountHandler.UpdateDiscountCode)
				adminDiscounts.DELETE("/:id", discountHandler.DeleteDiscountCode)
			}
		}

		// Review routes
		reviews := v1.Group("/reviews")
		{
			// Public route (create review)
			reviews.POST("", reviewHandler.CreateReview)

			// Admin only routes
			adminReviews := reviews.Group("")
			adminReviews.Use(middleware.AuthMiddleware(cfg))
			adminReviews.Use(middleware.RequireAdmin())
			{
				adminReviews.GET("", reviewHandler.GetAllReviews)
				adminReviews.GET("/:id", reviewHandler.GetReview)
				adminReviews.PUT("/:id", reviewHandler.UpdateReview)
				adminReviews.DELETE("/:id", reviewHandler.DeleteReview)
			}
		}

		// Product reviews (public)
		products.GET("/:id/reviews", reviewHandler.GetProductReviews)

		// Location routes
		locations := v1.Group("/locations")
		{
			// Public routes
			locations.GET("", locationHandler.GetActiveLocations)
			locations.GET("/all", locationHandler.GetAllLocations)
			locations.GET("/:id", locationHandler.GetLocation)

			// Admin only routes
			adminLocations := locations.Group("")
			adminLocations.Use(middleware.AuthMiddleware(cfg))
			adminLocations.Use(middleware.RequireAdmin())
			{
				adminLocations.POST("", locationHandler.CreateLocation)
				adminLocations.PUT("/:id", locationHandler.UpdateLocation)
				adminLocations.DELETE("/:id", locationHandler.DeleteLocation)
			}
		}

		// Gallery routes
		gallery := v1.Group("/gallery")
		{
			// Public routes
			gallery.GET("/active", galleryHandler.GetActiveImages)
			gallery.GET("/type/:type", galleryHandler.GetImagesByType)
			gallery.GET("/:id", galleryHandler.GetImage)

			// Admin only routes
			adminGallery := gallery.Group("")
			adminGallery.Use(middleware.AuthMiddleware(cfg))
			adminGallery.Use(middleware.RequireAdmin())
			{
				adminGallery.GET("", galleryHandler.GetAllImages)
				adminGallery.POST("", galleryHandler.CreateImage)
				adminGallery.POST("/upload", galleryHandler.UploadImage)
				adminGallery.PUT("/:id", galleryHandler.UpdateImage)
				adminGallery.DELETE("/:id", galleryHandler.DeleteImage)
			}
		}

		// Site Config routes (carousel, etc.)
		siteConfig := v1.Group("/config")
		{
			// Public: obtener carrusel y about us
			siteConfig.GET("/carousel", siteConfigHandler.GetCarousel)
			siteConfig.GET("/about", siteConfigHandler.GetAboutUs)

			// Admin: actualizar carrusel y about us
			adminConfig := siteConfig.Group("")
			adminConfig.Use(middleware.AuthMiddleware(cfg))
			adminConfig.Use(middleware.RequireAdmin())
			{
				adminConfig.PUT("/carousel", siteConfigHandler.UpdateCarousel)
				adminConfig.PUT("/about", siteConfigHandler.UpdateAboutUs)
			}
		}

		// Cart routes (authenticated users only)
		cart := v1.Group("/cart")
		cart.Use(middleware.AuthMiddleware(cfg))
		{
			cart.GET("", cartHandler.GetCart)
			cart.POST("/items", cartHandler.AddItem)
			cart.PUT("/items/:productId", cartHandler.UpdateItemQuantity)
			cart.DELETE("/items/:productId", cartHandler.RemoveItem)
			cart.DELETE("", cartHandler.ClearCart)
			cart.POST("/sync", cartHandler.SyncCart)
		}

		// Test endpoint
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	logger.Info("Routes configured successfully")
}
