package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/AbhilashJN/blockchain-remittances-BE/models"
	"github.com/AbhilashJN/blockchain-remittances-BE/transaction"

	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/stellar/go/clients/horizon"
)

func pong(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "pong")
	}
}

// ListenForPayments returns
func ListenForPayments(bank BankConfig) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, bank.StellarAddresses.Distributor, &cursor,
		func(transaction horizon.Transaction) {
			if err := receivePayment(bank, transaction); err != nil {
				log.Printf("In callback of StreamTransactions: %s", err.Error())
			}
		},
	)

	if err != nil {
		fmt.Printf("shit happened")
		panic(err)
	}

}

func receivePayment(bank BankConfig, transaction horizon.Transaction) error {
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
	receiverAccountID, senderAccountID, senderName := fields[0], fields[1], fields[2]
	operation := txe.Tx.Operations[0].Body.PaymentOp
	amount := float64(operation.Amount) / 1e7 // TODO: Verify the validity of this
	assetInfo, ok := operation.Asset.GetAlphaNum4()
	if !ok {
		return errors.New("GetAlphaNum4() failed: Could not extract alpha4 asset from the envelope operation")
	}

	fmt.Printf("Asset code: %q\n", assetInfo.AssetCode)
	fmt.Printf("Amount: %f\n", amount)
	fmt.Printf("From bank account: %q, name: %q \n", senderAccountID, senderName)
	fmt.Printf("Bank account to credit: %q\n", receiverAccountID)

	var receiverAccount models.Account
	var bankPoolAccount models.Account

	if err := bank.DB.Where("ID = ?", receiverAccountID).First(&receiverAccount).Error; err != nil {
		return err
	}
	if err := bank.DB.Where("ID = ?", bank.BankPoolAccID).First(&bankPoolAccount).Error; err != nil {
		return err
	}
	if err := bank.DB.Find(&receiverAccount).Update("Balance", receiverAccount.Balance+amount).Error; err != nil {
		return err
	}
	if err := bank.DB.Find(&bankPoolAccount).Update("Balance", bankPoolAccount.Balance-amount).Error; err != nil {
		return err
	}

	receiverTransactionDetails := models.Transaction{AccountID: receiverAccountID, Name: senderName, TransactionType: "credit", From: senderAccountID, Amount: amount, ID: transaction.ID}

	bankPoolTransactionDetails := models.Transaction{AccountID: bank.BankPoolAccID, Name: receiverAccount.Name, TransactionType: "debit", To: receiverAccountID, Amount: amount, ID: fmt.Sprintf("POOLTORCVR:%s", utils.CreateRandomString())}

	if err := bank.DB.Create(&receiverTransactionDetails).Error; err != nil {
		return err
	}
	if err := bank.DB.Create(&bankPoolTransactionDetails).Error; err != nil {
		return err
	}

	return nil
}

func sendPayment(w http.ResponseWriter, r *http.Request, bank BankConfig) {

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		amountQueryKey, senderNameQueryKey, receiverNameQueryKey, senderBankAccIDQueryKey := "Amount", "SenderName", "ReceiverName", "SenderBankAccountID"
		receiverBankAccIDQueryKey, receiverBankStellarDistAddressQueryKey := "ReceiverBankAccountID", "ReceiverBankStellarDistributorAddress"

		for _, fieldName := range []string{amountQueryKey, senderNameQueryKey, receiverNameQueryKey, senderBankAccIDQueryKey, receiverBankAccIDQueryKey, receiverBankStellarDistAddressQueryKey} {
			_, ok := r.PostForm[fieldName]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%s field not found in the request's body", fieldName)
				return
			}
		}

		senderName, receiverName, senderBankAccountID := r.PostFormValue(senderNameQueryKey), r.PostFormValue(receiverNameQueryKey), r.PostFormValue(senderBankAccIDQueryKey)
		receiverBankAccountID, receiverBankStellarDistributorAddress := r.PostFormValue(receiverBankAccIDQueryKey), r.PostFormValue(receiverBankStellarDistAddressQueryKey)
		amountToCredit := r.PostFormValue(amountQueryKey)

		amountInFloat, err := strconv.ParseFloat(amountToCredit, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "strconv.ParseFloat(amountToCredit, 64) failed\n: %v", err)
			return
		}

		resp, err := transaction.SendPaymentTransaction(amountToCredit, bank.StellarAddresses.Distributor, receiverBankStellarDistributorAddress,
			bank.StellarSeeds.Distributor, fmt.Sprintf("%s;%s;%s", receiverBankAccountID, senderBankAccountID, senderName),
			utils.BuildAsset(bank.StellarAddresses.Issuer, bank.StellarAssetCode))

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error in send payment transaction: %v", err)
			return
		}

		fmt.Printf("Successful payment transaction by %q on the stellar network\n:", bank.Name)

		var senderAccTransactionDetails = models.Transaction{AccountID: senderBankAccountID, Name: receiverName, TransactionType: "debit", To: receiverBankAccountID, Amount: amountInFloat, ID: resp.Hash}
		var poolAccTransactionDetails = models.Transaction{AccountID: bank.BankPoolAccID, Name: senderName, TransactionType: "credit", From: senderBankAccountID, Amount: amountInFloat, ID: fmt.Sprintf("SNDRTOPOOl:%s", utils.CreateRandomString())}

		var senderAccount models.Account
		var bankPoolAccount models.Account

		if err := bank.DB.Where("ID = ?", senderBankAccountID).First(&senderAccount).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Where("ID = ?", senderBankAccountID).First(&senderAccount).Error failed: %v`, err)
			return
		}

		if err := bank.DB.Where("ID = ?", bank.BankPoolAccID).First(&bankPoolAccount).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Where("ID = ?", senderBankAccountID).First(&senderAccount).Error failed: %v`, err)
			return
		}

		if err := bank.DB.First(&senderAccount).Update("Balance", senderAccount.Balance-amountInFloat).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Model(&senderAccount).Update("Balance", senderAccount.Balance-amountInFloat).Error failed: %v`, err)
			return
		}

		if err := bank.DB.First(&bankPoolAccount).Update("Balance", bankPoolAccount.Balance+amountInFloat).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Model(&senderAccount).Update("Balance", senderAccount.Balance-amountInFloat).Error failed: %v`, err)
			return
		}

		if err := bank.DB.Create(&senderAccTransactionDetails).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Create(&senderTransactionDetails).Error failed: %v`, err)
			return
		}

		if err := bank.DB.Create(&poolAccTransactionDetails).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Create(&senderTransactionDetails).Error failed: %v`, err)
			return
		}

		fmt.Fprintf(w, "success")
	}
}

