package main

import (
	"fmt"

	"github.com/stripe/stripe-go"

	stripeClient "github.com/stripe/stripe-go/client"
)

// PayoutWorkflow is awesome
type PayoutWorkflow struct {
	DineroAPI *dineroAPI
	StripeAPI *stripeClient.API
}

// HandleStripePayout says it all...
func (workflow *PayoutWorkflow) HandleStripePayout(p stripe.Payout) error {
	params := &stripe.BalanceTransactionListParams{}
	params.Payout = stripe.String(p.ID)

	i := workflow.StripeAPI.Balance.List(params)

	var feeAmount int64
	for i.Next() {
		bt := i.BalanceTransaction()
		feeAmount += bt.Fee
	}

	amount := float64(p.Amount) / 100
	fee := float64(feeAmount) / 100

	err := workflow.DineroAPI.AddStripePayout(p.ID, amount, fee)

	if err != nil {
		return fmt.Errorf("Stripe Payout - failed to update ledger with payout info: %s", err.Error())
	}

	return nil
}
