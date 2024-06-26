package Autolab

import "golang.org/x/oauth2"

// TokenSource implements the oauth2.TokenSource interface
// It is expected that there already exists a Token
// The main purpose is to be used with oauth2.NewClient
type TokenSource struct {
	ts TokenStore
	ac AutolabOAuthClient
}

type TokenStore interface {
	Load() (oauth2.Token, error)
	Save(oauth2.Token) error
}

func NewTokenSource(store TokenStore, ac AutolabOAuthClient) *TokenSource {
	return &TokenSource{
		ts: store,
		ac: ac,
	}
}

// Checks if the token is expired and refreshes it if necessary
// TODO: actually check if the token is expired;
// for now we just refresh every time
// Must return pointer to satisfy the interface
func (a TokenSource) Token() (*oauth2.Token, error) {
	token, err := a.ts.Load()
	if err != nil {
		return nil, err
	}
	newToken, err := a.ac.RefreshToken(token)
	if err != nil {
		return nil, err
	}
	a.ts.Save(*newToken)

	return newToken, nil
}
