package main

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"

	"github.com/AbhilashJN/blockchain-remittances-BE/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

// DatabaseConfig is config for database
type DatabaseConfig struct {
	Host, Port, User, Password, DbName string
}

// BankConfig is a config for a bank
type BankConfig struct {
	BankName      string
	Database      DatabaseConfig
	BankPoolAccID string
}

// ConfigForAllBanks is config for database
type ConfigForAllBanks struct {
	Banks []BankConfig
}

var configForAllBanks ConfigForAllBanks

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // ensures that Viper will read from environment variables as well.

	// Searches for config file in given paths and read it
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Confirm which config file is used
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())

	if err := viper.Unmarshal(&configForAllBanks); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func init() {
	readConfig()
	spew.Dump(configForAllBanks)
}

var sbiAccounts = []models.Account{
	models.Account{
		ID: "123456789SBIPOOL123456789", Balance: 50000000, Name: "SBI-POOL",
	},
	models.Account{
		ID: "123ABC", Balance: 500, Name: "Sreekar",
		// Transactions: []models.Transaction{
		// 	{From: "789GHI", Amount: 5, TransactionType: "credit", ID: "276XFHGSJGC7D6CSDCDBCGSDGV7S8VJDHSF8S8SD7SDDS9DNHGD6556"},
		// },
	},
	models.Account{
		ID: "456DEF", Balance: 1000, Name: "Abhilash",
		// Transactions: []models.Transaction{
		// 	{From: "321KLM", Amount: 15, TransactionType: "credit", ID: "F7SD6F8SF6DS8FDS5SD65F76FSDF6S8F68S6FSD8F6SD8F6786F88"},
		// },
	},
}
var jpmAccounts = []models.Account{
	models.Account{
		ID: "987654321JPMORGANPOOL987654321", Balance: 70000000, Name: "JPMORGAN-POOL",
	},
	models.Account{
		ID: "789GHI", Balance: 1500, Name: "Milan",
		// Transactions: []models.Transaction{
		// 	{To: "123ABC", Amount: 5, TransactionType: "debit", ID: "276XFHGSJGC7D6CSDCDBCGSDGV7S8VJDHSF8S8SD7SDDS9DNHGD6556"},
		// },
	},
	models.Account{
		ID: "321KLM", Balance: 2000, Name: "Sandeep",
		// Transactions: []models.Transaction{
		// 	{To: "456DEF", Amount: 15, TransactionType: "debit", ID: "F7SD6F8SF6DS8FDS5SD65F76FSDF6S8F68S6FSD8F6SD8F6786F88"},
		// },
	},
}

func seedTables(bank string, db *gorm.DB) {
	var accounts []models.Account
	if bank == "SBI" {
		accounts = sbiAccounts
	} else {
		accounts = jpmAccounts
	}
	// spew.Dump(accounts)
	for _, account := range accounts {
		if result := db.Create(&account); result.Error != nil {
			fmt.Println(result.Error)
		}
	}
}

func main() {
	for _, bankConfig := range configForAllBanks.Banks {
		database := bankConfig.Database
		dbConnectionParams := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", database.Host, database.Port, database.User, database.DbName, database.Password)

		db, err := gorm.Open("postgres", dbConnectionParams)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		db.DropTableIfExists(&models.Transaction{}, &models.Account{})                                    // transactions table should be dropped before accounts table
		db.AutoMigrate(&models.Account{}, &models.Transaction{})                                          // auto-migration of foreign keys does'nt happen
		db.Model(&models.Transaction{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE") // Foreign key need to define manually

		seedTables(bankConfig.BankName, db)
	}
}
