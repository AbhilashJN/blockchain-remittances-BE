## Horizon RESTful API Server.
- Applications interact with the Stellar network through the "Horizon" RESTful HTTP API server.
- As the server communicates through HTTP, a browser, cURL, and the StellarSDK of supported languages can talk to the server.

## Stellar Core
- The Horizon server talks to the Stellar Core instance to do all the work.
- Stellar Core software does all the hard work of validating and agreeing with other instances of Stellar Core on the status of every transaction through the Stellar Consensus Protocol(SCP).
- Some Stellar Cores have the Horizon server also with them.
- You can host a Stellar Core instance without relying on a third-party.

## Stellar Network
- Stellar network is a worldwide collection of Stellar Cores run by various induviduals and orgs.

## Testnet
- Testnet is a small Stellar Network for testing applications open to developers.
- This network has 3 Stellar Cores.

## Accounts
- Accounts hold all your money inside Stellar and allow you to send and receive payments.
- Every Stellar account has a public key and a secret seed.
- The public key is always safe to share.
- Other people need it to identify your account and verify that you authorized a transaction.
- The seed, however, is private information that proves you own your account.
- You should never share the seed with anyone.

## Secret Seed
- The seed is actually the single secret piece of data that is used to generate both the public and private key for your account.
- Stellar’s tools use the seed instead of the private key for convenience.
- To have full access to an account, you only need to provide a seed instead of both a public key and a private key.

## Lumens and Stroops
- Lumens are the built-in currency of the Stellar network.
- 1 Stroop = 0.0000001 Lumen.
- Stroops are easier to talk about than such tiny fractions of a lumen.

## Creating an Account
- Each account must have a minimum balance of 1 lumen.
- In the real world, you’ll usually pay an exchange that sells lumens in order to create a new account.
- On the testnet it is free to create an account.
- You’ll need to only send the public key to a Stellar server to create an account.
- The created account ID is equal to the public key.

## Account ID
- The public key is used as the accountID.
- So for any activity with the Stellar Network, only the public key can be used as the account identifier.
- The seed also can be used as the accountID, since the public key can be generated from the seed.s

## Balences in Accounts
-  Accounts can carry multiple balances - one for each type of currency they hold.

## Operations
- Actions that change things in Stellar, like sending payments, changing your account, or making offers to trade various kinds of currencies, are called operations.

## Transaction
- In order to actually perform an operation, you create a transaction.
- A transaction is just a group of operations accompanied by some extra information, like what account is making the transaction and a cryptographic signature to verify that the transaction is authentic.
- If any operation in the transaction fails, they all fail.
  - For example, let’s say you have 100 lumens and you make two payment operations of 60 lumens each. 
  - If you make two transactions (each with one operation), the first will succeed and the second will fail because you don’t have enough lumens. You’ll be left with 40 lumens. 
  - However, if you group the two payments into a single transaction, they will both fail and you’ll be left with the full 100 lumens still in your account.
- Every transaction costs a small fee called **base fee**.
- It is 100 stroops / operation.

## Building a transaction
1. Confirm that the account ID you are sending to actually exists by loading the associated account data from the Stellar network through the horizon API server.
    1. You can also use this call to perform any other verification you might want to do on a destination account.
    2. **If you are writing banking software, for example, this is a good place to insert regulatory compliance checks and KYC verification.**
2. Load data for the account you are sending from. 
    1. An account can only perform one transaction at a time and has something called a sequence number.
        1. **In situations where you need to perform a high number of transactions in a short period of time (for example, a bank might perform transactions on behalf of many customers using one Stellar account), you can create several Stellar accounts that work simultaneously.**
    2. A Sequence number helps Stellar verify the order of transactions.
    3. A transaction’s sequence number needs to match the account’s sequence number, so you need to get the account’s current sequence number from the network.
3. Start building a transaction. 
    1. This requires an account object, not just an account ID, because it will increment the account’s sequence number.
4. Add the payment operation to the account.
    1. Note that you need to specify the type of asset you are sending.
    2. **Stellar’s “native” currency is the lumen, but you can send any type of asset or currency you like, from dollars to bitcoin to any sort of asset you trust the issuer to redeem.**
    3. NOTE:- You should also note that the amount is a **string** rather than a **number**. When working with extremely small fractions or large values, floating point math can introduce small inaccuracies.
5. Optionally, you can add your own metadata, called a memo, to a transaction. 
    1. Stellar doesn’t do anything with this data, but you can use it for any purpose you’d like. 
    2. **If you are a bank that is receiving or sending payments on behalf of other people, for example, you might include the actual person the payment is meant for here.**
