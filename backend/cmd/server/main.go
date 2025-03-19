package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourusername/codehawk/backend/internal/db"
	"github.com/yourusername/codehawk/backend/internal/repository"
	"github.com/yourusername/codehawk/backend/internal/service"
	"github.com/yourusername/codehawk/backend/pkg/ai"
	"github.com/yourusername/codehawk/backend/pkg/analyzer"
)

// Application configuration
type Config struct {
	Port           int
	ApiKey         string
	LogLevel       string
	DbConfig       db.Config
	AiEnabled      bool
	AiEndpoint     string
	AiApiKey       string
	AiModel        string
	AiProvider     string
	CachingEnabled bool
}

// Global services
var (
	linterRegistry *analyzer.LinterRegistry
	analysisService *service.AnalysisService
	userRepository repository.UserRepository
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	config := loadConfig()

	// Initialize components
	dbConn, err := setupDatabase(config.DbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	analysisRepo := repository.NewPostgresAnalysisRepository(dbConn)
	userRepository = repository.NewPostgresUserRepository(dbConn)

	// Initialize linter registry
	linterRegistry = analyzer.NewLinterRegistry()
	linterRegistry.RegisterDefaultLinters()

	// Initialize AI service if enabled
	var aiService ai.AISuggestionService
	if config.AiEnabled {
		aiService = setupAIService(config)
	}

	// Initialize analysis service
	analysisService = service.NewAnalysisService(
		linterRegistry,
		analysisRepo,
		aiService,
		config.AiEnabled,
	)

	// Set up the server
	router := setupRouter(config)
	
	serverAddr := fmt.Sprintf(":%d", config.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting CodeHawk API server on port %d\n", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

// Load configuration from environment variables
func loadConfig() Config {
	config := Config{
		Port:     8080,
		ApiKey:   "demo_api_key_123",
		LogLevel: "info",
		DbConfig: db.NewDefaultConfig(),
	}

	// Load from environment variables
	if port, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		config.Port = port
	}

	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		config.ApiKey = apiKey
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Database config
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.DbConfig.Host = dbHost
	}

	if dbPort, err := strconv.Atoi(os.Getenv("DB_PORT")); err == nil {
		config.DbConfig.Port = dbPort
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.DbConfig.User = dbUser
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.DbConfig.Password = dbPassword
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.DbConfig.Database = dbName
	}

	if dbSSLMode := os.Getenv("DB_SSL_MODE"); dbSSLMode != "" {
		config.DbConfig.SSLMode = dbSSLMode
	}

	// AI Service configuration
	if aiEnabled, err := strconv.ParseBool(os.Getenv("AI_ENABLED")); err == nil {
		config.AiEnabled = aiEnabled
	}

	if aiEndpoint := os.Getenv("AI_ENDPOINT"); aiEndpoint != "" {
		config.AiEndpoint = aiEndpoint
	}

	if aiApiKey := os.Getenv("AI_API_KEY"); aiApiKey != "" {
		config.AiApiKey = aiApiKey
	}

	if aiModel := os.Getenv("AI_MODEL"); aiModel != "" {
		config.AiModel = aiModel
	}

	if aiProvider := os.Getenv("AI_PROVIDER"); aiProvider != "" {
		config.AiProvider = aiProvider
	}

	if cachingEnabled, err := strconv.ParseBool(os.Getenv("CACHING_ENABLED")); err == nil {
		config.CachingEnabled = cachingEnabled
	} else {
		config.CachingEnabled = true // Default to enabled
	}

	return config
}

// Set up database connection
func setupDatabase(config db.Config) (*sqlx.DB, error) {
	return db.Connect(config)
}

// Set up AI service
func setupAIService(config Config) ai.AISuggestionService {
	// Create base AI service configuration
	aiConfig := ai.NewDefaultConfig()
	
	// Override with configuration
	if config.AiEndpoint != "" {
		aiConfig.Endpoint = config.AiEndpoint
	}
	
	if config.AiApiKey != "" {
		aiConfig.APIKey = config.AiApiKey
	}
	
	if config.AiModel != "" {
		aiConfig.Model = config.AiModel
	}
	
	aiConfig.EnableCaching = config.CachingEnabled

	// Create the provider
	provider := ai.GetProviderFromString(config.AiProvider)
	
	// Create the base service
	baseService := ai.NewAIService(aiConfig, provider)
	
	// Add retry capability
	retryConfig := ai.NewDefaultRetryConfig()
	retryableService := ai.NewRetryableAIService(baseService, retryConfig)
	
	// Add caching if enabled
	if config.CachingEnabled {
		cachedService := ai.NewCachedAIService(retryableService, aiConfig)
		
		// Start cache cleanup
		if cachingService, ok := cachedService.(*ai.CachedAIService); ok {
			cachingService.StartCacheCleanup(10 * time.Minute)
		}
		
		return cachedService
	}
	
	return retryableService
}

// Set up the HTTP router
func setupRouter(config Config) *gin.Engine {
	// Set Gin mode
	if config.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	
	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	
	// API key middleware for all routes except health check
	router.Use(func(c *gin.Context) {
		// Skip API key check for health check endpoint
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		
		// Get API key
		apiKey := config.ApiKey
		
		// Check if request has the API key
		requestApiKey := c.GetHeader("X-API-Key")
		if requestApiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "API key is required",
			})
			c.Abort()
			return
		}
		
		// Validate API key
		if requestApiKey != apiKey {
			// Check if it's a user API key
			if user, err := userRepository.GetUserByAPIKey(c.Request.Context(), requestApiKey); err != nil || user == nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  "error",
					"message": "Invalid API key",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	})
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "UP",
			"time":     time.Now().Format(time.RFC3339),
			"version":  "0.1.0",
			"linters":  linterRegistry.GetSupportedLanguages(),
			"ai":       config.AiEnabled,
		})
	})
	
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Code analysis endpoint
		v1.POST("/analyze", handleAnalyzeCode)
		
		// Get analysis results by ID
		v1.GET("/analysis/:id", handleGetAnalysisById)
		
		// Get language-specific rules
		v1.GET("/rules/:language", handleGetLanguageRules)
		
		// Get supported languages
		v1.GET("/languages", handleGetSupportedLanguages)
		
		// Webhook notifications
		v1.POST("/webhook/notify", handleWebhookNotify)
	}
	
	return router
}

