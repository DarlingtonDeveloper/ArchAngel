package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/codehawk/backend/pkg/analyzer"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Global linter registry
var linterRegistry *analyzer.LinterRegistry

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize the linter registry
	initLinterRegistry()

	// Set up the server
	router := setupRouter()
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting CodeHawk API server on port %s\n", port)
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

// Initialize linter registry with supported linters
func initLinterRegistry() {
	linterRegistry = analyzer.NewLinterRegistry()
	
	// Register Python linter
	pythonLinter := analyzer.NewPythonLinter(map[string]string{
		"pylintPath": os.Getenv("PYLINT_PATH"),
		"timeout":    "15s",
	})
	linterRegistry.Register(pythonLinter)
	
	// Register JavaScript linter
	jsLinter := analyzer.NewJavaScriptLinter(map[string]string{
		"eslintPath": os.Getenv("ESLINT_PATH"),
		"configPath": os.Getenv("ESLINT_CONFIG"),
		"timeout":    "15s",
	})
	linterRegistry.Register(jsLinter)
	
	// Register TypeScript linter
	tsLinter := analyzer.NewTypeScriptLinter(map[string]string{
		"eslintPath": os.Getenv("ESLINT_PATH"),
		"configPath": os.Getenv("TSLINT_CONFIG"),
		"timeout":    "15s",
	})
	linterRegistry.Register(tsLinter)
	
	log.Printf("Initialized linter registry with %d linters\n", len(linterRegistry.GetSupportedLanguages()))
}

func setupRouter() *gin.Engine {
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
		
		// Get API key from environment (for development)
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			apiKey = "demo_api_key_123" // Default development API key
		}
		
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid API key",
			})
			c.Abort()
			return
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
		})
	})
	
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Code analysis endpoint
		v1.POST("/analyze", analyzeCode)
		
		// Get analysis results by ID
		v1.GET("/analysis/:id", getAnalysisById)
		
		// Get language-specific rules
		v1.GET("/rules/:language", getLanguageRules)
		
		// Get supported languages
		v1.GET("/languages", getSupportedLanguages)
		
		// Webhook notifications
		v1.POST("/webhook/notify", webhookNotify)
	}
	
	return router
}

// Handler for code analysis endpoint
func analyzeCode(c *gin.Context) {
	var request struct {
		Code     string                 `json:"code" binding:"required"`
		Language string                 `json:"language" binding:"required"`
		Context  string                 `json:"context"`
		Options  map[string]interface{} `json:"options"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Check if we have a linter for this language
	linter, ok := linterRegistry.GetLinter(request.Language)
	if !ok {
		// Fall back to generic analysis
		result := analyzeGenericCode(request.Code)
		
		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"id":        fmt.Sprintf("analysis-%d", time.Now().UnixNano()),
			"language":  request.Language,
			"context":   request.Context,
			"timestamp": time.Now().Format(time.RFC3339),
			"issues":    result,
		})
		return
	}
	
	// Analyze the code using the appropriate linter
	ctx := context.Background()
	result, err := linter.Analyze(ctx, request.Code, request.Options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Analysis failed: " + err.Error(),
		})
		return
	}
	
	// Generate suggestions if not already included
	if len(result.Suggestions) == 0 {
		suggestions, err := linter.SuggestFixes(ctx, request.Code, result.Issues)
		if err == nil && len(suggestions) > 0 {
			result.Suggestions = suggestions
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"id":          fmt.Sprintf("analysis-%d", time.Now().UnixNano()),
		"language":    request.Language,
		"context":     request.Context,
		"timestamp":   time.Now().Format(time.RFC3339),
		"issues":      result.Issues,
		"suggestions": result.Suggestions,
	})
}

// Handler for getting analysis by ID
func getAnalysisById(c *gin.Context) {
	id := c.Param("id")
	
	// In a real implementation, we would retrieve the analysis from a database
	// For now, return a mock response
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"id":        id,
		"language":  "python",
		"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		"issues": []map[string]interface{}{
			{
				"line":     1,
				"message":  "Missing docstring",
				"severity": "warning",
				"ruleId":   "missing-docstring",
			},
			{
				"line":     1,
				"message":  "Use logging instead of print",
				"severity": "suggestion",
				"ruleId":   "print-used",
			},
		},
	})
}

// Handler for getting language-specific rules
func getLanguageRules(c *gin.Context) {
	language := c.Param("language")
	
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
	case "javascript":
	case "typescript":
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
func getSupportedLanguages(c *gin.Context) {
	languages := linterRegistry.GetSupportedLanguages()
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"languages": languages,
	})
}

// Handler for webhook notifications
func webhookNotify(c *gin.Context) {
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

// Utility functions for analyzing code when no specific linter is available

func analyzeGenericCode(code string) []map[string]interface{} {
	// This would be replaced with generic code analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Line too long (>100 characters)",
			"severity": "warning",
			"ruleId":   "line-length",
		},
	}
}