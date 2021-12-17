package berbix

type CreateHostedTransactionOptions struct {
	CreateTransactionOptions
	HostedOptions HostedOptions `json:"hosted_options"`
}

type hostedTransactionResponse struct {
	tokenResponse
	HostedURL string `json:"hosted_url"`
}

type CreateHostedTransactionResponse struct {
	Tokens    Tokens
	HostedURL string
}

type HostedOptions struct {
	CompletionEmail string `json:"completion_email"`
	RedirectURL     string `json:"redirect_url"`
}

type CreateAPIOnlyTransactionOptions struct {
	CreateTransactionOptions
	APIOnlyOptions APIOnlyOptions `json:"api_only_options"`
}

type CreateAPIOnlyTransactionResponse struct {
	// embed Tokens to make it possible to add properties later
	Tokens Tokens
}

type APIOnlyOptions struct {
	IDType    string `json:"id_type"`
	IDCountry string `json:"id_country"`
}