6. **Now that the transaction has all the data it needs, you have to cryptographically sign it using your secret seed.**
    1. This proves that the data actually came from you and not someone impersonating you.
7. And finally, send it to the Stellar network through the Horizon server.

## Transaction status
- It’s possible that you will not receive a response from Horizon server due to a bug, network conditions, etc. 
- In such situation it’s impossible to determine the status of your transaction. 
- That’s why you should always save a built transaction (or transaction encoded in XDR format) in a variable or a database and resubmit it if you don’t know it’s status. 
- If the transaction has already been successfully applied to the ledger, Horizon will simply return the saved result and not attempt to submit the transaction again. 
- Only in cases where a transaction’s status is unknown (and thus will have a chance of being included into a ledger) will a resubmission to the network occur.

## Receive Payments
- You don’t actually need to do anything to receive payments into a Stellar account.
- If a payer makes a successful transaction to send assets to you, those assets will automatically be added to your account.
- However, you’ll want to know that someone has actually paid you. 
    - **If you are a bank accepting payments on behalf of others, you need to find out what was sent to you so you can disburse funds to the intended recipient.** 
    - If you are operating a retail business, you need to know that your customer actually paid you before you hand them their merchandise. 
    - And if you are an automated rental car with a Stellar account, you’ll probably want to verify that the customer in your front seat actually paid before that person can turn on your engine.

## Transacting in Other Currencies
- One of the amazing things about the Stellar network is that you can send and receive many types of assets, such as US dollars, Nigerian naira, digital currencies like bitcoin, **or even your own new kind of asset.**
- While Stellar’s native asset, the lumen, is fairly simple, **all other assets can be thought of like a credit issued by a particular account.**
- **In fact, when you trade US dollars on the Stellar network, you don’t actually trade US dollars—you trade US dollars from a particular account.**
- Assets have 3 props
    - Type
    - Code
    - Issuer - The issuer is the ID of the account that created the asset. 
- **Understanding what account issued the asset is important.**
- **You need to trust that, if you want to redeem your dollars on the Stellar network for actual dollar bills, the issuer will be able to provide them to you**.
- **Because of this, you’ll usually only want to trust major financial institutions for assets that represent national currencies.**

## Multi-currency transactions.
- You can send Nigerian naira to a friend in Germany and have them receive euros.
- These multi-currency transactions are made possible by a built-in market mechanism where people can make offers to buy and sell different types of assets.
-  Stellar will automatically find the best people to exchange currencies with in order to convert your naira to euros. 
-  **This system is called distributed exchange.**

## Anchor
- Anchors are entities that people trust to hold their deposits and issue credits into the Stellar network for those deposits.
- **All money transactions in the Stellar network (except lumens) occur in the form of credit issued by anchors, so anchors act as a bridge between existing currencies and the Stellar network.**
- **Most anchors are organizations like banks, savings institutions, farmers’ co-ops, central banks, and remittance companies.**

## Assets are credits from a particular account
- One of Stellar’s most powerful features is the ability to trade any kind of asset, US dollars, Nigerian naira, bitcoins, special coupons, ICO tokens or just about anything you like.
- This works in Stellar because an asset is really just a credit from a particular account.  
- **When you trade US dollars on the Stellar network, you don’t actually trade US dollars—you trade US dollars credited from a particular account.**
- Often, that account will be a bank, but if your neighbor had a banana plant, they might issue banana assets that you could trade with other people.

## Asset properties
- Asset_Code 
    - A short identifier of 1–12 letters or numbers, such as USD, or EUR. It can be anything you like, even AstroDollars.
- Asset_Issuer
    - The ID of the account that issues the asset.

## Issuing a New Asset Type
- To issue a new type of asset, all you need to do is choose a code. 
- It can be any combination of up to 12 letters or numbers, but you should use the appropriate ISO 4217 code (e.g. USD for US dollars) or ISIN for national currencies or securities. 
- Once you’ve chosen a code, you can begin paying people using that asset code.
- You don’t need to do anything to declare your asset on the network.
- However, other people can’t receive your asset until they’ve chosen to trust it. 
- Because a Stellar asset is really a credit, you should trust that the issuer can redeem that credit if necessary later on. 
- You might not want to trust your neighbor to issue banana assets if they don’t even have a banana plant, for example.

## Trustline
- **An account can create a trustline, or a declaration that it trusts a particular asset, using the change trust operation.**
- A trustline can also be limited to a particular amount. 
- If your banana-growing neighbor only has a few plants, you might not want to trust them for more than about 200 bananas. 
- Note: each trustline increases an account’s minimum balance by 0.5 lumens (the base reserve). For more details, see the fees guide.
- **Once you’ve chosen an asset code and someone else has created a trustline for your asset, you’re free to start making payment operations to them using your asset.**
- If someone you want to pay doesn’t trust your asset, you might also be able to use the distributed exchange.

