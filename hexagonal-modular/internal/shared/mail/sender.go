// Package mail holds the Sender port (interface) plus shared email template
// helpers. Adapters live in sub-packages: mail/smtp/ (real SMTP via net/smtp)
// and mail/noop/ (logs only — for local dev without SMTP credentials).
//
// The Sender port is the second outbound port in the system. Like the
// repository ports, the service layer depends only on the interface.
//
// templates.go imports the three modules' models (Customer + Tier + RewardType).
// That's the price of a single source of truth for email wording. Same
// controlled-cross-module pattern as shared/repository/transactor.go and
// shared/httpx/errors.go.
package mail

import (
	"context"
	"fmt"

	customermodel "hexagonal-modular/internal/customer/model"
	rewardmodel "hexagonal-modular/internal/reward/model"
	"hexagonal-modular/internal/shared/model"
)

// Sender is the outbound port for transactional email. Implementations
// deliver the message however they like (SMTP, queue, log, fake).
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// ----- Template helpers ------------------------------------------------------
// Centralizing templates means both the regular purchase endpoint and the
// webhook produce the identical email, and there is exactly one place to
// change the wording.

func PurchaseReceipt(customer customermodel.Customer, amountCents int64, pointsEarned int, newTier model.Tier) (string, string) {
	dollars := float64(amountCents) / 100.0
	body := fmt.Sprintf(
		"Hi %s,\n\nThanks for your $%.2f purchase. You earned %d points!\nYour balance is now %d points (%s).\n\n— Tu Café",
		customer.Name, dollars, pointsEarned, customer.Points, newTier.Name,
	)
	return "Thanks for your purchase", body
}

func TierUpgrade(customer customermodel.Customer, newTier model.Tier) (string, string) {
	body := fmt.Sprintf(
		"Hi %s,\n\nCongratulations — you've reached our %s tier!\nYou now earn points at %.2fx on every purchase.\n\n— Tu Café",
		customer.Name, newTier.Name, newTier.Multiplier,
	)
	return fmt.Sprintf("Welcome to %s!", newTier.Name), body
}

func RewardRedeemed(customer customermodel.Customer, rewardType rewardmodel.RewardType, cost, remaining int) (string, string) {
	body := fmt.Sprintf(
		"Hi %s,\n\nYou redeemed a %s for %d points.\nYour remaining balance is %d points.\n\n— Tu Café",
		customer.Name, rewardType, cost, remaining,
	)
	return "Your reward is ready", body
}
