package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Server holds shared dependencies.
type Server struct {
	DB *gorm.DB
}

/* ===========================
   Accounts
   ===========================*/

// POST /accounts
func (s *Server) createAccount(c *gin.Context) {
	var in struct {
		DocumentNumber string `json:"document_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || in.DocumentNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document_number is required"})
		return
	}

	// If an account with the same document exists, return it.
	var existing Account
	if err := s.DB.Where("document_number = ?", in.DocumentNumber).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"account_id":      existing.ID,
			"document_number": existing.DocumentNumber,
		})
		return
	}

	acc := Account{DocumentNumber: in.DocumentNumber}
	if err := s.DB.Create(&acc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"account_id":      acc.ID,
		"document_number": acc.DocumentNumber,
	})
}

// GET /accounts and GET /accounts/:accountId
func (s *Server) getAccount(c *gin.Context) {
	idStr := c.Param("accountId")
	if idStr == "" {
		idStr = c.Query("accountId")
	}
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accountId is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid accountId"})
		return
	}

	var acc Account
	if err := s.DB.First(&acc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id":      acc.ID,
		"document_number": acc.DocumentNumber,
	})
}

/* ===========================
   Transactions
   ===========================*/

// POST /transactions
func (s *Server) createTransaction(c *gin.Context) {
	var in struct {
		AccountID       uint    `json:"account_id" binding:"required"`
		OperationTypeID int     `json:"operation_type_id" binding:"required"`
		Amount          float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id, operation_type_id and amount are required"})
		return
	}

	// Validate account exists
	var acc Account
	if err := s.DB.First(&acc, in.AccountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "account not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	// Validate operation type exists
	var op OperationType
	if err := s.DB.First(&op, in.OperationTypeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "operation type not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	// Enforce sign: 1/2/3 -> negative; 4 -> positive
	amount := in.Amount
	switch in.OperationTypeID {
	case 1, 2, 3:
		if amount > 0 {
			amount = -amount
		}
	case 4:
		if amount < 0 {
			amount = -amount
		}
	}

	t := Transaction{
		AccountID:       in.AccountID,
		OperationTypeID: in.OperationTypeID, // int in your model
		Amount:          amount,
		EventDate:       time.Now().UTC(),
	}

	if err := s.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transaction_id":   t.ID,
		"account_id":       t.AccountID,
		"operation_type_id": t.OperationTypeID,
		"amount":           t.Amount,
		"event_date":       t.EventDate,
	})
}
