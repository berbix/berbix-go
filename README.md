# Berbix Go SDK

This Berbix Go library provides simple interfaces to interact with the Berbix API.

## Installation

    go get github.com/berbix/berbix-go

## Usage

### Constructing a client

    import "github.com/berbix/berbix-go"
    
    client := NewClient(secret, &ClientOptions{})

### Create a transaction

    tokens, err := client.CreateTransaction(&CreateTransactionOptions{
        CustomerUID: "internal_customer_uid", # ID for the user in client database
        TemplateKey: "your_template_key", # Template key for this transaction,
    })

### Create tokens from refresh token

    refreshToken := "" # fetched from database
    transactionTokens := TokensFromRefresh(refreshToken)

### Fetch transaction data

    transactionData, err := client.FetchTransaction(transactionTokens)

## Reference

### `Client`

##### NewClient(secret string, options *ClientOptions) Client

Supported options:

- `HTTPClient` - An optional override for the default HTTP client.

#### Methods

##### `CreateTransaction(options *CreateTransactionOptions) (*Tokens, error)`

Creates a transaction within Berbix to initialize the client SDK. Typically after creating
a transaction, you will want to store the refresh token in your database associated with the
currently active user session.

Supported options:

- `Email` - Previously verified email address for a user.
- `Phone` - Previously verified phone number for a user.
- `CustomerUID` - An ID or identifier for the user in your system.
- `TemplateKey` - The template key for this transaction.

##### `FetchTransaction(tokens *Tokens) (*TransactionMetadata, error)`

Fetches all of the information associated with the transaction. If the user has already completed the steps of the transaction, then this will include all of the elements of the transaction payload as described on the [Berbix developer docs](https://developers.berbix.com).

##### `RefreshTokens(tokens *Tokens) (*Tokens, error)`

This is typically not needed to be called explicitly as it will be called by the higher-level
SDK methods, but can be used to get fresh client or access tokens.

##### `ValidateSignature(secret string, body string, header string) (bool, error)`

This method validates that the content of the webhook has not been forged. This should be called for every endpoint that is configured to receive a webhook from Berbix.

Parameters:

- `secret` - This is the secret associated with that webhook. NOTE: This is distinct from the API secret and can be found on the webhook configuration page of the dashboard.
- `body` - The full request body from the webhook. This should take the raw request body prior to parsing.
- `header` - The value in the 'X-Berbix-Signature' header.

##### `DeleteTransaction(tokens *Tokens) error`

Permanently deletes all submitted data associated with the transaction corresponding to the tokens provided.

##### `UpdateTransaction(tokens *Tokens, options *UpdateTransactionOptions) (*TransactionMetadata, error)`

Changes a transaction's "action", for example upon review in your systems. Returns the updated transaction upon success.

Options:

- `Action: string` - Action taken on the transaction. Typically this will either be "accept" or "reject".
- `Note: string` - An optional note explaining the action taken.

##### `OverrideTransaction(tokens *Tokens, options *OverrideTransactionOptions) error`

Completes a previously created transaction, and overrides its return payload and flags to match the provided parameters.

Parameters:

- `ResponsePayload: string` - A string describing the payload type to return when fetching transaction metadata, e.g. "us-dl". See [our testing guide](https://docs.berbix.com/docs/testing) for possible options.
- `Flags: []string` - An optional list of flags to associate with the transaction (independent of the payload's contents), e.g. ["id_under_18", "id_under_21"]. See [our flags documentation](https://docs.berbix.com/docs/id-flags) for a list of flags.

### `Tokens`

#### Properties

##### `AccessToken: string`

This is the short-lived bearer token that the backend SDK uses to identify requests associated with a given transaction. This is not typically needed when using the higher-level SDK methods.

##### `ClientToken: string`

This is the short-lived token that the frontend SDK uses to identify requests associated with a given transaction. After transaction creation, this will typically be sent to a frontend SDK.

##### `RefreshToken: string`

This is the long-lived token that allows you to create new tokens after the short-lived tokens have expired. This is typically stored in the database associated with the given user session.

##### `TransactionID: int64`

The internal Berbix ID number associated with the transaction.

##### `expiry: time.Time`

The time at which the access and client tokens will expire.

#### Static methods

##### `TokensFromRefresh(refreshToken string) *Tokens`

Creates a tokens object from a refresh token, which can be passed to higher-level SDK methods. The SDK will handle refreshing the tokens for accessing relevant data.
