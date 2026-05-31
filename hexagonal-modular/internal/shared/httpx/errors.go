package httpx

import (
	"errors"
	"net/http"

	customermodel "hexagonal-modular/internal/customer/model"
	purchasemodel "hexagonal-modular/internal/purchase/model"
	rewardmodel "hexagonal-modular/internal/reward/model"
	"hexagonal-modular/internal/shared/validate"
)

// MapDomainError translates a domain or validation error returned by the
// service layer into the appropriate HTTP status code and JSON error body.
// The controller's only job, when a service call fails, is to call this.
//
// This is one of the few files in the codebase that imports all three
// modules' models — the price of having a single cross-module error mapper.
func MapDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, validate.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, customermodel.ErrCustomerNotFound),
		errors.Is(err, purchasemodel.ErrPurchaseNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, customermodel.ErrEmailAlreadyExists),
		errors.Is(err, purchasemodel.ErrAlreadyRefunded),
		errors.Is(err, purchasemodel.ErrPointsAlreadySpent),
		errors.Is(err, rewardmodel.ErrInsufficientPoints):
		WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, rewardmodel.ErrUnknownReward):
		WriteError(w, http.StatusBadRequest, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
