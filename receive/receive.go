package receive

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/stellar/go/clients/horizon"
)

// ListenForPayments returns
func ListenForPayments(bankName, distributorStellarAddressOfBank, issuerStellarAddressOfBank string) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, distributorStellarAddressOfBank, &cursor, func(transaction horizon.Transaction) {
		if err := handleTransaction(bankName, distributorStellarAddressOfBank, issuerStellarAddressOfBank, transaction); err != nil {
			log.Printf("In callback of StreamTransactions: %s", err.Error())
		}

	})

	if err != nil {
		fmt.Printf("shit happened")
		panic(err)
	}

}

func handleTransaction(bankName, distributorStellarAddressOfBank, issuerStellarAddressOfBank string, transaction horizon.Transaction) error {
	if distributorStellarAddressOfBank == transaction.Account {
		return nil
	}

	if issuerStellarAddressOfBank == transaction.Account {
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

	transactionDetails := &db.TransactionDetails{TransactionType: "credit", From: senderAccountID, Amount: amount, TransactionID: transaction.ID}

	fmt.Printf("Asset code: %q\n", assetInfo.AssetCode)
	fmt.Printf("Amount: %f\n", transactionDetails.Amount)
	fmt.Printf("From bank account: %q, name: %q \n", transactionDetails.From, senderName)
	fmt.Printf("Bank account to credit: %q\n", customerAccountIDtoCredit)

	updatedCustomerAccountInfo, updatedBankPoolAccountInfo, err := db.UpdateCustomerBankAccountBalence(transactionDetails, bankName, customerAccountIDtoCredit)
	if err != nil {
		return err
	}
	// spew.Dump(updatedAccountDetails)
	fmt.Println("\n\nReceiver customer bank account details after succesful transaction")
	fmt.Printf("Account holder name: %q\n", updatedCustomerAccountInfo.Name)
	fmt.Printf("Account holder balance: %f\n", updatedCustomerAccountInfo.Balance)
	fmt.Println("--------------------------------------------Transaction history----------------------------------------------")
	for _, tx := range updatedCustomerAccountInfo.Transactions {
		fmt.Printf("TransactionID: %q\nTransactionType: %q\nFrom: %q\nAmount: %f\n", tx.TransactionID, tx.TransactionType, tx.From, tx.Amount)
		fmt.Println("------------------------------------------------------------------------------------------")
	}
	fmt.Println("--------------------------------------------Transaction history----------------------------------------------")

	fmt.Println("\n\nBank pool account detils")
	// fmt.Printf("updatedBankPoolAccountInfo : %T\n\n", updatedBankPoolAccountInfo)
	fmt.Printf("Balance: %f\n", updatedBankPoolAccountInfo.Balance)
	fmt.Println("--------------------------------------------Transaction history----------------------------------------------")
	for _, tx := range updatedBankPoolAccountInfo.Transactions {
		fmt.Printf("TransactionID: %q\nTransactionType: %q\nTo: %q\n, Amount: %f\n", tx.TransactionID, tx.TransactionType, tx.To, tx.Amount)
	}
	fmt.Println("--------------------------------------------Transaction history----------------------------------------------")
	return nil
}
