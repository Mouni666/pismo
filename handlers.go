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

	// If an account with the same document exists, return 409 (tests expect this).
	var existing Account
	if err := s.DB.Where("document_number = ?", in.DocumentNumber).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "account already exists"})
		return
	}

	acc := Account{DocumentNumber: in.DocumentNumber}
	if err := s.DB.Create(&acc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	// Return the full struct (tests unmarshal into Account).
	c.JSON(http.StatusCreated, acc)
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

	// Return the full struct.
	c.JSON(http.StatusOK, acc)
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

	// Normalize amount by operation type
	amt := normalizeAmount(in.OperationTypeID, in.Amount)

	t := Transaction{
		AccountID:       in.AccountID,
		OperationTypeID: in.OperationTypeID,
		Amount:          amt,
		EventDate:       time.Now().UTC(),
	}

	if err := s.DB.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	// Return the full struct (tests unmarshal into Transaction).
	c.JSON(http.StatusCreated, t)
}
