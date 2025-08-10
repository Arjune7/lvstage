package clicks

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClickHandler struct {
	service Service
}

func NewHandler(service Service) *ClickHandler {
	return &ClickHandler{service: service}
}

func (h *ClickHandler) HandleClick(c *gin.Context) {
	var req ClickRequestData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.UserIP = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	// Optional: support client-sent EventID for idempotency
	if req.EventID == uuid.Nil {
		req.EventID = uuid.New()
	}

	if err := h.service.RecordClick(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process click"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":    "click queued",
		"event_id":  req.EventID.String(),
		"timestamp": req.Timestamp,
	})
}
