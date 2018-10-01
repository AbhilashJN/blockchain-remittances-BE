package main

import (
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/data"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func main() {
	// SourceSeed:SBNSSMFUYGUIPXMOEKSAGD524THQKE6S5NUEGZVZ3LNB422G427DNAIL IssuerSeed:SBZK7WTUPFIY55HRRW4SYH2KFZVMLH7STS2PTQHOW4RAF7UWRB74BGKM DistributorSeed:SDP3446MMEKWDSUKJ65GASOBNL6L42UVBPAXKKGOMCSVARB3WMLHJVQY
	// SourceSeed:SB7SZBJ5BUYWQGQZ3TVSTDN7FY7F6ERPFI22TBJRDBYHOKUDRVLG4KFW IssuerSeed:SBJW3HMQG4AUCFO63YGNCBJRQAODMGOOQBLQT74TBSUYHYPDU56WZB3V DistributorSeed:SDZ2T2L4LT76MIPUIA5LJLYGPDSF4CWIWVID5U32H6N62N5OEZ2VSTBX
	fmt.Println("seeding SBI...")
	if err := db.WriteCustomerBankAccountDetails("SBI", "123ABC", &data.CustomerBankAccountDetails{Name: "Sreekar", Balance: 500.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("SBI", "456DEF", &data.CustomerBankAccountDetails{Name: "Abhilash", Balance: 1000.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("SBI", "SBI-POOL-ID", &data.CustomerBankAccountDetails{Name: "SBI-POOl-ACC", Balance: 1e7}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeding JPMORGAN...")
	if err := db.WriteCustomerBankAccountDetails("JPM", "789GHI", &data.CustomerBankAccountDetails{Name: "Milan", Balance: 1500.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("JPM", "321KLM", &data.CustomerBankAccountDetails{Name: "Sandeep", Balance: 2000.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("JPM", "JPM-POOL-ID", &data.CustomerBankAccountDetails{Name: "JPM-POOL-ACC", Balance: 1e7}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeding BankStellarSeeds...")
	if err := db.WriteStellarSeedsForBank("SBI", &data.StellarSeeds{
		Keys: data.Keys{
			Source:      "SBNSSMFUYGUIPXMOEKSAGD524THQKE6S5NUEGZVZ3LNB422G427DNAIL",
			Issuer:      "SBZK7WTUPFIY55HRRW4SYH2KFZVMLH7STS2PTQHOW4RAF7UWRB74BGKM",
			Distributor: "SDP3446MMEKWDSUKJ65GASOBNL6L42UVBPAXKKGOMCSVARB3WMLHJVQY",
		},
	}); err != nil {
		log.Fatal(err)
	}

	if err := db.WriteStellarSeedsForBank("JPM", &data.StellarSeeds{
		Keys: data.Keys{
			Source:      "SB7SZBJ5BUYWQGQZ3TVSTDN7FY7F6ERPFI22TBJRDBYHOKUDRVLG4KFW",
			Issuer:      "SBJW3HMQG4AUCFO63YGNCBJRQAODMGOOQBLQT74TBSUYHYPDU56WZB3V",
			Distributor: "SDZ2T2L4LT76MIPUIA5LJLYGPDSF4CWIWVID5U32H6N62N5OEZ2VSTBX",
		},
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeding CustomerPool...")
	if err := db.WriteCustomerDetailsToCustomerPoolDB("9976543210", &data.CustomerDetails{CustomerName: "Sreekar", BankName: "SBI", BankAccountID: "123ABC", StellarDistributionAddressOfBank: "GC2JUDOWWCREXJPNGCL4IFBF6C6EVFVEHBSJWQT26T6A63TWIIOYQZQH"}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerDetailsToCustomerPoolDB("9876543210", &data.CustomerDetails{CustomerName: "Abhilash", BankName: "SBI", BankAccountID: "456DEF", StellarDistributionAddressOfBank: "GC2JUDOWWCREXJPNGCL4IFBF6C6EVFVEHBSJWQT26T6A63TWIIOYQZQH"}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerDetailsToCustomerPoolDB("8976543210", &data.CustomerDetails{CustomerName: "Milan", BankName: "JPM", BankAccountID: "789GHI", StellarDistributionAddressOfBank: "GCG2S7CUX4VWXNW5LL3V7CGD36ZBV6TED43LN4B772M5JQ7Z7I43SEOT"}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerDetailsToCustomerPoolDB("8876543210", &data.CustomerDetails{CustomerName: "Sandeep", BankName: "JPM", BankAccountID: "321KLM", StellarDistributionAddressOfBank: "GCG2S7CUX4VWXNW5LL3V7CGD36ZBV6TED43LN4B772M5JQ7Z7I43SEOT"}); err != nil {
		log.Fatal(err)
	}
}
