package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

// StellarSeedsOfBank is

type data interface {
	String()
}

//StellarSeedsOfBank implements data, contains
type StellarSeedsOfBank struct {
	SourceSeed, IssuerSeed, DistributorSeed string
}

//TransactionDetails contains
type TransactionDetails struct {
	From, To, TransactionID string
	TransactionType         string // debit or credit
	Amount                  float64
}

//CustomerBankAccountDetails implements data, contains
type CustomerBankAccountDetails struct {
	Name         string
	Balance      float64
	Transactions []TransactionDetails
}

// CustomerDetails implements data, contains
type CustomerDetails struct {
	CustomerName, BankName, BankAccountID string
}

func (sab *StellarSeedsOfBank) String() {
	fmt.Printf("%+v \n", sab)
}

func (sab *CustomerBankAccountDetails) String() {
	fmt.Printf("%+v \n", sab)
}

func (sab *CustomerDetails) String() {
	fmt.Printf("%+v \n", sab)
}

//gobDecode returns
// func decodeByteSlice(dBdata []byte, target data) (data, error) {
// 	fmt.Printf("\n\n In decodeByteSlice, dBdata: %s \n\n", dBdata)
// 	err := json.Unmarshal(dBdata, target)
// 	if err != nil {
// 		return nil, fmt.Errorf("json.Unmarshal(dBdata, target) failed: %s", err.Error())
// 	}

// 	return target, nil
// }

func gobEncode(structValue interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(structValue)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte, target data) (data, error) {
	// fmt.Printf("In gobDecode: %s", data)
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

// UpdateCustomerBankAccountBalence returns
func UpdateCustomerBankAccountBalence(txDetails *TransactionDetails, bankName, customerAccountID string) (*CustomerBankAccountDetails, *CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"
	bankPoolAccountID := bankName + "-POOL-ID"
	// fmt.Printf("bankPoolAccountID : %q", bankPoolAccountID)

	var updatedRecordOfCustomer *CustomerBankAccountDetails
	var updatedRecordOfBank *CustomerBankAccountDetails

	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))

			accDetOfCustomer, err := gobDecode(bucket.Get([]byte(customerAccountID)), &CustomerBankAccountDetails{})
			if err != nil {
				return fmt.Errorf("gobDecode(bucket.Get([]byte(customerAccountID)), &CustomerBankAccountDetails{}) failed: %s", err.Error())
			}

			accDetOfBankPool, err := gobDecode(bucket.Get([]byte(bankPoolAccountID)), &CustomerBankAccountDetails{})
			if err != nil {
				return fmt.Errorf("gobDecode(bucket.Get([]byte(bankPoolAccountID)), &CustomerBankAccountDetails{}) failed: %s", err.Error())
			}

			customerAccountDetails, ok := accDetOfCustomer.(*CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}
			bankPoolaccountDetails, ok := accDetOfBankPool.(*CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}

			switch txDetails.TransactionType {
			case "credit":
				fmt.Printf("In credit case")
				bankPoolaccountDetails.Balance -= txDetails.Amount
				customerAccountDetails.Balance += txDetails.Amount
				customerAccountDetails.Transactions = append(customerAccountDetails.Transactions, *txDetails)
				bankPoolaccountDetails.Transactions = append(bankPoolaccountDetails.Transactions, TransactionDetails{
					To: customerAccountID, TransactionType: "debit", Amount: txDetails.Amount, TransactionID: fmt.Sprintf("PoolToCustomerTx:CustomerAccId-%s", customerAccountID),
				})
				fmt.Printf("bankPoolaccountDetails: %+v", bankPoolaccountDetails)
			case "debit":
				bankPoolaccountDetails.Balance += txDetails.Amount
				customerAccountDetails.Balance -= txDetails.Amount
				customerAccountDetails.Transactions = append(customerAccountDetails.Transactions, *txDetails)
				bankPoolaccountDetails.Transactions = append(bankPoolaccountDetails.Transactions, TransactionDetails{
					From: customerAccountID, TransactionType: "credit", Amount: txDetails.Amount, TransactionID: fmt.Sprintf("CustomerToPoolTx::CustomerAccId-%s", customerAccountID),
				})
			default:
				return errors.New("invalid updateType param passed. should be 'credit' or 'debit'")
			}

			customerAccountDetailsEncoded, err := gobEncode(customerAccountDetails)
			if err != nil {
				return fmt.Errorf("gobEncode(customerAccountDetails) failed: %s", err.Error())
			}

			bankPoolaccountDetailsEncoded, err := gobEncode(bankPoolaccountDetails)
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

	return updatedRecordOfCustomer, updatedRecordOfBank, nil
}

