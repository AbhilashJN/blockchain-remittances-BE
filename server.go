package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func pong(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintln(w, "pong")
	}
}

func registration(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		err := db.WriteCustomerDetailsToCommonCustomersDB(r.FormValue("PhoneNumber"), &db.CustomerDetails{
			CustomerName:  r.FormValue("CustomerName"),
			BankName:      r.FormValue("BankName"),
			BankAccountID: r.FormValue("BankAccountID"),
		})
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		cd, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "registration failed: %v", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Registration successful: %s => %+v", r.FormValue("PhoneNumber"), cd)
	}
}

func getReceiverInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		cd, err := db.ReadCustomerDetailsFromCommonCustomersDB(r.FormValue("PhoneNumber"))
		if err != nil {
			fmt.Fprintf(w, "reading failed: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(*cd)
	}
}

func sendPayment(w http.ResponseWriter, r *http.Request) {
	senderBankStellarSeeds, err := db.ReadStellarSeedsOfBank(*bankNameFlag)
	if err != nil {
		fmt.Fprintf(w, "ReadStellarSeedsOfBank failed: %v", err)
		return
	}

	senderBankStellarAddressKP, err := GetSIDkeyPairsOfBank(senderBankStellarSeeds)
	if err != nil {
		fmt.Fprintf(w, "GetSIDkeyPairsOfBank failed: %v", err)
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		senderName := r.FormValue("senderName")
		senderBankAccountID := r.FormValue("senderBankAccountID")
		// receiverName := r.FormValue("receiverName")
		receiverBank := r.FormValue("receiverBankName")
		receiverBankAccountID := r.FormValue("receiverBankAccountID")

		receiverBankStellarSeeds, err := db.ReadStellarSeedsOfBank(receiverBank)
		if err != nil {
			fmt.Fprintf(w, "ReadStellarSeedsOfBank failed: %v", err)
			return
		}

		receiverBankStellarAddressKP, err := GetSIDkeyPairsOfBank(receiverBankStellarSeeds)
		if err != nil {
			fmt.Fprintf(w, "GetSIDkeyPairsOfBank failederror: %v", err)
			return
		}

		if err := sendAssetFromAtoB(senderBankStellarAddressKP.Distributor,
			receiverBankStellarAddressKP.Distributor,
			senderBankStellarSeeds.DistributorSeed,
			buildAsset(senderBankStellarAddressKP.Issuer, *bankNameFlag+"T"),
			r.FormValue("Amount"),
			fmt.Sprintf("%s;%s;%s", receiverBankAccountID, senderBankAccountID, senderName)); err != nil {
			fmt.Fprintf(w, "error: %v", err)
			return
		}

		fmt.Fprintf(w, "success")
	}
}

//StartServer starts the server
func StartServer() {
	http.HandleFunc("/ping", pong)
	http.HandleFunc("/registration", registration)
	http.HandleFunc("/getReceiverInfo", getReceiverInfo)
	http.HandleFunc("/sendPayment", sendPayment)
	http.ListenAndServe(fmt.Sprintf("localhost:%s", *portNumFlag), nil)
}
