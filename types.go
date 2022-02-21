package berbix

import (
	"time"
)

type CreateTransactionOptions struct {
	CustomerUID                          string `json:"customer_uid"`
	TemplateKey                          string `json:"template_key"`
	Email                                string `json:"email"`
	Phone                                string `json:"phone"`
	ConsentsToAutomatedFacialRecognition string `json:"consents_to_automated_facial_recognition"`
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
