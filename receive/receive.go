package receive

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"
)

// ListenForPayments returns
func ListenForPayments(bankName, stellarAddressOfBank string) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, stellarAddressOfBank, &cursor, func(transaction horizon.Transaction) {
		if err := handleTransaction(bankName, stellarAddressOfBank, transaction); err != nil {
			log.Println(err)
		}

	})

	if err != nil {
		panic(err)
	}

}

func decodeTransactionEnvelope(data string) (xdr.TransactionEnvelope, error) {

	rawr := strings.NewReader(data)
	b64r := base64.NewDecoder(base64.StdEncoding, rawr)

	var tx xdr.TransactionEnvelope
	bytesRead, err := xdr.Unmarshal(b64r, &tx)

	fmt.Printf("read %d bytes\n", bytesRead)

	if err != nil {
		return tx, err
	}
	return tx, nil
}

func handleTransaction(bankName, stellarAddressOfBank string, transaction horizon.Transaction) error {
	if stellarAddressOfBank == transaction.Account {
		return nil
	}

	tx, err := decodeTransactionEnvelope(transaction.EnvelopeXdr)
	if err != nil {
		return err
	}
	// spew.Dump(transaction) //pretty print function
	fields := strings.Split(transaction.Memo, ";")
	accountIDtoCredit, senderAccountID, senderName := fields[0], fields[1], fields[2]
	operation := tx.Tx.Operations[0].Body.PaymentOp
	amount := float64(operation.Amount) / 1e7 // TODO: Verify the validity of this
	assetInfo, ok := operation.Asset.GetAlphaNum4()
	if !ok {
		return errors.New("GetAlphaNum4() failed: Could not extract alpha4 asset from the envelope operation")
	}

	transactionDetails := &db.TransactionDetails{From: senderAccountID, To: accountIDtoCredit, Amount: amount, TransactionID: transaction.ID}
	fmt.Printf("Received a transaction from stellar network..\n")
	fmt.Printf("Asset code: %q\n", assetInfo.AssetCode)
	fmt.Printf("Amount: %f\n", transactionDetails.Amount)
	fmt.Printf("From bank account: %q, name: %q \n", transactionDetails.From, senderName)
	fmt.Printf("To bank account: %q\n", transactionDetails.To)

	updatedAccountDetails, err := db.UpdateCustomerBankAccountBalence(bankName, transactionDetails, "credit")
	if err != nil {
		return err
	}
	// spew.Dump(updatedAccountDetails)
	fmt.Printf("Receiver customer bank account details after succesful transaction\n")
	fmt.Printf("Account holder name: %q\n", updatedAccountDetails.Name)
	fmt.Printf("Account holder balance: %f\n", updatedAccountDetails.Balance)

	return nil
}
