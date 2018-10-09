package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
func listenForPayments(bank BankConfig) {
	ctx := context.Background()

	cursor := horizon.Cursor("now")

	fmt.Println("Waiting for a payment...")

	err := horizon.DefaultTestNetClient.StreamTransactions(ctx, bank.StellarAddresses.Distributor, &cursor,
		func(transaction horizon.Transaction) {
			if bank.StellarAddresses.Distributor == transaction.Account {
				log.Printf("\ntransaction created by same bank\n")
				return
			}

			if bank.StellarAddresses.Issuer == transaction.Account {
				log.Printf("\ntransaction created by issuer account\n")
				return
			}

			if err := receivePayment(bank, transaction); err != nil {
				log.Printf("\nIn callback of StreamTransactions: %s\n", err.Error())
			}
		},
	)

	if err != nil {
		fmt.Printf("shit happened")
		panic(err)
	}

}

func receivePayment(bank BankConfig, transaction horizon.Transaction) error {

	log.Println("\n\nReceived a transaction from stellar network..")

	paymentInfo, err := utils.DecodeTransactionEnvelope(transaction)
	if err != nil {
		return err
	}

	var receiverAccount models.Account
	var bankPoolAccount models.Account

	if err := bank.DB.Where("ID = ?", paymentInfo.ReceiverAccountID).First(&receiverAccount).Error; err != nil {
		return err
	}
	if err := bank.DB.Where("ID = ?", bank.BankPoolAccID).First(&bankPoolAccount).Error; err != nil {
		return err
	}
	if err := bank.DB.Find(&receiverAccount).Update("Balance", receiverAccount.Balance+paymentInfo.Amount).Error; err != nil {
		return err
	}
	if err := bank.DB.Find(&bankPoolAccount).Update("Balance", bankPoolAccount.Balance-paymentInfo.Amount).Error; err != nil {
		return err
	}

	receiverTransactionDetails := models.Transaction{AccountID: paymentInfo.ReceiverAccountID, Name: paymentInfo.SenderName, TransactionType: "credit", From: paymentInfo.SenderAccountID, Amount: paymentInfo.Amount, TxID: paymentInfo.TxID}

	bankPoolTransactionDetails := models.Transaction{AccountID: bank.BankPoolAccID, Name: receiverAccount.Name, TransactionType: "debit", To: paymentInfo.ReceiverAccountID, Amount: paymentInfo.Amount, TxID: fmt.Sprintf("POOLTORCVR:%s", utils.CreateRandomString())}

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
			http.Error(w, fmt.Sprintf("ParseForm() err: %v", err), http.StatusBadRequest)
			return
		}

		amountQueryKey, senderNameQueryKey, receiverNameQueryKey, senderBankAccIDQueryKey := "Amount", "SenderName", "ReceiverName", "SenderBankAccountID"
		receiverBankAccIDQueryKey, receiverBankStellarDistAddressQueryKey := "ReceiverBankAccountID", "ReceiverBankStellarDistributorAddress"

		for _, fieldName := range []string{amountQueryKey, senderNameQueryKey, receiverNameQueryKey, senderBankAccIDQueryKey, receiverBankAccIDQueryKey, receiverBankStellarDistAddressQueryKey} {
			_, ok := r.PostForm[fieldName]
			if !ok {
				http.Error(w, fmt.Sprintf("%s field not found in the request's body", fieldName), http.StatusBadRequest)
				return
			}
		}

		senderName, receiverName, senderBankAccountID := r.PostFormValue(senderNameQueryKey), r.PostFormValue(receiverNameQueryKey), r.PostFormValue(senderBankAccIDQueryKey)
		receiverBankAccountID, receiverBankStellarDistributorAddress := r.PostFormValue(receiverBankAccIDQueryKey), r.PostFormValue(receiverBankStellarDistAddressQueryKey)
		amountToCredit := r.PostFormValue(amountQueryKey)

		amountInFloat, err := strconv.ParseFloat(amountToCredit, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if receiverBankStellarDistributorAddress == bank.StellarAddresses.Distributor {
			var senderAccount models.Account
			var receiverAccount models.Account
			if err := bank.DB.Where("ID = ?", senderBankAccountID).First(&senderAccount).Update("Balance", senderAccount.Balance-amountInFloat).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := bank.DB.Where("ID = ?", receiverBankAccountID).First(&receiverAccount).Update("Balance", receiverAccount.Balance+amountInFloat).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			txID := utils.CreateRandomString()
			var senderTransactionDetails = models.Transaction{AccountID: senderBankAccountID, TxID: txID, To: receiverBankAccountID, TransactionType: "debit", Name: receiverName, Amount: amountInFloat}
			var receiverTransactionDetails = models.Transaction{AccountID: receiverBankAccountID, TxID: txID, From: senderBankAccountID, TransactionType: "credit", Name: senderName, Amount: amountInFloat}

			if err := bank.DB.Create(&senderTransactionDetails).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := bank.DB.Create(&receiverTransactionDetails).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "success")
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

		log.Printf("Successful payment transaction by %q on the stellar network\n:", bank.Name)

		var senderAccTransactionDetails = models.Transaction{AccountID: senderBankAccountID, Name: receiverName, TransactionType: "debit", To: receiverBankAccountID, Amount: amountInFloat, TxID: resp.Hash}
		var poolAccTransactionDetails = models.Transaction{AccountID: bank.BankPoolAccID, Name: senderName, TransactionType: "credit", From: senderBankAccountID, Amount: amountInFloat, TxID: fmt.Sprintf("SNDRTOPOOl:%s", utils.CreateRandomString())}

		var senderAccount models.Account
		var bankPoolAccount models.Account

		if err := bank.DB.Where("ID = ?", senderBankAccountID).First(&senderAccount).Update("Balance", senderAccount.Balance-amountInFloat).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := bank.DB.Where("ID = ?", bank.BankPoolAccID).First(&bankPoolAccount).Update("Balance", bankPoolAccount.Balance+amountInFloat).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := bank.DB.Create(&senderAccTransactionDetails).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := bank.DB.Create(&poolAccTransactionDetails).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "success")
	}
}

func getAccountDetails(w http.ResponseWriter, r *http.Request, bank BankConfig) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("ParseForm() err: %v", err), http.StatusBadRequest)
			return
		}

		bankAccountIDQueryKey := "BankAccountID"

		if _, ok := r.Form[bankAccountIDQueryKey]; !ok {
			http.Error(w, fmt.Sprintf("%s parameter not found in the query string", bankAccountIDQueryKey), http.StatusBadRequest)
			return
		}

		bankAccountID := r.FormValue(bankAccountIDQueryKey)

		var account models.Account
		if err := bank.DB.Where("ID = ?", bankAccountID).Preload("Transactions").First(&account).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				http.Error(w, "account not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(account); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		var accountTxDetails = models.Transaction{AccountID: accountID, Name: "Self", TransactionType: "debit", To: "", Amount: amountInFloat, TxID: utils.CreateRandomString()}

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

		var accountTxDetails = models.Transaction{AccountID: accountID, Name: "Self", TransactionType: "credit", To: "", Amount: amountInFloat, TxID: utils.CreateRandomString()}

		var userAccount models.Account
		if err := bank.DB.Where("ID = ?", accountID).First(&userAccount).Update("Balance", userAccount.Balance+amountInFloat).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "%q account not found", accountID)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
