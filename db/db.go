package db

import (
	"errors"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"
	"github.com/AbhilashJN/blockchain-remittances-BE/utils"
	"github.com/boltdb/bolt"
)

// WriteCustomerBankAccountDetails returns
func WriteCustomerBankAccountDetails(bankName string, customerBankAccountID string, customerBankAccountDetails *data.CustomerBankAccountDetails) error {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	if err := db.Update(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			encoded, err := utils.GobEncode(customerBankAccountDetails)
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
func ReadCustomerBankAccountDetails(bankName, customerBankAccountID string) (*data.CustomerBankAccountDetails, error) {
	// Open the database.
	db, err := bolt.Open(bankName+".db", 0666, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	bucketName := "AccountDetails"

	var accountDetails data.Data

	if err := db.View(
		func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(bucketName))
			key := []byte(customerBankAccountID)

			accountDetails, err = utils.GobDecode(bucket.Get(key), &data.CustomerBankAccountDetails{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}
	customerAccountDetails, ok := accountDetails.(*data.CustomerBankAccountDetails)
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

			_, err = tx.CreateBucket([]byte("StellarSeeds"))
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
func WriteStellarSeedsForBank(bankName string, seeds *data.StellarSeeds) error {
	db, err := bolt.Open(bankName+".db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarSeeds"))
			key := []byte(bankName)

			encoded, err := utils.GobEncode(seeds)
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
func ReadStellarSeedsOfBank(bankName string) (*data.StellarSeeds, error) {
	db, err := bolt.Open(bankName+".db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data.Data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("StellarSeeds"))
			key := []byte(bankName)

			dataVal, err = utils.GobDecode(b.Get(key), &data.StellarSeeds{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	stellarAddress, ok := dataVal.(*data.StellarSeeds)
	if !ok {
		return nil, errors.New("Could not read Stellar Bank Address : Type assertion failed")
	}
	return stellarAddress, nil
}

// CreateCustomerPoolDB returns
func CreateCustomerPoolDB() error {
	db, err := bolt.Open("CustomerPool.db", 0600, nil)
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

//ReadCustomerDetailsFromCustomerPoolDB returns
func ReadCustomerDetailsFromCustomerPoolDB(phoneNumber string) (*data.CustomerDetails, error) {
	db, err := bolt.Open("CustomerPool.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var dataVal data.Data
	if err := db.View(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			dataVal, err = utils.GobDecode(b.Get(key), &data.CustomerDetails{})
			if err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	cd, ok := dataVal.(*data.CustomerDetails)
	if !ok {
		return nil, errors.New("Could not read Customer Details: Type assertion failed")
	}

	return cd, nil
}

// WriteCustomerDetailsToCustomerPoolDB returns
func WriteCustomerDetailsToCustomerPoolDB(phoneNumber string, customerDetails *data.CustomerDetails) error {
	db, err := bolt.Open("CustomerPool.db", 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("CustomerDetails"))
			key := []byte(phoneNumber)

			encoded, err := utils.GobEncode(customerDetails)
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
