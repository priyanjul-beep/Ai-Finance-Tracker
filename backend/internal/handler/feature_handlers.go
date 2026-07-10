// Package handler – subscription, goal, tag, and notification HTTP handlers.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/middleware"
	"github.com/priyanjul/ai-finance-tracker/internal/usecase"
)

// ─────────────────────────────────────────────────────────────────────────────
// SubscriptionHandler
// ─────────────────────────────────────────────────────────────────────────────

// SubscriptionHandler handles /subscriptions routes.
type SubscriptionHandler struct {
	uc *usecase.SubscriptionUseCase
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(uc *usecase.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{uc: uc}
}

// Create godoc
// @Summary      Add a subscription
// @Tags         subscriptions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateSubscriptionRequest true "Subscription"
// @Success      201  {object} dto.SubscriptionDTO
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GetByID godoc
// @Summary      Get a subscription by ID
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Subscription ID"
// @Success      200 {object} dto.SubscriptionDTO
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.uc.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, "subscription not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update godoc
// @Summary      Update a subscription
// @Tags         subscriptions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Subscription ID"
// @Param        body body dto.UpdateSubscriptionRequest true "Updates"
// @Success      200  {object} dto.SubscriptionDTO
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortNotFound(c, "subscription not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
// @Summary      Delete a subscription
// @Tags         subscriptions
// @Security     BearerAuth
// @Param        id path string true "Subscription ID"
// @Success      204
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, "subscription not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary      List subscriptions
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Param        active query bool false "Active only"
// @Success      200 {array} dto.SubscriptionDTO
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	activeOnly := c.Query("active") == "true"
	results, err := h.uc.List(c.Request.Context(), userID, activeOnly)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": results, "total": len(results)})
}

// GetUpcoming godoc
// @Summary      Get upcoming subscription renewals
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Param        days query int false "Days ahead (default 7)"
// @Success      200 {array} dto.SubscriptionDTO
// @Router       /subscriptions/upcoming [get]
func (h *SubscriptionHandler) GetUpcoming(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	results, err := h.uc.GetUpcoming(c.Request.Context(), userID, days)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": results, "total": len(results)})
}

// ─────────────────────────────────────────────────────────────────────────────
// GoalHandler
// ─────────────────────────────────────────────────────────────────────────────

// GoalHandler handles /goals routes.
type GoalHandler struct {
	uc *usecase.GoalUseCase
}

// NewGoalHandler creates a new GoalHandler.
func NewGoalHandler(uc *usecase.GoalUseCase) *GoalHandler {
	return &GoalHandler{uc: uc}
}

// Create godoc
// @Summary      Create a financial goal
// @Tags         goals
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateGoalRequest true "Goal"
// @Success      201 {object} dto.GoalDTO
// @Router       /goals [post]
func (h *GoalHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GetByID godoc
// @Summary      Get a goal by ID
// @Tags         goals
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Goal ID"
// @Success      200 {object} dto.GoalDTO
// @Router       /goals/{id} [get]
func (h *GoalHandler) GetByID(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	result, err := h.uc.GetByID(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		abortNotFound(c, "goal not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update godoc
// @Summary      Update a goal
// @Tags         goals
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Goal ID"
// @Param        body body dto.UpdateGoalRequest true "Updates"
// @Success      200 {object} dto.GoalDTO
// @Router       /goals/{id} [put]
func (h *GoalHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortNotFound(c, "goal not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Contribute godoc
// @Summary      Add a contribution to a goal
// @Tags         goals
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path  string true "Goal ID"
// @Param        body body  object true "Contribution"
// @Success      200 {object} dto.GoalDTO
// @Router       /goals/{id}/contribute [post]
func (h *GoalHandler) Contribute(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var body struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Contribute(c.Request.Context(), userID, c.Param("id"), body.Amount)
	if err != nil {
		abortNotFound(c, "goal not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
// @Summary      Delete a goal
// @Tags         goals
// @Security     BearerAuth
// @Param        id path string true "Goal ID"
// @Success      204
// @Router       /goals/{id} [delete]
func (h *GoalHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, "goal not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary      List goals
// @Tags         goals
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "Filter by status (active|completed|paused)"
// @Success      200 {array} dto.GoalDTO
// @Router       /goals [get]
func (h *GoalHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	results, err := h.uc.List(c.Request.Context(), userID, c.Query("status"))
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"goals": results, "total": len(results)})
}

// ─────────────────────────────────────────────────────────────────────────────
// TagHandler
// ─────────────────────────────────────────────────────────────────────────────

// TagHandler handles /tags routes.
type TagHandler struct {
	uc *usecase.TagUseCase
}

// NewTagHandler creates a new TagHandler.
func NewTagHandler(uc *usecase.TagUseCase) *TagHandler {
	return &TagHandler{uc: uc}
}

// Create godoc
// @Summary      Create a tag
// @Tags         tags
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateTagRequest true "Tag"
// @Success      201 {object} dto.TagDTO
// @Router       /tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Create(c.Request.Context(), userID, req)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Update godoc
// @Summary      Update a tag
// @Tags         tags
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string true "Tag ID"
// @Param        body body dto.UpdateTagRequest true "Updates"
// @Success      200 {object} dto.TagDTO
// @Router       /tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortBadRequest(c, err.Error())
		return
	}
	result, err := h.uc.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		abortNotFound(c, "tag not found")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
// @Summary      Delete a tag
// @Tags         tags
// @Security     BearerAuth
// @Param        id path string true "Tag ID"
// @Success      204
// @Router       /tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		abortNotFound(c, "tag not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary      List all tags for current user
// @Tags         tags
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.TagDTO
// @Router       /tags [get]
func (h *TagHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	results, err := h.uc.List(c.Request.Context(), userID)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": results, "total": len(results)})
}

// AddToExpense godoc
// @Summary      Attach a tag to an expense
// @Tags         tags
// @Security     BearerAuth
// @Param        expense_id path string true "Expense ID"
// @Param        tag_id     path string true "Tag ID"
// @Success      204
// @Router       /expenses/{expense_id}/tags/{tag_id} [post]
func (h *TagHandler) AddToExpense(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.AddToExpense(c.Request.Context(), userID, c.Param("id"), c.Param("tag_id")); err != nil {
		abortNotFound(c, "tag or expense not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveFromExpense godoc
// @Summary      Detach a tag from an expense
// @Tags         tags
// @Security     BearerAuth
// @Param        expense_id path string true "Expense ID"
// @Param        tag_id     path string true "Tag ID"
// @Success      204
// @Router       /expenses/{expense_id}/tags/{tag_id} [delete]
func (h *TagHandler) RemoveFromExpense(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.RemoveFromExpense(c.Request.Context(), userID, c.Param("id"), c.Param("tag_id")); err != nil {
		abortNotFound(c, "tag or expense not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// ─────────────────────────────────────────────────────────────────────────────
// NotificationHandler
// ─────────────────────────────────────────────────────────────────────────────

// NotificationHandler handles /notifications routes.
type NotificationHandler struct {
	uc *usecase.NotificationUseCase
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(uc *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

// List returns a paginated list of notifications for the current user.
func (h *NotificationHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	p := parsePagination(c)
	notifType := c.Query("type")

	result, err := h.uc.List(c.Request.Context(), userID, notifType, p.Page, p.Limit)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// UnreadCount returns the number of unread notifications (cached).
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	count, err := h.uc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, dto.UnreadCountResponse{Count: count})
}

// MarkRead marks a single notification as read.
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.MarkAsRead(c.Request.Context(), c.Param("id"), userID); err != nil {
		abortNotFound(c, "notification not found")
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkAllRead marks all notifications as read for the current user.
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		abortServerError(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete removes a notification.
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	if err := h.uc.Delete(c.Request.Context(), c.Param("id"), userID); err != nil {
		abortNotFound(c, "notification not found")
		return
	}
	c.Status(http.StatusNoContent)
}

