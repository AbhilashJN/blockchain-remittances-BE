package receive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/AbhilashJN/blockchain-remittances-BE/bank"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"

	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/stellar/go/clients/horizon"
)

// ListenForPayments returns
func ListenForPayments(bank *bank.Bank) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, bank.StellarAddresses.Distributor, &cursor,
		func(transaction horizon.Transaction) {
			if err := handleTransaction(bank, transaction); err != nil {
				log.Printf("In callback of StreamTransactions: %s", err.Error())
			}
		},
	)

	if err != nil {
		fmt.Printf("shit happened")
		panic(err)
	}

}

func handleTransaction(bank *bank.Bank, transaction horizon.Transaction) error {
	if bank.StellarAddresses.Distributor == transaction.Account {
		return nil
	}

	if bank.StellarAddresses.Issuer == transaction.Account {
		fmt.Println("transaction from issuer account")
		return nil
	}

	fmt.Println("\n\nReceived a transaction from stellar network..")

	txe, err := utils.DecodeTransactionEnvelope(transaction.EnvelopeXdr)
	if err != nil {
		return err
	}
	// spew.Dump(transaction) //pretty print function
	fields := strings.Split(transaction.Memo, ";")
	customerAccountIDtoCredit, senderAccountID, senderName := fields[0], fields[1], fields[2]
	operation := txe.Tx.Operations[0].Body.PaymentOp
	amount := float64(operation.Amount) / 1e7 // TODO: Verify the validity of this
	assetInfo, ok := operation.Asset.GetAlphaNum4()
	if !ok {
		return errors.New("GetAlphaNum4() failed: Could not extract alpha4 asset from the envelope operation")
	}

	transactionDetails := &data.TransactionDetails{TransactionType: "credit", From: senderAccountID, Amount: amount, TransactionID: transaction.ID}

	fmt.Printf("Asset code: %q\n", assetInfo.AssetCode)
	fmt.Printf("Amount: %f\n", transactionDetails.Amount)
	fmt.Printf("From bank account: %q, name: %q \n", transactionDetails.From, senderName)
	fmt.Printf("Bank account to credit: %q\n", customerAccountIDtoCredit)
	updatedCustomerAccountInfo, updatedBankPoolAccountInfo, err := bank.UpdateCustomerBankAccountBalence(transactionDetails, customerAccountIDtoCredit)
	if err != nil {
		return err
	}
	utils.LogAccountDetails(updatedCustomerAccountInfo, updatedBankPoolAccountInfo)
	return nil
}
