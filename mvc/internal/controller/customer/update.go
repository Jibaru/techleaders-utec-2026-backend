package customer

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"mvc-coffee-loyalty/internal/controller/httpx"
	"mvc-coffee-loyalty/internal/model"
	customerview "mvc-coffee-loyalty/internal/view/customer"
)

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseID(r, "id")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid customer id")
		return
	}

	var req customerview.UpdateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := httpx.ValidateName(name); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		updates["name"] = name
	}
	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if err := httpx.ValidateEmail(email); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		updates["email"] = email
	}
	if len(updates) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	var customer model.Customer
	err = c.db.WithContext(r.Context()).First(&customer, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "customer not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "could not load customer")
		return
	}

	if err := c.db.WithContext(r.Context()).Model(&customer).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			httpx.WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "could not update customer")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, customerview.NewResponse(customer))
}