// WriteCustomerBankAccountDetails returns
func WriteCustomerBankAccountDetails(bankName string, customerBankAccountID string, customerBankAccountDetails *CustomerBankAccountDetails) error {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	// Execute several commands within a read-write transaction.
	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			encoded, err := gobEncode(customerBankAccountDetails)
			if err != nil {
				return err
			}

			if err := bucket.Put(key, encoded); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// ReadCustomerBankAccountDetails returns
func ReadCustomerBankAccountDetails(bankName, customerBankAccountID string) (*CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	var accountDetails data
	// Read the value back from a separate read-only transaction.
	if err := db.View(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			accountDetails, err = gobDecode(bucket.Get(key), &CustomerBankAccountDetails{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}
	customerAccountDetails, ok := accountDetails.(*CustomerBankAccountDetails)
	if !ok {
		return nil, errors.New("Could not read Customer Bank Account Details : Type assertion failed")
	}
	return customerAccountDetails, nil
}

// CreateDBForBank returns
func CreateDBForBank(bankName string) error {
	db, err := bolt.Open(bankName+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("AccountDetails"))
			if err != nil {
				return err
			}

			// encoded, err := gobEncode(&CustomerBankAccountDetails{
			// 	Balance: 10000000.0,
			// 	Name:    bankName,
			// })
			// if err != nil {
			// 	return err
			// }

			// if err := b.Put([]byte(bankName), encoded); err != nil {
			// 	return err
			// }

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// CreateBankStellarSeedsDB returns
func CreateBankStellarSeedsDB() error {
	db, err := bolt.Open("BankStellarSeeds.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("StellarSeeds"))
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// WriteStellarSeedsForBank returns
func WriteStellarSeedsForBank(bankName string, addresses *StellarSeedsOfBank) error {
	db, err := bolt.Open("BankStellarSeeds.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarSeeds"))
			key := []byte(bankName)

			encoded, err := gobEncode(addresses)
			if err != nil {
				return err
			}

			err = b.Put(key, encoded)
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

// ReadStellarSeedsOfBank returns
func ReadStellarSeedsOfBank(bankName string) (*StellarSeedsOfBank, error) {
	db, err := bolt.Open("BankStellarSeeds.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarSeeds"))
			key := []byte(bankName)

			dataVal, err = gobDecode(b.Get(key), &StellarSeedsOfBank{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	stellarAddress, ok := dataVal.(*StellarSeedsOfBank)
	if !ok {
		return nil, errors.New("Could not read Stellar Bank Address : Type assertion failed")
	}
	return stellarAddress, nil
}

// CreateCommonCustomersDB returns
func CreateCommonCustomersDB() error {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("CustomerDetails"))
			if err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

//ReadCustomerDetailsFromCommonCustomersDB returns
func ReadCustomerDetailsFromCommonCustomersDB(phoneNumber string) (*CustomerDetails, error) {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			dataVal, err = gobDecode(b.Get(key), &CustomerDetails{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	cd, ok := dataVal.(*CustomerDetails)
	if !ok {
		return nil, errors.New("Could not read Customer Details: Type assertion failed")
	}

	return cd, nil
}

// WriteCustomerDetailsToCommonCustomersDB returns
func WriteCustomerDetailsToCommonCustomersDB(phoneNumber string, customerDetails *CustomerDetails) error {
	db, err := bolt.Open("CommonCustomers.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			encoded, err := gobEncode(customerDetails)
			if err != nil {
				return err
			}

			err = b.Put(key, encoded)
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}
