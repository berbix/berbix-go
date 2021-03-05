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
