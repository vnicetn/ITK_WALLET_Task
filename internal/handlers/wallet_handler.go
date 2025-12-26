package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
	"github.com/itk/wallet/internal/service"
)

type WalletHandler struct {
	service *service.WalletService
}

func NewWalletHandler(service *service.WalletService) *WalletHandler {
	return &WalletHandler{
		service: service,
	}
}

type OperationRequest struct {
	WalletID      uuid.UUID            `json:"valletId" binding:"required"`
	OperationType models.OperationType `json:"operationType" binding:"required"`
	Amount        int                  `json:"amount" binding:"required"`
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	walletIDStr := c.Param("WALLET_UUID")

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid wallet ID"})
		return
	}

	balance, err := h.service.GetBalance(c.Request.Context(), walletID)
	if err != nil {
		if strings.Contains(err.Error(), "balance not found") || strings.Contains(err.Error(), "not found") {
			c.AbortWithStatusJSON(404, gin.H{"error": "wallet not found"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(200, gin.H{"balance": balance})
}

func (h *WalletHandler) ProcessOperation(c *gin.Context) {
	var req OperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.AbortWithStatusJSON(400, gin.H{"error": "amount must be greater than zero"})
		return
	}

	_, err := h.service.UpdateBalance(c.Request.Context(), req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		if strings.Contains(err.Error(), "wallet not found") || strings.Contains(err.Error(), "balance not found") {
			_, err := h.service.CreateWallet(c.Request.Context(), req.WalletID)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "failed to create wallet"})
				return
			}

			_, err = h.service.UpdateBalance(c.Request.Context(), req.WalletID, req.OperationType, req.Amount)
			if err != nil {
				if strings.Contains(err.Error(), "insufficient funds") {
					c.AbortWithStatusJSON(400, gin.H{"error": "insufficient funds"})
					return
				}
				if strings.Contains(err.Error(), "invalid operationType") || strings.Contains(err.Error(), "amount cannot be zero") {
					c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
					return
				}
				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
				return
			}
			c.JSON(200, gin.H{"message": "operation completed"})
			return
		}

		if strings.Contains(err.Error(), "insufficient funds") {
			c.AbortWithStatusJSON(400, gin.H{"error": "insufficient funds"})
			return
		}

		if strings.Contains(err.Error(), "invalid operationType") || strings.Contains(err.Error(), "amount cannot be zero") {
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}

		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(200, gin.H{"message": "operation completed"})
}
