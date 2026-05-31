package mail

import (
	"fmt"

	"mvc-service-repo/internal/model"
)

// Templates return (subject, body) pairs for each event. Centralizing them
// here means both the regular purchase endpoint and the webhook produce the
// identical email — and there's exactly one place to change the wording.

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
