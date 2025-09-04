package main

import (
	"errors"

	"gorm.io/gorm"
)

func seedOps(db *gorm.DB) error {
	ops := []OperationType{
		{ID: 1, Description: "PURCHASE"},
		{ID: 2, Description: "INSTALLMENT PURCHASE"},
		{ID: 3, Description: "WITHDRAWAL"},
		{ID: 4, Description: "PAYMENT"},
	}
	for _, op := range ops {
		var existing OperationType
		if err := db.First(&existing, op.ID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&op).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}