// Handler for code analysis endpoint
func handleAnalyzeCode(c *gin.Context) {
	var request service.AnalysisRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Add user ID if authenticated via API key
	requestApiKey := c.GetHeader("X-API-Key")
	if user, err := userRepository.GetUserByAPIKey(c.Request.Context(), requestApiKey); err == nil && user != nil {
		request.UserID = user.ID
	}
	
	// Analyze the code
	result, err := analysisService.AnalyzeCode(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Analysis failed: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// Handler for getting analysis by ID
func handleGetAnalysisById(c *gin.Context) {
	id := c.Param("id")
	
	// Get the analysis
	result, err := analysisService.GetAnalysisById(c.Request.Context(), id)
	if err != nil {
		// Check for specific errors
		if err.Error() == "analysis not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Analysis not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get analysis: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// Handler for getting language-specific rules
func handleGetLanguageRules(c *gin.Context) {
	language := c.Param("language")
	
	// Get linter for this language
	linter, ok := linterRegistry.GetLinter(language)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Language not supported",
		})
		return
	}
	
	// In a real implementation, we would get rules from the linter
	// For now, return sample data
	
	var rules []map[string]interface{}
	
	switch language {
	case "python":
		rules = []map[string]interface{}{
			{
				"id":          "missing-docstring",
				"name":        "Missing Docstring",
				"description": "Functions, classes, and modules should have docstrings",
				"severity":    "warning",
			},
			{
				"id":          "print-used",
				"name":        "Use Logging",
				"description": "Use logging module instead of print statements",
				"severity":    "suggestion",
			},
		}
	case "javascript", "typescript":
		rules = []map[string]interface{}{
			{
				"id":          "semi",
				"name":        "Missing Semicolon",
				"description": "Statements should end with a semicolon",
				"severity":    "error",
			},
			{
				"id":          "no-var",
				"name":        "Prefer Const/Let",
				"description": "Use const or let instead of var",
				"severity":    "warning",
			},
		}
	default:
		rules = []map[string]interface{}{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"language": language,
		"rules":    rules,
	})
}

// Handler for getting supported languages
func handleGetSupportedLanguages(c *gin.Context) {
	languages := linterRegistry.GetSupportedLanguages()
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"languages": languages,
	})
}

// Handler for webhook notifications
func handleWebhookNotify(c *gin.Context) {
	var request struct {
		AnalysisId string `json:"analysis_id" binding:"required"`
		Event      string `json:"event" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// TODO: Process webhook notification
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Received notification for analysis %s: %s", request.AnalysisId, request.Event),
	})
}