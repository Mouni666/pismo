package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// envOr returns the value of an environment variable or a default
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// For Docker Compose use host "db", for local run use "localhost"
	dsn := envOr("DB_DSN", "postgres://pismo:pismo@localhost:5432/pismo?sslmode=disable")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto-migrate tables and seed operation types
	if err := db.AutoMigrate(&Account{}, &OperationType{}, &Transaction{}); err != nil {
		log.Fatal(err)
	}
	if err := seedOps(db); err != nil {
		log.Fatal(err)
	}

	s := &Server{DB: db}
	r := gin.Default()

	// --- Robust CORS config (lets Swagger 8081 call API 8080) ---
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // in production, whitelist specific URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	// -------------------------------------------------------------

	// API routes
	r.POST("/accounts", s.createAccount)
	r.GET("/accounts/:accountId", s.getAccount) // path param form
	r.GET("/accounts", s.getAccount)            // query param form
	r.POST("/transactions", s.createTransaction)

	// Serve the OpenAPI spec file
	r.StaticFile("/openapi.yaml", "./openapi.yaml")

	// Swagger UI at /docs
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `
<!DOCTYPE html>
<html>
  <head>
    <title>API Docs</title>
    <link rel="stylesheet" type="text/css"
          href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/openapi.yaml',
          dom_id: '#swagger-ui',
        });
      };
    </script>
  </body>
</html>
`)
	})

	// Start server
	port := envOr("PORT", "8080")
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
