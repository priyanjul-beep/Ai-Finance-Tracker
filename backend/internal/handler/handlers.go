// Package handler – expense, income, budget, subscription, goal, tag, analytics,
// notification handlers.
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
	"github.com/priyanjul/ai-finance-tracker/internal/middleware"
)

// ─── Expense ──────────────────────────────────────────────────────────────────

// ExpenseHandler handles /expenses endpoints.
type ExpenseHandler struct{ svc interfaces.ExpenseService }

func NewExpenseHandler(svc interfaces.ExpenseService) *ExpenseHandler { return &ExpenseHandler{svc: svc} }

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *ExpenseHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	p := parsePagination(c)
	result, err := h.svc.List(c.Request.Context(), userID, p)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.svc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ExpenseHandler) Parse(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.AIExpenseParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.ParseFromText(c.Request.Context(), userID, req)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) Search(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	query := c.Query("q")
	if query == "" {
		abortBadRequest(c, "query parameter 'q' is required")
		return
	}
	result, err := h.svc.Search(c.Request.Context(), userID, query)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ExpenseHandler) GetDuplicates(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetDuplicates(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ─── Income ───────────────────────────────────────────────────────────────────

// IncomeHandler handles /incomes endpoints.
type IncomeHandler struct{ svc interfaces.IncomeService }

func NewIncomeHandler(svc interfaces.IncomeService) *IncomeHandler { return &IncomeHandler{svc: svc} }

func (h *IncomeHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *IncomeHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *IncomeHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	p := parsePagination(c)
	var from, to time.Time
	if f := c.Query("from"); f != "" {
		from, _ = time.Parse(time.RFC3339, f)
	}
	if t := c.Query("to"); t != "" {
		to, _ = time.Parse(time.RFC3339, t)
	}
	results, total, err := h.svc.List(c.Request.Context(), userID, from, to, p.Page, p.Limit)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": results, "total": total, "page": p.Page, "limit": p.Limit})
}

func (h *IncomeHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *IncomeHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.svc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Budget ───────────────────────────────────────────────────────────────────

// BudgetHandler handles /budgets endpoints.
type BudgetHandler struct{ svc interfaces.BudgetService }

func NewBudgetHandler(svc interfaces.BudgetService) *BudgetHandler { return &BudgetHandler{svc: svc} }

func (h *BudgetHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *BudgetHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *BudgetHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	result, err := h.svc.List(c.Request.Context(), userID, year, month)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"budgets": result, "total": len(result)})
}

func (h *BudgetHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.svc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.svc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Analytics ────────────────────────────────────────────────────────────────

// AnalyticsHandler handles /analytics endpoints.
type AnalyticsHandler struct{ svc interfaces.AnalyticsService }

func NewAnalyticsHandler(svc interfaces.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) MonthlyReport(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	month, _ := strconv.Atoi(c.Param("month"))
	year, _ := strconv.Atoi(c.Param("year"))
	result, err := h.svc.GetMonthlyReport(c.Request.Context(), userID, month, year)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) YearlyReport(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	year, _ := strconv.Atoi(c.Param("year"))
	result, err := h.svc.GetYearlyReport(c.Request.Context(), userID, year)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) Predictions(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetPredictions(c.Request.Context(), userID)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) Insights(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.svc.GetInsights(c.Request.Context(), userID)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"insights": result})
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

func abortBadRequest(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Code: "BAD_REQUEST", Message: msg})
}

func abortNotFound(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: msg})
}

func abortServerError(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Code: "INTERNAL_ERROR", Message: msg})
}

func parsePagination(c *gin.Context) dto.PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return dto.PaginationParams{
		Page:  page,
		Limit: limit,
		Sort:  c.Query("sort"),
		Order: c.DefaultQuery("order", "desc"),
	}
}
