//Credit to https://github.com/simonswine/vault-plugin-auth-google
package google

import (
	"context"
	"errors"


	"golang.org/x/oauth2"
	goauth "google.golang.org/api/oauth2/v2"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	googleAuthCodeParameterName = "code"
	stateParameterName          = "state"
)

func (b *backend) loginPath() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "login",
			Fields: map[string]*framework.FieldSchema{
				googleAuthCodeParameterName: &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Google authentication code. Required.",
				},
				stateParameterName: {
					Type:        framework.TypeString,
					Description: "State parameter used by web login. If used the web method is used. Optional.",
				},
			},

			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.UpdateOperation:         b.pathLogin,
				logical.AliasLookaheadOperation: b.pathLogin,
			},
		},
	}
}

func (b *backend) pathLogin(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	
	code := data.Get(googleAuthCodeParameterName).(string)

	config, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	authType := typeCLI

	// use web config if state is set
	if stateValue := data.Get(stateParameterName).(string); len(stateValue) > 0 {
		statePath := b.statePath(stateValue)

		state, err := b.state(ctx, req, statePath)
		if err != nil {
			return nil, err
		}

		// no matching state found
		if state == nil {
			return logical.ErrorResponse("this state can't be found or has already been used"), nil
		}

		authType = state.Type

		if err := b.deleteState(ctx, req, statePath); err != nil {
			return nil, err
		}

		if err := b.cleanupStates(ctx, req); err != nil {
			return nil, err
		}
	}
	
	oauth2config := config.oauth2Config(authType)
	
	token, err := b.user.oauth2Exchange(ctx, code, oauth2config)
	if err != nil {
		return nil, err
	}

	user, err := b.authenticate(ctx, config, token, authType)
	if err != nil {
		return nil, err
	}

	if !config.authorised(user) {
		return logical.ErrorResponse("user is not allowed to login"), nil
	}

	encodedToken, err := encodeToken(token)
	if err != nil {
		return nil, err
	}

	ttl, maxTTL := config.ttlForType(authType)

	var policies []string
	p, err := b.User(ctx, req.Storage, user.Email)
	if err != nil {
		return nil, err
	}
	if p == nil {
		policies = []string{"default"}
	} else {
		policies = p.Policies
	}

	resp := &logical.Response{
		Auth: &logical.Auth{
			InternalData: map[string]interface{}{
				"token": encodedToken,
				"type":  authType,
			},
			Policies: policies,
			Metadata: map[string]string{
				"username": user.Email,
				"domain":   user.Hd,
			},
			DisplayName: user.Email,
			LeaseOptions: logical.LeaseOptions{
				TTL:       ttl,
				MaxTTL:    maxTTL,
				Renewable: true,
			},
			Alias: &logical.Alias{
				Name: user.Email,
				Metadata: map[string]string{
					"username":   user.Email,
					"domain":     user.Hd,
					"first_name": user.GivenName,
					"last_name":  user.FamilyName,
				},
			},
		},
	}

	return resp, nil
}

func (b *backend) pathRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	encodedToken, ok := req.Auth.InternalData["token"].(string)
	if !ok {
		return nil, errors.New("no refresh token from previous login")
	}

	config, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	token, err := decodeToken(encodedToken)
	if err != nil {
		return nil, err
	}

	authType, ok := req.Auth.InternalData["type"].(string)

	user, err := b.authenticate(ctx, config, token, authType)
	if err != nil {
		return nil, err
	}

	if !config.authorised(user) {
		return logical.ErrorResponse("user is not allowed to login"), nil
	}

	resp := &logical.Response{Auth: req.Auth}

	return resp, nil
}

func (b *backend) authenticate(ctx context.Context, config *config, token *oauth2.Token, authType string) (*goauth.Userinfo, error) {
	oauth2config := config.oauth2Config(authType)

	user, err := b.user.authUser(ctx, oauth2config, token)
	if err != nil {
		return nil, err
	}

	return user, nil
}
