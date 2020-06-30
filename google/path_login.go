package google

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/admin/directory/v1"
	goauth "google.golang.org/api/oauth2/v2"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	loginPath                   = "login"
	googleAuthCodeParameterName = "code"
	stateParameterName          = "state"
)

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

	user, groups, err := b.authenticate(ctx, config, token, authType)
	if err != nil {
		return nil, err
	}

	if !config.authorised(user, groups) {
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
    return nil, nil
  }

  policies = p.Policies
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

  //need to map policies here resp.Auth.Policies []string `json:"policies" mapstructure:"policies" structs:"policies"`
	setGroups(resp.Auth, user, groups)

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

	user, groups, err := b.authenticate(ctx, config, token, authType)
	if err != nil {
		return nil, err
	}

	if !config.authorised(user, groups) {
		return logical.ErrorResponse("user is not allowed to login"), nil
	}

	resp := &logical.Response{Auth: req.Auth}

	// Remove old aliases
	resp.Auth.GroupAliases = nil

	setGroups(resp.Auth, user, groups)

	return resp, nil
}

func setGroups(auth *logical.Auth, user *goauth.Userinfo, groups []*admin.Group) {
	// add every associated group
	for _, group := range groups {
		auth.GroupAliases = append(auth.GroupAliases, &logical.Alias{
			Name: group.Email,
			Metadata: map[string]string{
				"name":        group.Name,
				"aliases":     strings.Join(group.Aliases, ","),
				"description": group.Description,
			},
		})
	}

	// add a group alias for it's domain
	auth.GroupAliases = append(auth.GroupAliases, &logical.Alias{
		Name: fmt.Sprintf("@%s", user.Hd),
	})
}

func (b *backend) authenticate(ctx context.Context, config *config, token *oauth2.Token, authType string) (*goauth.Userinfo, []*admin.Group, error) {
	oauth2config := config.oauth2Config(authType)

	user, err := b.user.authUser(ctx, oauth2config, token)
	if err != nil {
		return nil, nil, err
	}

	groups, err := b.groups.groupsPerUser(ctx, config, user.Email)
	if err != nil {
		b.Logger().Warn("querying the admin directory API for the groups for the user failed: ", "user", user, "error", "err")
		groups = []*admin.Group{}
	}

	return user, groups, nil
}
