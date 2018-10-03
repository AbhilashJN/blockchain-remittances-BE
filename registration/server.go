package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

func checkUserValidity(bankName, bankAccountID string, db *gorm.DB) error {
	result := db.Where(&Bank{Name: bankName})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func registerNewUser(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// checkUserValidity(r.FormValue("BankName"), r.FormValue("BankAccountID"), db)

		// 	err := db.WriteCustomerDetailsToCustomerPoolDB(r.FormValue("PhoneNumber"), &data.CustomerDetails{
		// 		BankName:      r.FormValue("BankName"),
		// 		BankAccountID: r.FormValue("BankAccountID"),
		// 		// StellarDistributionAddressOfBank: bank.StellarAddresses.Distributor,
		// 	})
		// 	if err != nil {
		// 		fmt.Fprintf(w, "registration failed: %v", err)
		// 		return
		// 	}

		// 	customerDetails, err := db.ReadCustomerDetailsFromCustomerPoolDB(r.FormValue("PhoneNumber"))
		// 	if err != nil {
		// 		fmt.Fprintf(w, "registration failed: %v", err)
		// 		return
		// 	}

		// 	w.Header().Set("Content-Type", "application/json")
		// 	w.WriteHeader(http.StatusCreated)
		// 	jsEncoder := json.NewEncoder(w)
		// 	err = jsEncoder.Encode(struct {
		// 		CustomerDetails *data.CustomerDetails
		// 		Port            string
		// 	}{
		// 		customerDetails,
		// 		os.Getenv("PORT"),
		// 	})
		// 	if err != nil {
		// 		fmt.Fprintf(w, "jsEncoder.Encode(customerDetails) failed:\n error %v", err)
		// 		return
		// 	}
	}
}

// func makeHandler(fn func(http.ResponseWriter, *http.Request, *gorm.DB), db *gorm.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		fn(w, r, db)
// 	}
// }

// ServerConfig is config for app server
type ServerConfig struct {
	Port string
}

// DatabaseConfig is config for database
type DatabaseConfig struct {
	Host, Port, User, Password, DbName string
}

// Config is config for the entire app
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

var config Config

// User structure defines the model for users table
type User struct {
	gorm.Model
	Name          string
	Bank          Bank `gorm:"foreignkey:BankID"`
	BankID        uint
	BankAccountID string
	PhoneNumber   string
}

// Bank structure defines the model for banks table
type Bank struct {
	gorm.Model
	Name                string
	StellarAppServerURL string
	DistributorAddress  string
}

var users = []User{
	{PhoneNumber: "9976543210", Name: "Sreekar", Bank: Bank{Name: "SBI", StellarAppServerURL: "localhost:7070", DistributorAddress: "GC2JUDOWWCREXJPNGCL4IFBF6C6EVFVEHBSJWQT26T6A63TWIIOYQZQH"}, BankAccountID: "123ABC"},
	{PhoneNumber: "9876543210", Name: "Abhilash", Bank: Bank{Name: "SBI", StellarAppServerURL: "localhost:7070", DistributorAddress: "GC2JUDOWWCREXJPNGCL4IFBF6C6EVFVEHBSJWQT26T6A63TWIIOYQZQH"}, BankAccountID: "456DEF"},
	{PhoneNumber: "8976543210", Name: "Milan", Bank: Bank{Name: "JP MORGAN", StellarAppServerURL: "localhost:6060", DistributorAddress: "GCG2S7CUX4VWXNW5LL3V7CGD36ZBV6TED43LN4B772M5JQ7Z7I43SEOT"}, BankAccountID: "789GHI"},
	{PhoneNumber: "8876543210", Name: "Sandeep", Bank: Bank{Name: "JP MORGAN", StellarAppServerURL: "localhost:6060", DistributorAddress: "GCG2S7CUX4VWXNW5LL3V7CGD36ZBV6TED43LN4B772M5JQ7Z7I43SEOT"}, BankAccountID: "321KLM"},
}

func seedTables(db *gorm.DB) {
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			fmt.Println(err)
		}
	}
}

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

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func init() {
	readConfig()
	// fmt.Printf("%+v\n", config)

}

func main() {
	dbConfig := config.Database
	dbConnectionParams := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DbName, dbConfig.Password)

	db, err := gorm.Open("postgres", dbConnectionParams)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.DropTableIfExists(&User{}, &Bank{})
	db.AutoMigrate(&User{}, &Bank{})
	db.Model(&User{}).AddForeignKey("bank_id", "banks(id)", "CASCADE", "CASCADE")

	seedTables(db)

	// serverAddress := fmt.Sprintf("localhost:%s", config.Server.Port)
	// print(serverAddress)
	// http.HandleFunc("/registerNewUser", registerNewUser)

	// fmt.Println("\n\nRegistartion server is starting...")

	// if err := http.ListenAndServe(serverAddress, nil); err != nil {
	// 	log.Fatal(err)
	// }
}
