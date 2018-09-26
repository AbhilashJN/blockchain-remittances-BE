package recieve

import (
	"context"
	"fmt"

	"github.com/stellar/go/clients/horizon"
)

// ListenForPayments returns
func ListenForPayments(accountAddress string, messages chan string) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamPayments(ctx, accountAddress, &cursor, func(payment horizon.Payment) {
		handlePayment(accountAddress, payment)
	})

	if err != nil {
		panic(err)
	}

}

func handlePayment(accountAddress string, payment horizon.Payment) {
	if payment.To != accountAddress {
		return
	}
	amount := payment.Amount
	fromCustomerBankAccount := payment.Memo.Value
	fromStellarAccount := payment.From
}
