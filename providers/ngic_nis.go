package providers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/requests"
)

const nisProviderName = "NGIC NIS"

// NISProvider represents a NGIC NIS based Identity Provider
type NISProvider struct {
	*OIDCProvider
}

const nisDefaultScope = "openid email profile"

// newNISProvider initiates a new NISProvider
func NewNISProvider(p *ProviderData, opts options.OIDCOptions) *NISProvider {
	p.setProviderDefaults(providerDefaults{
		name: nisProviderName,
	})

	provider := &NISProvider{
		OIDCProvider: NewOIDCProvider(p, opts),
	}

	return provider
}

var _ Provider = (*NISProvider)(nil)

func (p *NISProvider) GetLoginURL(redirectURI, state, nonce string, extraParams url.Values) string {
	url := p.OIDCProvider.GetLoginURL(redirectURI, state, nonce, extraParams)

	return url
}

func (p *NISProvider) Redeem(ctx context.Context, redirectURL, code, codeVerifier string) (*sessions.SessionState, error) {
	ss, err := p.OIDCProvider.Redeem(ctx, redirectURL, code, codeVerifier)

	return ss, err
}

func (p *NISProvider) EnrichSession(ctx context.Context, s *sessions.SessionState) error {
	userinfo, err := p.getUserInfo(ctx, s)
	if err != nil {
		return fmt.Errorf("failed to retrive user info: %v", err)
	}

	if userinfo.Name != "" {
		s.User = userinfo.Name
	}

	if userinfo.Email != "" {
		s.Email = userinfo.Email
	}

	if len(userinfo.Roles) > 0 {
		s.Groups = userinfo.Roles
	}

	return nil
}

type nisUserinfo struct {
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Roles          []string `json:"groups"`
	EmployeeNumber string   `json:"employee_num"`
	SAMAccountName string   `json:"SAMAccountName"`
	GivenName      string   `json:"GivenName"`
	Surname        string   `json:"Surname"`
	Sub            string   `json:"sub"`
}

func (p *NISProvider) getUserInfo(ctx context.Context, s *sessions.SessionState) (*nisUserinfo, error) {
	var userinfo nisUserinfo
	err := requests.New(p.ProfileURL.String()).
		WithContext(ctx).
		WithMethod("GET").
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+s.AccessToken).
		Do().
		UnmarshalInto(&userinfo)

	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	return &userinfo, nil
}

func (p *NISProvider) ValidateSession(ctx context.Context, s *sessions.SessionState) bool {
	valid := p.OIDCProvider.ValidateSession(ctx, s)

	return valid
}

func (p *NISProvider) RefreshSession(ctx context.Context, s *sessions.SessionState) (bool, error) {
	refreshed, err := p.OIDCProvider.RefreshSession(ctx, s)

	// Refresh could have failed or there was not session to refresh (with no error raised)
	if err != nil || !refreshed {
		return refreshed, err
	}

	return true, nil
}

func (p *NISProvider) CreateSessionFromToken(ctx context.Context, token string) (*sessions.SessionState, error) {
	ss, err := p.OIDCProvider.CreateSessionFromToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("could not create session from token: %v", err)
	}

	return ss, nil
}