## Discoverablity and Meta information
- Another thing that is important when you issue an asset is to provide clear information about what your asset represents. 
- This info can be discovered and displayed by clients so users know exactly what they are getting when they hold your asset. 
- To do this you must do two simple things. 
- First, add a section in your stellar.toml file that contains the necessary meta fields:
- ```toml
    # stellar.toml example asset
    [[CURRENCIES]]
    code="GOAT"
    issuer="GD5T6IPRNCKFOHQWT264YPKOZAWUMMZOLZBJ6BNQMUGPWGRLBK3U7ZNP"
    display_decimals=2 
    name="goat share"
    desc="1 GOAT token entitles you to a share of revenue from Elkins Goat Farm."
    conditions="There will only ever be 10,000 GOAT tokens in existence. We will distribute the revenue share annually on Jan. 15th"
    image="https://pbs.twimg.com/profile_images/666921221410439168/iriHah4f.jpg"
    ```
- **Second, use the set options operation to set the home_domain of your issuing account to the domain where the above stellar.toml file is hosted.**

## Stellar Adresses
- Stellar addresses provide an easy way for users to share payment details by using a syntax that interoperates across different domains and providers.
- Stellar addresses are divided into two parts separated by *, the username and the domain.
- For example: jed*stellar.org
    - jed is the username,
    - stellar.org is the domain.

## Federation server
- You can use the federation endpoint to look up an account id if you have a stellar address. 
    - You can also do reverse federation and look up a stellar addresses from account ids or transaction ids. This is useful to see who has sent you a payment.

## Compliance Protocol
- Complying with Anti-Money Laundering (AML) laws requires financial institutions (FIs) to know not only who their customers are sending money to but who their customers are receiving money from. 
- In some jurisdictions banks are able to trust the AML procedures of other licensed banks. 
- In other jurisdictions each bank must do its own sanction checking of both the sender and the receiver. 
- The Compliance Protocol handles all these scenarios.

## Compliance procedure
- The customer information that is exchanged between FIs is flexible but the typical fields are:
  - Full Name
  - Date of birth
  - Physical address
- The Compliance Protocol is an additional step after federation. 
- In this step the sending FI contacts the receiving FI to get permission to send the transaction. 
- To do this the receiving FI creates an AUTH_SERVER and adds its location to the stellar.toml of the FI.
- You can create your own endpoint that implements the compliance protocol or we have also created this simple compliance service that you can use.

## AUTH_SERVER
- The AUTH_SERVER provides one endpoint that is called by a sending FI to get approval to send a payment to one of the receiving FI’s customers. 
- The AUTH_SERVER url should be placed in organization’s stellar.toml file.

## How to become an anchor
- As an anchor, you should maintain at least two accounts:
  - An issuing account used only for issuing and destroying assets.
  - A base account used to transact with other Stellar accounts. It holds a balance of assets issued by the issuing account.

- **Customer Accounts**
  - One way to account for your customers’ funds is:- 
    - **Use federation and the memo field in transactions to send and receive payments on behalf of your customers.**
    - **In this approach, transactions intended for your customers are all made using your base account.**
    - **The memo field of the transaction is used to identify the actual customer a payment is intended for.**
    - **Using a single account requires you to do additional bookkeeping, but means you have fewer keys to manage and more control over accounts.** 
    - **If you already have existing banking systems, this is the simplest way to integrate Stellar with them.**
- **Data flow**
  - In order to act as an anchor, your infrastructure will need to:
    - Make payments.
    - Monitor a Stellar account and update customer accounts when payments are received.
    - Look up and respond to requests for federated addresses.
    - Comply with Anti-Money Laundering (AML) regulations.
  - Stellar provides a prebuilt federation server and regulatory compliance server designed for you to install and integrate with your existing infrastructure. 
  - The bridge server coordinates them and simplifies interacting with the Stellar network.

## Bridge server
- Stellar.org maintains a bridge server, which makes it easier to use the federation and compliance servers to send and receive payments.
- When using the bridge server, the only code you need to write is 
    - A private service to receive payment notifications 
    - And respond to regulatory checks from the bridge and compliance servers.
- When using the bridge server, you send payments by making an HTTP POST request to it instead of a Horizon server. 
- It doesn’t change a whole lot for simple transactions, but it will make the next steps of federation and compliance much simpler.

