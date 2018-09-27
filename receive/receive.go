package receive

import (
	"context"
	"encoding/base64"
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
		log.Println(err)
		return err
	}
	// spew.Dump(transaction) //pretty print function
	fields := strings.Split(transaction.Memo, ";")
	accountIDtoCredit, senderAccountID := fields[0], fields[1]
	amount := float64(tx.Tx.Operations[0].Body.PaymentOp.Amount) / 1e7

	transactionDetails := &db.TransactionDetails{From: senderAccountID, To: accountIDtoCredit, Amount: amount, TransactionID: transaction.ID}

	fmt.Printf("Transaction Details: %+v \n", transactionDetails)
	if err = db.UpdateCustomerBankAccountBalence(bankName, transactionDetails, "credit"); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
