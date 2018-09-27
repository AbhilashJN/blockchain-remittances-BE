package db

import (
	"encoding/json"
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
	Amount                  float64
}

//CustomerBankAccountDetails implements data, contains
type CustomerBankAccountDetails struct {
	Name         string
	Balance      float64
	Transactions []*TransactionDetails
}

//CustomerDetails implements data, contains
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

//decodeByteSlice returns
func decodeByteSlice(dBdata []byte, target data) (data, error) {
	err := json.Unmarshal(dBdata, target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// UpdateCustomerBankAccountBalence returns
func UpdateCustomerBankAccountBalence(bankName string, transactionDetails *TransactionDetails, updateType string) (*CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	var updatedRecord *CustomerBankAccountDetails
	// Execute several commands within a read-write transaction.
	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			customerAccountID := []byte(transactionDetails.To)
			// bankPoolAccountID := []byte(transactionDetails.To)

			accDetOfCustomer, err := decodeByteSlice(bucket.Get(customerAccountID), &CustomerBankAccountDetails{})
			if err != nil {
				return err
			}

			// accDetOfBankPool, err := decodeByteSlice(bucket.Get(bankPoolAccountID), &CustomerBankAccountDetails{})
			// if err != nil {
			// 	return err
			// }

			customerAccountDetails, ok := accDetOfCustomer.(*CustomerBankAccountDetails)
			if !ok {
				return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			}
			// bankPoolaccountDetails, ok := accDetOfBankPool.(*CustomerBankAccountDetails)
			// if !ok {
			// 	return errors.New("Could not update Customer Bank Account Details : Type assertion failed")
			// }

			switch updateType {
			case "credit":
				customerAccountDetails.Balance += transactionDetails.Amount
			case "debit":
				customerAccountDetails.Balance -= transactionDetails.Amount
			default:
				return errors.New("invalid updateType param passed. should be 'credit' or 'debit'")
			}

			encoded, err := json.Marshal(customerAccountDetails)
			if err != nil {
				return err
			}

			if err := bucket.Put(customerAccountID, encoded); err != nil {
				return err
			}

			updatedRecord = customerAccountDetails // copy in another var to return the updated record to the caller

			return nil
		},
	); err != nil {
		return nil, err
	}

	return updatedRecord, nil
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

			encoded, err := json.Marshal(customerBankAccountDetails)
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

			accountDetails, err = decodeByteSlice(bucket.Get(key), &CustomerBankAccountDetails{})
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
			b, err := tx.CreateBucket([]byte("AccountDetails"))
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(&CustomerBankAccountDetails{
				Balance: 10000000.0,
				Name:    bankName,
			})
			if err != nil {
				return err
			}

			if err := b.Put([]byte(bankName), encoded); err != nil {
				return err
			}

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

			encoded, err := json.Marshal(addresses)
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

			dataVal, err = decodeByteSlice(b.Get(key), &StellarSeedsOfBank{})
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

			dataVal, err = decodeByteSlice(b.Get(key), &CustomerDetails{})
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

			encoded, err := json.Marshal(customerDetails)
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
