package httpx

import (
	"errors"
	"net/http"

	"mvc-service-repo/internal/model"
	"mvc-service-repo/internal/validate"
)

// MapDomainError translates a domain or validation error returned by the
// service layer into the appropriate HTTP status code and JSON error body.
// The controller's only job, when a service call fails, is to call this.
func MapDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, validate.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrCustomerNotFound),
		errors.Is(err, model.ErrPurchaseNotFound):
		WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, model.ErrEmailAlreadyExists),
		errors.Is(err, model.ErrAlreadyRefunded),
		errors.Is(err, model.ErrPointsAlreadySpent),
		errors.Is(err, model.ErrInsufficientPoints):
		WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, model.ErrUnknownReward):
		WriteError(w, http.StatusBadRequest, err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
