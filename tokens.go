package berbix

import "time"

type tokenResponse struct {
	TransactionID int64 `json:"transaction_id"`
	RefreshToken string `json:"refresh_token"`
	AccessToken string `json:"access_token"`
	ClientToken string `json:"client_token"`
	ExpiresIn int `json:"expires_in"`
}

type Tokens struct {
	TransactionID int64
	AccessToken string
	RefreshToken string
	ClientToken string
	Expiry time.Time
}

func (t *Tokens) NeedsRefresh() bool {
	return t.AccessToken == "" || t.Expiry.IsZero() || t.Expiry.Before(time.Now())
}

func TokensFromRefresh(refreshToken string) *Tokens {
	return &Tokens{
		RefreshToken: refreshToken,
	}
}

func fromTokenResponse(response *tokenResponse) *Tokens {
	return &Tokens{
		TransactionID: response.TransactionID,
		AccessToken:   response.AccessToken,
		RefreshToken:  response.RefreshToken,
		ClientToken:   response.ClientToken,
		Expiry:        time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
	}
}