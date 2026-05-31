// Package mail holds the Sender port (interface) plus shared email template
// helpers. Adapters live in sub-packages: mail/smtp/ (real SMTP via net/smtp)
// and mail/noop/ (logs only — for local dev without SMTP credentials).
//
// The Sender port is the second outbound port in the system. Like the
// repository ports, the service layer depends only on the interface.
package mail

import (
	"context"
	"fmt"

	"hexagonal/internal/model"
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

func PurchaseReceipt(customer model.Customer, amountCents int64, pointsEarned int, newTier model.Tier) (string, string) {
	dollars := float64(amountCents) / 100.0
	body := fmt.Sprintf(
		"Hi %s,\n\nThanks for your $%.2f purchase. You earned %d points!\nYour balance is now %d points (%s).\n\n— Tu Café",
		customer.Name, dollars, pointsEarned, customer.Points, newTier.Name,
	)
	return "Thanks for your purchase", body
}

func TierUpgrade(customer model.Customer, newTier model.Tier) (string, string) {
	body := fmt.Sprintf(
		"Hi %s,\n\nCongratulations — you've reached our %s tier!\nYou now earn points at %.2fx on every purchase.\n\n— Tu Café",
		customer.Name, newTier.Name, newTier.Multiplier,
	)
	return fmt.Sprintf("Welcome to %s!", newTier.Name), body
}

func RewardRedeemed(customer model.Customer, rewardType model.RewardType, cost, remaining int) (string, string) {
	body := fmt.Sprintf(
		"Hi %s,\n\nYou redeemed a %s for %d points.\nYour remaining balance is %d points.\n\n— Tu Café",
		customer.Name, rewardType, cost, remaining,
	)
	return "Your reward is ready", body
}
