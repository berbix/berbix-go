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

Or, if you need the hosted URL

	options := &CreateHostedTransactionOptions{
		CreateTransactionOptions: CreateTransactionOptions{
			CustomerUID: "internal_customer_uid",
			TemplateKey: "your_template_key",
		},
		HostedOptions: HostedOptions{
			// Optional
			CompletionEmail: "example@example.com",
		},
	}
	resp, err := client.CreateHostedTransaction(options)
	if err != nil {
		// Handle error
	}

	hostedURL := resp.HostedURL

### Create tokens from refresh token

    refreshToken := "" # fetched from database
    transactionTokens := TokensFromRefresh(refreshToken)

### Fetch transaction data

    transactionData, err := client.FetchTransaction(transactionTokens)

## Reference

### `Client`

##### NewClient(secret string, options \*ClientOptions) Client

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

##### `CreateHostedTransaction(options *CreateHostedTransactionOptions) (*CreateHostedTransactionResponse, error)`

Behaves the same as `CreateTransaction()` with two key differences: it returns a URL for a hosted transaction
in addition to tokens and supports two optional parameters in addition to those supported for
`CreateTransaction()`:

- `CompletionEmail` - Where to send an email when the verification completes.
  
- `RedirectURL` - URL to redirect the user to after they complete the transaction. If not specified, the URL specified in the Berbix dashboard will be used instead.

##### `CreateAPIOnlyTransaction(options *CreateAPIOnlyTransactionOptions) (*CreateAPIOnlyTransactionResponse, error)`

Behaves similarly to `CreateTransaction()`, but creates a transaction for which images can be directly uploaded to the Berbix API via `UploadImages()`.
The tokens returned cannot be used to instantiate a Berbix Verify client SDK. `CreateAPIOnlyTransactionOptions` Must be set to a non-nil value.

The `APIOnlyOptions.IDType` property can optionally be set to a value representing the type of ID that will be uploaded
if the ID type is known in advance.  
See the descriptions for properties of `api_only_options` in the ["Create transaction" documentation](https://docs.berbix.com/reference/createtransaction) for a list of acceptable ID types.

The `APIOnlyOptions.IDCountry` property can optionally be set to a two-letter country code if the country that issued
the ID is known in advance.

Setting the country code and/or ID type can improve the accuracy of results in some cases.

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
- `OverrideFields: map[string]string` - An optional mapping from a [transaction field](https://docs.berbix.com/reference#gettransactionmetadata) to the desired override value, e.g. `OverrideFields = map[string]string{"date_of_birth" : "2000-12-09",}`

##### `UploadImages(tokens *Tokens, options *UploadImagesOptions) (*ImageUploadResult, error)`
Upload an image for an API-only transaction.

The `tokens` and `options` properties are required.

We recommend reading the [API-Only Integration Guide](https://docs.berbix.com/docs/api-only-integration-guide) to
understand how to set up an API-only integration. At a high level, images of various subjects should be uploaded in an
order dictated by the API, where the `NextStep` property in the returned `*ImageUploadResult` describes which image
should be uploaded next, or that no more images are expected (as indicated by `NextStepDone`).

The `Issues` property of `ImageUploadResult` specifies feedback on the image, such as whether the text was readable.
This can be useful for coaching end users on how to re-take an image if the `NextStep` indicates another image of the
same subject should be uploaded. See the descriptions of the [`Issue` property](#issues-issue) and the
[`IssueDetails` type](#issuedetails-issuedetails) below for more details.

The `Images` property of the `options` must contain at least one image, but more images may be required depending on which step in the verification process you have reached.
Refer to the [API documentation](https://docs.berbix.com/reference/uploadimages) for an up-to-date description of what and how many images are expected at each step.

See the [documentation of the `RawImage` type](#rawimage) below for more details on the values that should be passed in the `Images` slice.
See the [documentation for the corresponding API endpoint](https://docs.berbix.com/reference/uploadimages) for a
description of what images are expected in what situations and how to interpret the results of the response.



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

### `RawImage`

#### Properties

##### `Image: []byte`

Bytes representing the image to be uploaded. The image should be in a supported format, such as JPEG or PNG, without any extra encoding (such as hex or base 64) applied.
An updated list of supported formats is maintained in the [integration guide](https://docs.berbix.com/docs/api-only-integration-guide#uploading-photos).

##### `Subject: ImageSubject`

Value representing the subject of the image, such as the front of an ID document. The following `ImageSubject` constants
are provided as a convenience:
```go
const (
	ImageSubjectDocumentFront ImageSubject = "document_front"
	ImageSubjectDocumentBack  ImageSubject = "document_back"
	ImageSubjectBarcode       ImageSubject = "document_barcode"
	ImageSubjectSelfieFront   ImageSubject = "selfie_front"
	ImageSubjectSelfieLeft    ImageSubject = "selfie_left"
	ImageSubjectSelfieRight   ImageSubject = "selfie_right"
)
```


##### `ImageFormage: ImageFormat`

A value representing the format of an image. The following constants of type `ImageFormat` are provided as a convenience:
- `ImageFormatJPEG`
- `ImageFormatPNG`

### `ImageUploadResult`

#### Properties

##### `NextStep: NextStep`

Describes the next expected interaction with the SDK. The following constants for `NextStep` values are exposed.
```go
const (
	NextStepUploadDocumentFront  NextStep = "upload_document_front"
	NextStepUploadDocumentBack   NextStep = "upload_document_back"
	NextStepUploadSelfieBasic    NextStep = "upload_selfie_basic"
	NextStepUploadSelfieLiveness NextStep = "upload_selfie_liveness"
	NextStepDone NextStep = "done"
)
```

The `NextStepUpload*` values indicate that the next expected interaction with the API is to upload more images,
while `NextStepDone` indicates no more uploads are required or expected.

##### `Issues: []Issue`

A slice of values describing the issues, if any, with the upload.
This SDK has constants for the following values which may appear in `Issues`:

```go
const (
	IssueBadUpload                 Issue = "bad_upload"
	IssueTextUnreadable            Issue = "text_unreadable"
	IssueNoFaceOnIDDetected        Issue = "no_face_on_id_detected"
	IssueIncompleteBarcodeDetected Issue = "incomplete_barcode_detected"
	IssueUnsupportedIDType         Issue = "unsupported_id_type"
	IssueBadSelfie                 Issue = "bad_selfie"
)
```

`IssueBadUpload` is a catch-all value that is used when the issue with the image doesn't fall into other
  categories and/or Berbix is obfuscating the problem with the image as it may relate to fraud.

`IssueIncompleteBarcodeDetected` can indicate that an incomplete barcode was in the uploaded image, or that the
  barcode was entirely missing.


##### `IssueDetails: IssueDetails`

Additional details related to the issues identified by `Issues`. See [`IssueDetails` below](#issuedetails).

### `IssueDetails`

##### `UnsupportedIDType: *UnsupportedIDTypeFeedback`

This property may be set to a non-`nil` value of the `IssueUnsupportedIDType` value is present in the `Issues` slice
in `ImageUplaodResult`. The `VisaPageOfPassport` property of `UnsupportedIDTypeFeedback` will be set to `true` if it
appears as if the visa page of a passport was uploaded, rather than the photo ID page.

#### Static methods

##### `TokensFromRefresh(refreshToken string) *Tokens`

Creates a tokens object from a refresh token, which can be passed to higher-level SDK methods. The SDK will handle refreshing the tokens for accessing relevant data.