func getAccountDetails(w http.ResponseWriter, r *http.Request, bank BankConfig) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		bankAccountIDQueryKey := "BankAccountID"

		if _, ok := r.Form[bankAccountIDQueryKey]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s parameter not found in the query string", bankAccountIDQueryKey)
			return
		}

		bankAccountID := r.FormValue(bankAccountIDQueryKey)

		var account models.Account
		if err := bank.DB.Where("ID = ?", bankAccountID).Preload("Transactions").First(&account).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "account not found")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Where("ID = ?", bankAccountID).First(&account).Error failed:\n %v`, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(account); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "json.NewEncoder(w).Encode(account) failed:\n %v", err)
			return
		}
	}
}

func withdrawAmount(w http.ResponseWriter, r *http.Request, bank BankConfig) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		amountQueryKey, accountIDQueryKey := "Amount", "AccountID"

		for _, fieldName := range []string{amountQueryKey, accountIDQueryKey} {
			_, ok := r.PostForm[fieldName]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%s field not found in the request's body", fieldName)
				return
			}
		}
		amount, accountID := r.PostFormValue(amountQueryKey), r.PostFormValue(accountIDQueryKey)
		amountInFloat, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "strconv.ParseFloat(amount, 64) failed\n: %v", err)
			return
		}

		var accountTxDetails = models.Transaction{AccountID: accountID, Name: "Self", TransactionType: "debit", To: "", Amount: amountInFloat, ID: utils.CreateRandomString()}

		var userAccount models.Account
		if err := bank.DB.Where("ID = ?", accountID).First(&userAccount).Update("Balance", userAccount.Balance-amountInFloat).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, fmt.Sprintf("%q account not found", accountID))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Where("ID = ?", accountID).First(&userAccount).Update("Balance", userAccount.Balance-amountInFloat).Error failed: %v`, err)
			return
		}

		if err := bank.DB.Create(&accountTxDetails).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Create(&accountTxDetails).Error failed: %v`, err)
			return
		}

		fmt.Fprintf(w, "success")
	}
}

func depositAmount(w http.ResponseWriter, r *http.Request, bank BankConfig) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		amountQueryKey, accountIDQueryKey := "Amount", "AccountID"

		for _, fieldName := range []string{amountQueryKey, accountIDQueryKey} {
			_, ok := r.PostForm[fieldName]
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "%s field not found in the request's body", fieldName)
				return
			}
		}
		amount, accountID := r.PostFormValue(amountQueryKey), r.PostFormValue(accountIDQueryKey)
		amountInFloat, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "strconv.ParseFloat(amount, 64) failed\n: %v", err)
			return
		}

		var accountTxDetails = models.Transaction{AccountID: accountID, Name: "Self", TransactionType: "credit", To: "", Amount: amountInFloat, ID: utils.CreateRandomString()}

		var userAccount models.Account
		if err := bank.DB.Where("ID = ?", accountID).First(&userAccount).Update("Balance", userAccount.Balance+amountInFloat).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, fmt.Sprintf("%q account not found", accountID))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Where("ID = ?", accountID).First(&userAccount).Update("Balance", userAccount.Balance-amountInFloat).Error failed: %v`, err)
			return
		}

		if err := bank.DB.Create(&accountTxDetails).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `bank.DB.Create(&accountTxDetails).Error failed: %v`, err)
			return
		}

		fmt.Fprintf(w, "success")
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, BankConfig), bank BankConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, bank)
	}
}

//StartServer starts the server
func StartServer(bank BankConfig) {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/sendPayment", makeHandler(sendPayment, bank))
	http.HandleFunc("/accountDetails", makeHandler(getAccountDetails, bank))
	http.HandleFunc("/withdrawAmount", makeHandler(withdrawAmount, bank))
	http.HandleFunc("/depositAmount", makeHandler(depositAmount, bank))
	fmt.Println("\n\nserver is starting...")
	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", bank.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
