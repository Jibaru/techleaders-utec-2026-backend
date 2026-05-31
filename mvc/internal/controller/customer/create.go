package customer

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var req customerview.CreateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := httpx.ValidateName(name); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := httpx.ValidateEmail(email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	customer := model.Customer{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Points:    0,
		CreatedAt: time.Now().UTC(),
	}

	if err := c.db.WithContext(r.Context()).Create(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			httpx.WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "could not save customer")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, customerview.NewResponse(customer))
}
