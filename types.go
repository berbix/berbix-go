package berbix

import (
	"fmt"
	"time"
)

type CreateTransactionOptions struct {
	CustomerUID string `json:"customer_uid"`
	TemplateKey string `json:"template_key"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type UpdateTransactionOptions struct {
	Action string `json:"action"`
	Note   string `json:"note"`
}

type OverrideTransactionOptions struct {
	ResponsePayload string            `json:"response_payload"`
	Flags           []string          `json:"flags"`
	OverrideFields  map[string]string `json:"override_fields"`
}

type DuplicateInfo struct {
	CustomerUID string `json:"customer_uid"`

	TransactionID int64 `json:"transaction_id"`
}

type TransactionImages struct {
	// Full image captured from the user. This will be present for front, selfie and liveness.
	FullImage string `json:"full_image,omitempty"`

	// Cropped ID image captured from the user. This will be present for front and back.
	CroppedImage string `json:"cropped_image,omitempty"`

	// Cropped face image captured from the user. This will be present for front, selfie and liveness.
	FaceImage string `json:"face_image,omitempty"`
}

type TransactionImageSet struct {
	Front    *TransactionImages `json:"front,omitempty"`
	Back     *TransactionImages `json:"back,omitempty"`
	Selfie   *TransactionImages `json:"selfie,omitempty"`
	Liveness *TransactionImages `json:"liveness,omitempty"`
}

type TransactionSource struct {
	// The value of the field as determined by this source
	Value string `json:"value"`

	// The type of source
	Type string `json:"type"`

	// The confidence level for this source
	Confidence string `json:"confidence"`
}

type TransactionField struct {
	// The highest confidence value for this field
	Value string `json:"value"`

	// The confidence level for this field
	Confidence string `json:"confidence"`

	// The underlying sources of the data for this field
	Sources []TransactionSource `json:"sources"`
}

type TransactionFieldSet struct {
	// The given name of the person completing the flow.
	GivenName *TransactionField `json:"given_name,omitempty"`

	// The middle name of the person completing the flow.
	MiddleName *TransactionField `json:"middle_name,omitempty"`

	// The family name of the person completing the flow.
	FamilyName *TransactionField `json:"family_name,omitempty"`

	// The date of birth of the person completing the flow.
	DateOfBirth *TransactionField `json:"date_of_birth,omitempty"`

	// The sex of the person completing the flow. Available upon request if required for your use case.
	Sex *TransactionField `json:"sex,omitempty"`

	// The age of the person completing the flow
	Age *TransactionField `json:"age,omitempty"`

	// The nationality of the person completing the flow
	Nationality *TransactionField `json:"nationality,omitempty"`

	// The expiry date of the ID collected in the flow
	IDExpiryDate *TransactionField `json:"id_expiry_date,omitempty"`

	// The issue date of the ID collected in the flow
	IDIssueDate *TransactionField `json:"id_issue_date,omitempty"`

	// The ID number of the ID collected in the flow
	IDNumber *TransactionField `json:"id_number,omitempty"`

	// The type of the ID collected in the flow
	IDType *TransactionField `json:"id_type,omitempty"`

	// The issuer of the ID collected in the flow
	IDIssuer *TransactionField `json:"id_issuer,omitempty"`

	// The email address as verified in the flow
	EmailAddress *TransactionField `json:"email_address,omitempty"`

	// The phone number as verified in the flow
	PhoneNumber *TransactionField `json:"phone_number,omitempty"`

	// The street address collected in the flow
	AddressStreet *TransactionField `json:"address_street,omitempty"`

	// The city of the address collected in the flow
	AddressCity *TransactionField `json:"address_city,omitempty"`

	// The subdivision of the address collected in the flow
	AddressSubdivision *TransactionField `json:"address_subdivision,omitempty"`

	// The postal code of the address collected in the flow
	AddressPostalCode *TransactionField `json:"address_postal_code,omitempty"`

	// The country of the address collected in the flow
	AddressCountry *TransactionField `json:"address_country,omitempty"`

	// The unit of the address collected in the flow
	AddressUnit *TransactionField `json:"address_unit,omitempty"`
}

type TransactionMetadata struct {
	// String representing the entity's type.
	Entity string `json:"entity"`

	// Berbix Transaction ID represented by the associated metadata.
	ID int64 `json:"id"`

	// Any flags associated with the verifications for this transaction.
	Flags []string `json:"flags"`

	// The action as configured in the customer dashboard for the given verification state.
	Action string `json:"action,omitempty"`

	// Data field values extracted from the verification sets.
	Fields *TransactionFieldSet `json:"fields,omitempty"`

	// Short-lived URLs of images collected from the end user.
	Images *TransactionImageSet `json:"images,omitempty"`

	// When the transaction was originally created.
	CreatedAt time.Time `json:"created_at"`

	// The user's unique identifier in your systems as provided in transaction creation.
	CustomerUID string `json:"customer_uid"`

	// A list of CustomerUIDs and Berbix Transaction IDs associated with those duplicates if duplicates of the photo ID are identified for the given transaction.
	Duplicates []DuplicateInfo `json:"duplicates"`

	// The link to Berbix dashboard page for this transaction.
	DashboardURL string `json:"dashboard_url,omitempty"`

	// Optional information about the response. Used in test mode only.
	ImplementationInfo string `json:"implementation_info,omitempty"`
}

type (
	ImageSubject string
	ImageFormat  string
)

const (
	ImageSubjectDocumentFront ImageSubject = "document_front"
	ImageSubjectDocumentBack  ImageSubject = "document_back"
	ImageSubjectBarcode       ImageSubject = "document_barcode"
	ImageSubjectSelfieFront   ImageSubject = "selfie_front"
	ImageSubjectSelfieLeft    ImageSubject = "selfie_left"
	ImageSubjectSelfieRight   ImageSubject = "selfie_right"

	ImageFormatPNG  ImageFormat = "image/png"
	ImageFormatJPEG ImageFormat = "image/jpeg"
)

type ImageData struct {
	ImageSubject ImageSubject `json:"image_subject"`
	Format       ImageFormat  `json:"format"`

	// The base64-encoded image data
	Data string `json:"data"`
}

type ImageUploadRequest struct {
	Images []ImageData `json:"images"`
}

type NextStep string

type ImageUploadResponse struct {
	UploadFeedback *UploadFeedback `json:"upload_feedback"`
	NextStep       NextStep        `json:"next_step"`
}

type UploadFeedback struct {
	// Catch-all property to make it easy to check if there was a problem with the uploaded image in a forward-compatible, way even as new properties are added to the `upload_feedback` object.
	BadUpload bool `json:"bad_upload"`

	// If present, indicates that the text on the uploaded ID was unreadable.
	TextUnreadable *struct{} `json:"text_unreadable"`

	// If present, indicates that no face was detected in the image of an ID that we expect to have a portrait.
	NoFaceOnIDDetected *struct{} `json:"no_face_on_id_detected"`

	// If present, indicates that we were not able to detect a complete barcode in the upload, potentially due to part of the barcode being out of frame.
	IncompleteBarcodeDetected *struct{} `json:"incomplete_barcode_detected"`

	// If present, indicates than an unsupported ID type, such as a military ID or the visa page of a passport, was uploaded.
	UnsupportedIDType *UnsupportedIDTypeFeedback `json:"unsupported_id_type"`

	// Catch-all error indicating that there was an issue validating the provided selfie image, potentially because the image is not of someone's face.
	BadSelfie *struct{} `json:"bad_selfie"`
}

type UnsupportedIDTypeFeedback struct {
	// Indicates that the visa page of a passport was uploaded, rather than the ID page.
	VisaPageOfPassport *bool `json:"visa_page_of_passport,omitempty"`
}

const (
	NextStepUploadDocumentFront  NextStep = "upload_document_front"
	NextStepUploadDocumentBack   NextStep = "upload_document_back"
	NextStepUploadSelfieBasic    NextStep = "upload_selfie_basic"
	NextStepUploadSelfieLiveness NextStep = "upload_selfie_liveness"
	// NextStepDone indicates that no more uploads are expected.
	NextStepDone NextStep = "done"
)

type ImageUploadResult struct {
	// The NextStep of ImageUploadResponse should be inspected to determine the next step the
	// API expects, e.g. uploading another image.
	// Similarly, the UploadFeedback will indicate legibility issues or other information that
	// could immediately be determined from the upload.
	ImageUploadResponse
}

type InvalidUploadForStateResponse struct {
	// The next step to take given the current state.
	NextStep string `json:"next_step"`
	// A message describing the error to aid debugging
	Message string `json:"message"`
}

type InvalidStateErr struct {
	InvalidUploadForStateResponse
}

func (i InvalidStateErr) Error() string {
	return fmt.Sprintf(
		"Invalid API interaction for the current state of the tranasction. Next expected step: %q. Message: %q.",
		i.NextStep, i.Message)
}

type errorMessage struct {
	// This is exported so that it can be accessed outside the package when
	// errorMessage is embedded in other structs
	Message string
}

func (e errorMessage) Error() string {
	return e.Message
}

type TransactionDoesNotExistErr struct {
	errorMessage
}

type PayloadTooLargeErr struct {
	errorMessage
}

type GenericErrorResponse struct {
	StatusCode int    `json:"code"`
	Readable   string `json:"readable,omitempty"`
	Message    string `json:"message"`
}

func (g *GenericErrorResponse) Error() string {
	return fmt.Sprintf("Got response code %d with message %q", g.StatusCode, g.Message)
}
