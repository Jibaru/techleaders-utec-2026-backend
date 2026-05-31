package purchase

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"hexagonal/internal/mail"
	"hexagonal/internal/model"
	"hexagonal/internal/repository"
	"hexagonal/internal/validate"
)

// RecordInput is the unified input for recording a purchase from either the
// customer-facing API or an inbound payment webhook. ExternalPaymentID enables
// idempotency — if a purchase already exists for that id, the existing
// purchase is returned and Duplicate is true.
type RecordInput struct {
	CustomerID        uuid.UUID
	AmountCents       int64
	ExternalPaymentID *string
	OccurredAt        time.Time
}

type RecordResult struct {
	Purchase     model.Purchase
	Customer     model.Customer
	PreviousTier model.Tier
	NewTier      model.Tier
	Duplicate    bool
}

func (s *Service) Record(ctx context.Context, in RecordInput) (RecordResult, error) {
	if err := validate.Amount(in.AmountCents); err != nil {
		return RecordResult{}, err
	}
	if in.ExternalPaymentID != nil {
		trimmed := strings.TrimSpace(*in.ExternalPaymentID)
		if err := validate.ExternalID(trimmed); err != nil {
			return RecordResult{}, err
		}
		in.ExternalPaymentID = &trimmed
	}

	// Idempotency check: if we have already processed this external_payment_id,
	// return the existing purchase rather than creating a duplicate.
	if in.ExternalPaymentID != nil {
		existing, err := s.purchases.FindByExternalPaymentID(ctx, *in.ExternalPaymentID)
		switch {
		case err == nil:
			customer, err := s.customers.FindByID(ctx, existing.CustomerID)
			if err != nil {
				return RecordResult{}, err
			}
			tier := model.TierForPoints(customer.Points)
			return RecordResult{
				Purchase:     existing,
				Customer:     customer,
				PreviousTier: tier,
				NewTier:      tier,
				Duplicate:    true,
			}, nil
		case errors.Is(err, model.ErrPurchaseNotFound):
			// expected — fall through to creation
		default:
			return RecordResult{}, err
		}
	}

	customer, err := s.customers.FindByID(ctx, in.CustomerID)
	if err != nil {
		return RecordResult{}, err
	}

	previousTier := model.TierForPoints(customer.Points)
	pointsEarned := computePoints(in.AmountCents, previousTier.Multiplier)

	createdAt := in.OccurredAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	purchase := model.Purchase{
		ID:                uuid.New(),
		CustomerID:        customer.ID,
		AmountCents:       in.AmountCents,
		PointsEarned:      pointsEarned,
		ExternalPaymentID: in.ExternalPaymentID,
		CreatedAt:         createdAt,
	}

	err = s.tx.RunTx(ctx, func(repos repository.TxRepos) error {
		if err := repos.Purchases.Save(ctx, &purchase); err != nil {
			return err
		}
		customer.Points += pointsEarned
		return repos.Customers.UpdatePoints(ctx, customer.ID, customer.Points)
	})

	// Race: another concurrent webhook processed the same external_payment_id.
	// Re-fetch and treat as idempotent success.
	if errors.Is(err, model.ErrDuplicatePurchase) && in.ExternalPaymentID != nil {
		existing, lookupErr := s.purchases.FindByExternalPaymentID(ctx, *in.ExternalPaymentID)
		if lookupErr != nil {
			return RecordResult{}, lookupErr
		}
		customer, lookupErr := s.customers.FindByID(ctx, existing.CustomerID)
		if lookupErr != nil {
			return RecordResult{}, lookupErr
		}
		tier := model.TierForPoints(customer.Points)
		return RecordResult{
			Purchase:     existing,
			Customer:     customer,
			PreviousTier: tier,
			NewTier:      tier,
			Duplicate:    true,
		}, nil
	}
	if err != nil {
		return RecordResult{}, err
	}

	newTier := model.TierForPoints(customer.Points)

	// Side-effects: receipt + (if applicable) tier upgrade emails.
	// Log-and-continue on send failures so a flaky SMTP doesn't fail the call
	// after the DB transaction already committed.
	subject, body := mail.PurchaseReceipt(customer, in.AmountCents, pointsEarned, newTier)
	if err := s.mailer.Send(ctx, customer.Email, subject, body); err != nil {
		slog.ErrorContext(ctx, "send purchase receipt", "err", err, "customer_id", customer.ID)
	}
	if previousTier.Name != newTier.Name {
		subject, body := mail.TierUpgrade(customer, newTier)
		if err := s.mailer.Send(ctx, customer.Email, subject, body); err != nil {
			slog.ErrorContext(ctx, "send tier upgrade", "err", err, "customer_id", customer.ID)
		}
	}

	return RecordResult{
		Purchase:     purchase,
		Customer:     customer,
		PreviousTier: previousTier,
		NewTier:      newTier,
		Duplicate:    false,
	}, nil
}

// computePoints turns dollars into loyalty points using the tier multiplier.
// 1 dollar = 1 base point, rounded down after the multiplier is applied.
func computePoints(amountCents int64, multiplier float64) int {
	dollars := float64(amountCents) / 100.0
	return int(math.Floor(dollars * multiplier))
}
