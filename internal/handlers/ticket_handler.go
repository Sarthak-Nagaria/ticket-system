package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Sarthak-Nagaria/ticket-system/internal/middleware"
	"github.com/Sarthak-Nagaria/ticket-system/internal/models"
	"github.com/Sarthak-Nagaria/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TicketHandler holds dependencies required by ticket endpoints.
type TicketHandler struct {
	DB *gorm.DB
}

// NewTicketHandler constructs a TicketHandler.
func NewTicketHandler(db *gorm.DB) *TicketHandler {
	return &TicketHandler{DB: db}
}

type createTicketRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// currentUserID extracts the authenticated user's ID set by AuthRequired.
func currentUserID(c *gin.Context) uint {
	val, exists := c.Get(middleware.ContextUserID)
	if !exists {
		return 0
	}

	id, ok := val.(uint)
	if !ok {
		return 0
	}

	return id
}

// CreateTicket handles POST /tickets.
//
// Sample request:
//
//	{"title": "Server down", "description": "Production API returning 500s"}
//
// Sample success response (201):
//
//	{
//	  "id": 1, "title": "Server down", "description": "Production API returning 500s",
//	  "status": "open", "user_id": 1,
//	  "created_at": "2026-06-25T10:00:00Z", "updated_at": "2026-06-25T10:00:00Z"
//	}
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)

	if title == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "title cannot be empty")
		return
	}

	ticket := models.Ticket{
		Title:       title,
		Description: description,
		Status:      models.StatusOpen,
		UserID:      currentUserID(c),
	}

	if err := h.DB.Create(&ticket).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":          ticket.ID,
		"title":       ticket.Title,
		"description": ticket.Description,
		"status":      ticket.Status,
		"user_id":     ticket.UserID,
		"created_at":  ticket.CreatedAt,
		"updated_at":  ticket.UpdatedAt,
	})
}

// ListTickets handles GET /tickets. Returns only tickets owned by the
// authenticated user.
//
// Sample success response (200):
//
//	{"tickets": [{"id": 1, "title": "Server down", "status": "open", ...}]}
func (h *TicketHandler) ListTickets(c *gin.Context) {
	var tickets []models.Ticket
	if err := h.DB.Where("user_id = ?", currentUserID(c)).Order("created_at desc").Find(&tickets).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"tickets": tickets})
}

// fetchOwnedTicket loads a ticket by ID and verifies it belongs to the
// requesting user. It writes the appropriate error response itself
// (404 if not found at all, 403 if it belongs to someone else) and
// returns ok=false in those cases.
func (h *TicketHandler) fetchOwnedTicket(c *gin.Context) (*models.Ticket, bool) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid ticket id")
		return nil, false
	}

	var ticket models.Ticket
	if err := h.DB.First(&ticket, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "ticket not found")
		} else {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch ticket")
		}
		return nil, false
	}

	if ticket.UserID != currentUserID(c) {
		utils.ErrorResponse(c, http.StatusForbidden, "you do not have access to this ticket")
		return nil, false
	}

	return &ticket, true
}

// GetTicket handles GET /tickets/:id.
//
// Sample success response (200):
//
//	{"id": 1, "title": "Server down", "description": "...", "status": "open", "user_id": 1, ...}
//
// Sample error response (404): {"error": "ticket not found"}
// Sample error response (403): {"error": "you do not have access to this ticket"}
func (h *TicketHandler) GetTicket(c *gin.Context) {
	ticket, ok := h.fetchOwnedTicket(c)
	if !ok {
		return
	}
	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":          ticket.ID,
		"title":       ticket.Title,
		"description": ticket.Description,
		"status":      ticket.Status,
		"user_id":     ticket.UserID,
		"created_at":  ticket.CreatedAt,
		"updated_at":  ticket.UpdatedAt,
	})
}

// UpdateStatus handles PATCH /tickets/:id/status.
//
// Sample request:
//
//	{"status": "in_progress"}
//
// Sample success response (200):
//
//	{"id": 1, "status": "in_progress", "updated_at": "2026-06-25T10:05:00Z"}
//
// Sample error response (400) - invalid transition:
//
//	{"error": "cannot transition ticket from 'closed' to 'open'"}
func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	ticket, ok := h.fetchOwnedTicket(c)
	if !ok {
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request payload: "+err.Error())
		return
	}

	status := strings.TrimSpace(req.Status)

	if !models.IsValidStatus(status) {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid status: must be one of open, in_progress, closed")
		return
	}

	newStatus := models.TicketStatus(status)

	if ticket.Status == models.StatusClosed {
		utils.ErrorResponse(c, http.StatusBadRequest, "closed ticket cannot be reopened or modified")
		return
	}

	if !models.CanTransition(ticket.Status, newStatus) {
		utils.ErrorResponse(c, http.StatusBadRequest,
			"cannot transition ticket from '"+string(ticket.Status)+"' to '"+string(newStatus)+"'")
		return
	}

	ticket.Status = newStatus
	if err := h.DB.Save(ticket).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update ticket status")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":         ticket.ID,
		"status":     ticket.Status,
		"updated_at": ticket.UpdatedAt,
	})
}
