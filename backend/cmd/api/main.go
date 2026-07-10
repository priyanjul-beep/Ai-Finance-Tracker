// main.go – application entry point.
// Wires all dependencies, registers routes, and starts the HTTP server
// with graceful shutdown.
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

	"github.com/gin-gonic/gin"

	"github.com/priyanjul/ai-finance-tracker/config"
	"github.com/priyanjul/ai-finance-tracker/internal/handler"
	"github.com/priyanjul/ai-finance-tracker/internal/middleware"
	"github.com/priyanjul/ai-finance-tracker/internal/repository"
	"github.com/priyanjul/ai-finance-tracker/internal/usecase"
	pkgAI "github.com/priyanjul/ai-finance-tracker/pkg/ai"
	"github.com/priyanjul/ai-finance-tracker/pkg/cache"
	"github.com/priyanjul/ai-finance-tracker/pkg/database"
	"github.com/priyanjul/ai-finance-tracker/pkg/email"
	"github.com/priyanjul/ai-finance-tracker/pkg/logger"
	"github.com/priyanjul/ai-finance-tracker/pkg/monitoring"
	"github.com/priyanjul/ai-finance-tracker/pkg/queue"
	"github.com/priyanjul/ai-finance-tracker/pkg/storage"
)

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	appLog, err := logger.New(cfg.LogLevel, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer appLog.Sync()

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		appLog.Fatal("database connect failed", "error", err)
	}
	defer database.Close(db)

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisCache, err := cache.New(cfg.RedisURL)
	if err != nil {
		appLog.Fatal("redis connect failed", "error", err)
	}
	defer redisCache.Close()

	// ── Queue ─────────────────────────────────────────────────────────────────
	queueClient, err := queue.NewClient(cfg.RedisURL)
	if err != nil {
		appLog.Fatal("queue init failed", "error", err)
	}
	defer queueClient.Close()

	// ── AI Provider ───────────────────────────────────────────────────────────
	gemini, err := pkgAI.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		appLog.Fatal("gemini init failed", "error", err)
	}
	defer gemini.Close()

	aiFactory := pkgAI.NewFactory()
	aiFactory.Register("gemini", gemini)
	aiProvider := aiFactory.Default()

	// ── Email ─────────────────────────────────────────────────────────────────
	emailSvc := email.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	// ── Storage ───────────────────────────────────────────────────────────────
	storageSvc, err := storage.NewLocalStorage(cfg.LocalStoragePath, fmt.Sprintf("http://localhost:%s/uploads", cfg.ServerPort))
	if err != nil {
		appLog.Fatal("storage init failed", "error", err)
	}
	_ = storageSvc // referenced by future file-upload handlers

	// ── Metrics ───────────────────────────────────────────────────────────────
	metrics := monitoring.NewMetrics()

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo     := repository.NewUser(db)
	sessionRepo  := repository.NewSession(db)
	expenseRepo  := repository.NewExpense(db)
	incomeRepo   := repository.NewIncome(db)
	budgetRepo   := repository.NewBudget(db)
	subRepo      := repository.NewSubscription(db)
	goalRepo     := repository.NewGoal(db)
	tagRepo      := repository.NewTag(db)
	notifRepo    := repository.NewNotification(db)
	auditRepo    := repository.NewAuditLog(db)
	merchantRepo := repository.NewMerchantMapping(db)
	healthRepo   := repository.NewHealthScore(db)

	_ = merchantRepo

	// ── Use-cases (application services) ─────────────────────────────────────
	authUC := usecase.NewAuth(
		userRepo, sessionRepo, emailSvc,
		cfg.JWTSecret, cfg.JWTRefreshSecret,
		cfg.JWTExpiry, cfg.RefreshTokenExpiry,
		"http://localhost:"+cfg.ServerPort,
	)

	expenseUC := usecase.NewExpense(
		expenseRepo, tagRepo, auditRepo, aiProvider, redisCache, queueClient, merchantRepo,
	)

	analyticsUC := usecase.NewAnalytics(
		expenseRepo, incomeRepo, budgetRepo, healthRepo, subRepo, aiProvider, redisCache,
	)

	incomeUC := usecase.NewIncome(incomeRepo, auditRepo, redisCache)
	budgetUC := usecase.NewBudget(budgetRepo, expenseRepo, queueClient, auditRepo)
	subscriptionUC := usecase.NewSubscription(subRepo, auditRepo, redisCache)
	goalUC := usecase.NewGoal(goalRepo, auditRepo, redisCache)
	tagUC  := usecase.NewTag(tagRepo)
	userUC := usecase.NewUser(userRepo)

	// ── Handlers ──────────────────────────────────────────────────────────────
	authH      := handler.NewAuth(authUC)
	userH      := handler.NewUserHandler(userUC)
	expenseH   := handler.NewExpenseHandler(expenseUC, cfg.TesseractPath, cfg.EnableOCR)
	analyticsH := handler.NewAnalyticsHandler(analyticsUC)
	incomeH     := handler.NewIncomeHandler(incomeUC)
	budgetH     := handler.NewBudgetHandler(budgetUC)
	subH        := handler.NewSubscriptionHandler(subscriptionUC)
	goalH       := handler.NewGoalHandler(goalUC)
	tagH        := handler.NewTagHandler(tagUC)
	notifH      := handler.NewNotificationHandler(notifRepo)

	// ── Router ────────────────────────────────────────────────────────────────
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.SecurityHeaders(),
		monitoring.GinMiddleware(metrics),
	)

	// Static uploads
	r.Static("/uploads", cfg.LocalStoragePath)

	// ── Public routes ─────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Health / readiness
		v1.GET("/health", monitoring.HealthHandler(map[string]func() error{
			"database": func() error {
				sqlDB, err := db.DB()
				if err != nil {
					return err
				}
				return sqlDB.Ping()
			},
			"redis": func() error { return redisCache.Client().Ping(context.Background()).Err() },
		}))

		// Prometheus metrics
		if cfg.PrometheusEnabled {
			v1.GET("/metrics", gin.WrapH(monitoring.PrometheusHandler()))
		}

		auth := v1.Group("/auth")
		{
			auth.POST("/signup",          authH.Signup)
			auth.POST("/login",           authH.Login)
			auth.POST("/refresh",         authH.RefreshToken)
			auth.POST("/forgot-password", authH.ForgotPassword)
			auth.POST("/reset-password",  authH.ResetPassword)
			auth.GET("/verify-email/:token", authH.VerifyEmail)
			auth.GET("/google/callback",  authH.GoogleOAuthCallback)
		}
	}

	// ── Protected routes ──────────────────────────────────────────────────────
	protected := v1.Group("")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		// User profile
		protected.GET("/user/profile", userH.GetProfile)
		protected.PUT("/user/profile", userH.UpdateProfile)

		users := protected.Group("/users")
		{
			users.POST("/logout",          authH.Logout)
			users.POST("/change-password", authH.ChangePassword)
		}

		// Expenses
		expenses := protected.Group("/expenses")
		{
			expenses.POST("",              expenseH.Create)
			expenses.GET("",               expenseH.List)
			expenses.GET("/search",        expenseH.Search)
			expenses.GET("/:id",           expenseH.Get)
			expenses.PUT("/:id",           expenseH.Update)
			expenses.DELETE("/:id",        expenseH.Delete)
			expenses.POST("/parse",        expenseH.Parse)
			expenses.POST("/voice-parse",  expenseH.VoiceParse)
			expenses.POST("/scan",         expenseH.ScanReceipt)
			expenses.GET("/:id/duplicates", expenseH.GetDuplicates)
		}

		// Analytics & Dashboard
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/dashboard",              analyticsH.Dashboard)
			analytics.GET("/insights",               analyticsH.Insights)
			analytics.GET("/predictions",            analyticsH.Predictions)
			analytics.GET("/monthly/:month/:year",   analyticsH.MonthlyReport)
			analytics.GET("/yearly/:year",           analyticsH.YearlyReport)
			analytics.GET("/health-score",           analyticsH.HealthScore)
		}

		// Income
		income := protected.Group("/income")
		{
			income.POST("",     incomeH.Create)
			income.GET("",      incomeH.List)
			income.GET("/:id",  incomeH.Get)
			income.PUT("/:id",  incomeH.Update)
			income.DELETE("/:id", incomeH.Delete)
		}

		// Budgets
		budgets := protected.Group("/budgets")
		{
			budgets.POST("",     budgetH.Create)
			budgets.GET("",      budgetH.List)
			budgets.GET("/:id",  budgetH.Get)
			budgets.PUT("/:id",  budgetH.Update)
			budgets.DELETE("/:id", budgetH.Delete)
		}

		// Subscriptions
		subs := protected.Group("/subscriptions")
		{
			subs.POST("",              subH.Create)
			subs.GET("",               subH.List)
			subs.GET("/upcoming",      subH.GetUpcoming)
			subs.GET("/:id",           subH.GetByID)
			subs.PUT("/:id",           subH.Update)
			subs.DELETE("/:id",        subH.Delete)
		}

		// Goals
		goals := protected.Group("/goals")
		{
			goals.POST("",              goalH.Create)
			goals.GET("",               goalH.List)
			goals.GET("/:id",           goalH.GetByID)
			goals.PUT("/:id",           goalH.Update)
			goals.DELETE("/:id",        goalH.Delete)
			goals.POST("/:id/contribute", goalH.Contribute)
		}

		// Tags
		tags := protected.Group("/tags")
		{
			tags.POST("",      tagH.Create)
			tags.GET("",       tagH.List)
			tags.PUT("/:id",   tagH.Update)
			tags.DELETE("/:id", tagH.Delete)
		}
		// Expense–tag associations (use :id to match the expenses group wildcard)
		protected.POST("/expenses/:id/tags/:tag_id",   tagH.AddToExpense)
		protected.DELETE("/expenses/:id/tags/:tag_id", tagH.RemoveFromExpense)

		// Notifications
		notifs := protected.Group("/notifications")
		{
			notifs.GET("",                notifH.List)
			notifs.PATCH("/:id/read",     notifH.MarkRead)
			notifs.PATCH("/read-all",     notifH.MarkAllRead)
			notifs.DELETE("/:id",         notifH.Delete)
		}
	}

	// ── HTTP server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		appLog.Info("server starting", "port", cfg.ServerPort, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLog.Fatal("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	appLog.Info("shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLog.Error("graceful shutdown failed", "error", err)
	}
	appLog.Info("server stopped")
}
