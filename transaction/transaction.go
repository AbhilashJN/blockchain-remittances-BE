package transaction

import (
	"context"
	"fmt"

	"github.com/AbhilashJN/blockchain-remittances-BE/account"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

// WatchLiveActivityOf streams a live activity of an account
func WatchLiveActivityOf(accountAddress string, messages chan string) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamPayments(ctx, accountAddress, &cursor, func(payment horizon.Payment) {
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
		messages <- fmt.Sprintf("payment info of %v printed", accountAddress)
	})

	if err != nil {
		panic(err)
	}

}

// Transact makes a transaction b/w a sender and receiver with an amount
func Transact(sender, receiver account.Account, amount string) {
	fmt.Println("_____________________________________before transaction_____________________________________")
	account.PrintAccountDetails(sender.Address)
	account.PrintAccountDetails(receiver.Address)

	fmt.Printf("Submitting transaction with sender: %v, receiver: %v, amount: %v to to the stellar horizon API server... \n", sender.Address, receiver.Address, amount)
	SubmitTransaction(sender, receiver, amount)

	fmt.Println("_____________________________________after transaction_______________________________________")
	account.PrintAccountDetails(sender.Address)
	account.PrintAccountDetails(receiver.Address)
}

// SubmitTransaction submits a transaction to the stellar horizon API server
func SubmitTransaction(source, destination account.Account, amount string) {
	// Make sure destination account exists
	if _, err := horizon.DefaultTestNetClient.LoadAccount(destination.Address); err != nil {
		panic(err)
	}

	// passphrase := network.TestNetworkPassphrase

	tx, err := build.Transaction(
		build.TestNetwork,
		build.SourceAccount{
			AddressOrSeed: source.Address,
		},
		build.AutoSequence{
			SequenceProvider: horizon.DefaultTestNetClient,
		},
		build.Payment(
			build.Destination{
				AddressOrSeed: destination.Address,
			},
			build.NativeAmount{
				Amount: amount,
			},
		),
	)

	if err != nil {
		panic(err)
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(source.Seed)
	if err != nil {
		panic(err)
	}

	txeB64, err := txe.Base64()
	if err != nil {
		panic(err)
	}

	// And finally, send it off to Stellar!
	resp, err := horizon.DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)
}
