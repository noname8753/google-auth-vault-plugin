//Credit to https://github.com/simonswine/vault-plugin-auth-google
package google

import (
	"context"

	"golang.org/x/oauth2"
	goauth "google.golang.org/api/oauth2/v2"
)

type googleProvider struct {
}

// UserProvider does the authentication of user with oauth2
type UserProvider interface {
	authUser(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*goauth.Userinfo, error)
	oauth2Exchange(ctx context.Context, code string, config *oauth2.Config) (*oauth2.Token, error)
}

var _ UserProvider = &googleProvider{}


func (p *googleProvider) oauth2Exchange(ctx context.Context, code string, config *oauth2.Config) (*oauth2.Token, error) {
	return config.Exchange(ctx, code)
}

func (p *googleProvider) authUser(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*goauth.Userinfo, error) {
	client := config.Client(ctx, token)
	userService, err := goauth.New(client)
	if err != nil {
		return nil, err
	}

	user, err := goauth.NewUserinfoV2MeService(userService).Get().Do()
	if err != nil {
		return nil, err
	}

	return user, nil
}