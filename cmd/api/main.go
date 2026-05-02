package main

import (
	"log"
	"net/http"
	"todo_api/internal/config"
	"todo_api/internal/databse"
	"todo_api/internal/handlers"
	"todo_api/internal/middleware"
	"todo_api/internal/models"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	var pool *pgxpool.Pool
	pool, err = databse.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// Cleanup expired refresh tokens on server startup. In production, consider
	// running this as a periodic background job instead.
	if err := repository.DeleteExpiredTokens(pool); err != nil {
		log.Printf("Warning: failed to delete expired refresh tokens: %v", err)
	}
	

	router := gin.Default()

	// ── Health check ──────────────────────────────────────────────────────────
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message":  "server is running...",
			"status":   "success",
			"database": "connected",
		})
	})

	// ── Auth routes (public) ──────────────────────────────────────────────────
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.CreateUserHandler(pool))
		auth.POST("/login", handlers.LoginHandler(pool, cfg))
		auth.POST("/refresh", handlers.RefreshHandler(pool, cfg))

		// Logout can be called without an access token (token may already be
		// expired), so AuthMiddleware is optional here. We include it so that
		// all_devices logout can resolve the user_id from the JWT when present.
		auth.POST("/logout", middleware.AuthMiddleware(cfg), handlers.LogoutHandler(pool, cfg))
	}

	// ── Protected: any authenticated user ────────────────────────────────────
	protected := router.Group("/todos")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.POST("/", handlers.CreateTodoHandler(pool))
		protected.GET("/", handlers.GetAllTodosHandler(pool))
		protected.GET("/:id", handlers.GetTodoByIdHandler(pool))
		protected.PUT("/:id", handlers.UpdateTodoHandler(pool))
		protected.DELETE("/:id", handlers.DeleteTodoHandler(pool))
	}

	// ── Admin-only routes ─────────────────────────────────────────────────────
	// Example: only admins can list all users or delete any todo.
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg), middleware.RequireRole(models.RoleAdmin))
	{
		// Placeholder — add your admin handlers here.
		admin.GET("/users", handlers.GetAllUsersHandler(pool))
		// admin.DELETE("/todos/:id", handlers.AdminDeleteTodoHandler(pool))
		admin.GET("/todos", handlers.AdminGetAllTodosHandler(pool))
	}

	// ── Middleware test ───────────────────────────────────────────────────────
	router.GET(
		"/protected-test",
		middleware.AuthMiddleware(cfg),
		handlers.TestProtectedHandler(),
	)

	log.Printf("Server starting on :%s", cfg.Port)
	router.Run(":" + cfg.Port)
}
