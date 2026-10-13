package usecase

import (
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"example.com/fx/internal/billing/domain"
	ledger "example.com/fx/internal/ledger/domain"
	"example.com/fx/internal/ledger/infra"
)
