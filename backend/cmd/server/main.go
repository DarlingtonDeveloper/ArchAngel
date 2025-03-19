package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

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
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
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
		
		// Webhook notifications
		v1.POST("/webhook/notify", webhookNotify)
	}
	
	return router
}

// Handler for code analysis endpoint
func analyzeCode(c *gin.Context) {
	var request struct {
		Code     string `json:"code" binding:"required"`
		Language string `json:"language" binding:"required"`
		Context  string `json:"context"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request: " + err.Error(),
		})
		return
	}
	
	// Generate a unique ID for this analysis
	analysisID := fmt.Sprintf("analysis-%d", time.Now().UnixNano())
	
	// Process the code based on the language
	var issues []map[string]interface{}
	switch request.Language {
	case "python":
		issues = analyzePythonCode(request.Code)
	case "javascript":
	case "typescript":
		issues = analyzeJavaScriptCode(request.Code)
	case "go":
		issues = analyzeGoCode(request.Code)
	case "java":
		issues = analyzeJavaCode(request.Code)
	case "csharp":
		issues = analyzeCSharpCode(request.Code)
	case "php":
		issues = analyzePHPCode(request.Code)
	case "ruby":
		issues = analyzeRubyCode(request.Code)
	default:
		issues = analyzeGenericCode(request.Code)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"id":     analysisID,
		"language": request.Language,
		"context": request.Context,
		"timestamp": time.Now().Format(time.RFC3339),
		"issues": issues,
	})
}

// Handler for getting analysis by ID
func getAnalysisById(c *gin.Context) {
	id := c.Param("id")
	
	// In a real implementation, we would retrieve the analysis from a database
	// For now, return a mock response
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"id":     id,
		"language": "python",
		"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		"issues": []map[string]interface{}{
			{
				"line":     1,
				"message":  "Missing docstring",
				"severity": "warning",
				"ruleId":   "PY001",
			},
			{
				"line":     1,
				"message":  "Use logging instead of print",
				"severity": "suggestion",
				"ruleId":   "PY002",
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
				"id":          "PY001",
				"name":        "Missing Docstring",
				"description": "Functions, classes, and modules should have docstrings",
				"severity":    "warning",
			},
			{
				"id":          "PY002",
				"name":        "Use Logging",
				"description": "Use logging module instead of print statements",
				"severity":    "suggestion",
			},
		}
	case "javascript":
	case "typescript":
		rules = []map[string]interface{}{
			{
				"id":          "JS001",
				"name":        "Missing Semicolon",
				"description": "Statements should end with a semicolon",
				"severity":    "error",
			},
			{
				"id":          "JS002",
				"name":        "Prefer Const",
				"description": "Use const for variables that are never reassigned",
				"severity":    "warning",
			},
		}
	default:
		rules = []map[string]interface{}{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"language": language,
		"rules":   rules,
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

// Language-specific analysis functions

func analyzePythonCode(code string) []map[string]interface{} {
	// This would be replaced with actual Python linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Missing docstring",
			"severity": "warning",
			"ruleId":   "PY001",
		},
		{
			"line":     1,
			"message":  "Use logging instead of print",
			"severity": "suggestion",
			"ruleId":   "PY002",
		},
	}
}

func analyzeJavaScriptCode(code string) []map[string]interface{} {
	// This would be replaced with actual JavaScript linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Missing semicolon",
			"severity": "error",
			"ruleId":   "JS001",
		},
		{
			"line":     1,
			"message":  "Prefer const over let",
			"severity": "warning",
			"ruleId":   "JS002",
		},
	}
}

func analyzeGoCode(code string) []map[string]interface{} {
	// This would be replaced with actual Go linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Exported function missing comment",
			"severity": "warning",
			"ruleId":   "GO001",
		},
		{
			"line":     1,
			"message":  "Error not handled",
			"severity": "error",
			"ruleId":   "GO002",
		},
	}
}

func analyzeJavaCode(code string) []map[string]interface{} {
	// This would be replaced with actual Java linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Class should be in its own file",
			"severity": "warning",
			"ruleId":   "JV001",
		},
		{
			"line":     1,
			"message":  "Missing Javadoc comment",
			"severity": "warning",
			"ruleId":   "JV002",
		},
	}
}

func analyzeCSharpCode(code string) []map[string]interface{} {
	// This would be replaced with actual C# linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Use var instead of explicit type",
			"severity": "suggestion",
			"ruleId":   "CS001",
		},
		{
			"line":     1,
			"message":  "Class name should be PascalCase",
			"severity": "warning",
			"ruleId":   "CS002",
		},
	}
}

func analyzePHPCode(code string) []map[string]interface{} {
	// This would be replaced with actual PHP linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Use strict comparison (===)",
			"severity": "warning",
			"ruleId":   "PHP001",
		},
		{
			"line":     1,
			"message":  "Function name should be camelCase",
			"severity": "warning",
			"ruleId":   "PHP002",
		},
	}
}

func analyzeRubyCode(code string) []map[string]interface{} {
	// This would be replaced with actual Ruby linting/analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Use snake_case for method names",
			"severity": "warning",
			"ruleId":   "RB001",
		},
		{
			"line":     1,
			"message":  "Prefer single quotes for strings",
			"severity": "suggestion",
			"ruleId":   "RB002",
		},
	}
}

func analyzeGenericCode(code string) []map[string]interface{} {
	// This would be replaced with generic code analysis
	return []map[string]interface{}{
		{
			"line":     1,
			"message":  "Line too long (>100 characters)",
			"severity": "warning",
			"ruleId":   "GEN001",
		},
	}
}