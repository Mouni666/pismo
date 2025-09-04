package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTest(t *testing.T) (*Server, *gin.Engine) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Account{}, &OperationType{}, &Transaction{}))
	require.NoError(t, seedOps(db))

	s := &Server{DB: db}
	r := gin.Default()
	r.POST("/accounts", s.createAccount)
	r.GET("/accounts/:accountId", s.getAccount)
	r.GET("/accounts", s.getAccount)
	r.POST("/transactions", s.createTransaction)
	return s, r
}

func TestNormalizeAmount(t *testing.T) {
	require.Equal(t, -10.0, normalizeAmount(1, 10))
	require.Equal(t, -10.0, normalizeAmount(2, 10))
	require.Equal(t, -10.0, normalizeAmount(3, 10))
	require.Equal(t, 10.0, normalizeAmount(4, 10))
	require.Equal(t, 10.0, normalizeAmount(4, -10))
}

func TestCreateAndGetAccount(t *testing.T) {
	_, r := setupTest(t)

	// create
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader([]byte(`{"document_number":"12345678900"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// duplicate -> 409
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader([]byte(`{"document_number":"12345678900"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)

	// get
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var acc Account
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &acc))
	require.Equal(t, uint(1), acc.ID)
	require.Equal(t, "12345678900", acc.DocumentNumber)
}

func TestCreateTransaction(t *testing.T) {
	_, r := setupTest(t)
	// need an account
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader([]byte(`{"document_number":"111"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// PAYMENT -> positive
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/transactions",
		bytes.NewReader([]byte(`{"account_id":1,"operation_type_id":4,"amount":100}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var tx Transaction
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	require.Equal(t, 4, tx.OperationTypeID)
	require.Equal(t, 100.0, tx.Amount)

	// PURCHASE -> negative
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/transactions",
		bytes.NewReader([]byte(`{"account_id":1,"operation_type_id":1,"amount":50}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tx))
	require.Equal(t, 1, tx.OperationTypeID)
	require.Equal(t, -50.0, tx.Amount)
}
