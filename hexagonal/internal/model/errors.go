package model

import "errors"

// Domain errors. Returned by repositories (after translating infra errors)
// and propagated by services. Controllers map them to HTTP status codes.
var (
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrPurchaseNotFound   = errors.New("purchase not found")
	ErrDuplicatePurchase  = errors.New("duplicate purchase for that external_payment_id")
	ErrAlreadyRefunded    = errors.New("purchase already refunded")
	ErrPointsAlreadySpent = errors.New("cannot refund: customer has already spent those points")
	ErrInsufficientPoints = errors.New("insufficient points to redeem reward")
	ErrUnknownReward      = errors.New("unknown reward type")
)
