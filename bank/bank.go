package bank

import (
	"errors"
	"fmt"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/boltdb/bolt"
)

// Bank is
type Bank struct {
	Name string
	*data.StellarSeeds
	*data.StellarAddresses
}

// UpdateCustomerBankAccountBalence returns
func (bank *Bank) UpdateCustomerBankAccountBalence(txDetails *data.TransactionDetails, customerAccountID string) (*data.CustomerBankAccountDetails, *data.CustomerBankAccountDetails, error) {
	// Open the database.
	bankName := bank.Name
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"
	bankPoolAccountID := bankName + "-POOL-ID"
	// fmt.Printf("bankPoolAccountID : %q", bankPoolAccountID)

	var updatedRecordOfCustomer *data.CustomerBankAccountDetails
	var updatedRecordOfBank *data.CustomerBankAccountDetails

	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))

			accDetOfCustomer, err := utils.GobDecode(bucket.Get([]byte(customerAccountID)), &data.CustomerBankAccountDetails{})
			if err != nil {
				return fmt.Errorf("gobDecode(bucket.Get([]byte(customerAccountID)), &data.CustomerBankAccountDetails{}) failed: %s", err.Error())
			}

			accDetOfBankPool, err := utils.GobDecode(bucket.Get([]byte(bankPoolAccountID)), &data.CustomerBankAccountDetails{})
			if err != nil {
				return fmt.Errorf("gobDecode(bucket.Get([]byte(bankPoolAccountID)), &data.CustomerBankAccountDetails{}) failed: %s", err.Error())
			}

			customerAccountDetails, ok := accDetOfCustomer.(*data.CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}
			bankPoolaccountDetails, ok := accDetOfBankPool.(*data.CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}

			switch txDetails.TransactionType {
			case "credit":
				// fmt.Printf("In credit case")
				bankPoolaccountDetails.Balance -= txDetails.Amount
				customerAccountDetails.Balance += txDetails.Amount
				customerAccountDetails.Transactions = append(customerAccountDetails.Transactions, *txDetails)
				bankPoolaccountDetails.Transactions = append(bankPoolaccountDetails.Transactions, data.TransactionDetails{
					To: customerAccountID, TransactionType: "debit", Amount: txDetails.Amount, TransactionID: fmt.Sprintf("PoolToCustomerTx:CustomerAccId-%s", customerAccountID),
				})
				// fmt.Printf("bankPoolaccountDetails: %+v", bankPoolaccountDetails)
			case "debit":
				bankPoolaccountDetails.Balance += txDetails.Amount
				customerAccountDetails.Balance -= txDetails.Amount
				customerAccountDetails.Transactions = append(customerAccountDetails.Transactions, *txDetails)
				bankPoolaccountDetails.Transactions = append(bankPoolaccountDetails.Transactions, data.TransactionDetails{
					From: customerAccountID, TransactionType: "credit", Amount: txDetails.Amount, TransactionID: fmt.Sprintf("CustomerToPoolTx::CustomerAccId-%s", customerAccountID),
				})
			default:
				return errors.New("invalid updateType param passed. should be 'credit' or 'debit'")
			}

			customerAccountDetailsEncoded, err := utils.GobEncode(customerAccountDetails)
			if err != nil {
				return fmt.Errorf("gobEncode(customerAccountDetails) failed: %s", err.Error())
			}

			bankPoolaccountDetailsEncoded, err := utils.GobEncode(bankPoolaccountDetails)
			if err != nil {
				return fmt.Errorf("gobEncode(bankPoolaccountDetails) failed: %s", err.Error())
			}

			if err := bucket.Put([]byte(customerAccountID), customerAccountDetailsEncoded); err != nil {
				return fmt.Errorf("bucket.Put([]byte(customerAccountID), customerAccountDetailsEncoded) failed: %s", err.Error())
			}

			if err := bucket.Put([]byte(bankPoolAccountID), bankPoolaccountDetailsEncoded); err != nil {
				return fmt.Errorf("bucket.Put([]byte(bankPoolAccountID), bankPoolaccountDetailsEncoded) failed: %s", err.Error())
			}

			updatedRecordOfCustomer = customerAccountDetails // copy in another var to return the updated record to the caller
			updatedRecordOfBank = bankPoolaccountDetails
			return nil
		},
	); err != nil {
		return nil, nil, err
	}
	fmt.Printf("UpdateCustomerBankAccountBalence successful")
	return updatedRecordOfCustomer, updatedRecordOfBank, nil
}

// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
// func (bank *Bank) method() {

// }
