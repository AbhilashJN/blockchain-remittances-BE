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
	fmt.Println("Payment type", payment.Type)
	fmt.Println("Payment Paging Token", payment.PagingToken)
	fmt.Println("Payment From", payment.From)
	fmt.Println("Payment To", payment.To)
	fmt.Println("Payment Asset Type", payment.AssetType)
	fmt.Println("Payment Asset Code", payment.AssetCode)
	fmt.Println("Payment Asset Issuer", payment.AssetIssuer)
	fmt.Println("Payment Amount", payment.Amount)
	fmt.Println("Payment Memo Type", payment.Memo.Type)
	fmt.Println("Payment Memo", payment.Memo.Value)
	fmt.Println("Payment Funder", payment.Funder)
	fmt.Println("Payment ID", payment.ID)
	fmt.Println("Payment Into", payment.Into)
	fmt.Println("Payment Links", payment.Links)
	fmt.Println("Payment SourceAccount", payment.SourceAccount)
	fmt.Println("Payment Account", payment.Account)
	fmt.Println("Payment TransactionHash", payment.TransactionHash)
}
