package main

import (
	"fmt"
	"log"

	"github.com/AbhilashJN/blockchain-remittances-BE/db"
)

func main() {
	// SourceSeed:SBNSSMFUYGUIPXMOEKSAGD524THQKE6S5NUEGZVZ3LNB422G427DNAIL IssuerSeed:SBZK7WTUPFIY55HRRW4SYH2KFZVMLH7STS2PTQHOW4RAF7UWRB74BGKM DistributorSeed:SDP3446MMEKWDSUKJ65GASOBNL6L42UVBPAXKKGOMCSVARB3WMLHJVQY
	// SourceSeed:SB7SZBJ5BUYWQGQZ3TVSTDN7FY7F6ERPFI22TBJRDBYHOKUDRVLG4KFW IssuerSeed:SBJW3HMQG4AUCFO63YGNCBJRQAODMGOOQBLQT74TBSUYHYPDU56WZB3V DistributorSeed:SDZ2T2L4LT76MIPUIA5LJLYGPDSF4CWIWVID5U32H6N62N5OEZ2VSTBX
	fmt.Println("seeding SBI...")
	if err := db.WriteCustomerBankAccountDetails("SBI", "123ABC", &db.CustomerBankAccountDetails{Name: "Sreekar", Balance: 500.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("SBI", "456DEF", &db.CustomerBankAccountDetails{Name: "Abhilash", Balance: 1000.0}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeding JPMORGAN...")
	if err := db.WriteCustomerBankAccountDetails("JPM", "789GHI", &db.CustomerBankAccountDetails{Name: "Milan", Balance: 1500.0}); err != nil {
		log.Fatal(err)
	}
	if err := db.WriteCustomerBankAccountDetails("JPM", "321KLM", &db.CustomerBankAccountDetails{Name: "Sandeep", Balance: 2000.0}); err != nil {
		log.Fatal(err)
	}

	fmt.Println("seeding BankStellarAdresses...")
	if err := db.WriteStellarAddressesForBank("SBI", &db.StellarAddressesOfBank{
		SourceSeed:      "SBNSSMFUYGUIPXMOEKSAGD524THQKE6S5NUEGZVZ3LNB422G427DNAIL",
		IssuerSeed:      "SBZK7WTUPFIY55HRRW4SYH2KFZVMLH7STS2PTQHOW4RAF7UWRB74BGKM",
		DistributorSeed: "SDP3446MMEKWDSUKJ65GASOBNL6L42UVBPAXKKGOMCSVARB3WMLHJVQY",
	}); err != nil {
		log.Fatal(err)
	}

	if err := db.WriteStellarAddressesForBank("JPM", &db.StellarAddressesOfBank{
		SourceSeed:      "SB7SZBJ5BUYWQGQZ3TVSTDN7FY7F6ERPFI22TBJRDBYHOKUDRVLG4KFW",
		IssuerSeed:      "SBJW3HMQG4AUCFO63YGNCBJRQAODMGOOQBLQT74TBSUYHYPDU56WZB3V",
		DistributorSeed: "SDZ2T2L4LT76MIPUIA5LJLYGPDSF4CWIWVID5U32H6N62N5OEZ2VSTBX",
	}); err != nil {
		log.Fatal(err)
	}
}
