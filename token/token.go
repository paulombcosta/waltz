package token

import (
	"net/http"

	"github.com/paulombcosta/waltz/session"
	"golang.org/x/oauth2"
)

func New(provider string, r *http.Request, w http.ResponseWriter, session session.SessionManager) CookieStoreTokenProvider {
	return CookieStoreTokenProvider{
		Provider: provider,
		Req:      r,
		Writer:   w,
		Session:  session,
	}
}

type CookieStoreTokenProvider struct {
	Provider string
	Req      *http.Request
	Writer   http.ResponseWriter
	Session  session.SessionManager
}

func (t CookieStoreTokenProvider) GetToken() (*oauth2.Token, error) {
	tokens, err := t.Session.GetSessionTokens(t.Provider, t.Req)
	if err != nil {
		return nil, err
	}
	if tokens == nil {
		return nil, nil
	}
	return tokens, nil
}

func (t CookieStoreTokenProvider) RefreshToken() (*oauth2.Token, error) {
	return t.Session.RefreshToken(t.Provider, t.Req, t.Writer)
}
