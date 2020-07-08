package google

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func Backend() *backend {
	gp := &googleProvider{}
	b := &backend{
		user:   gp,
	}
	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),

		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
				"cli_code_url",
				"web_code_url",
			},
			LocalStorage: []string{
				framework.WALPrefix,
			},
		},
		Paths: framework.PathAppend(
			b.loginPath(),
			b.cliCodeURLPath(),
			b.webCodeURLPath(),
			b.pathConfig(),
			b.pathUsers(),
			b.pathUsersList(),
		),
		BackendType: logical.TypeCredential,
	}
	return b
}

type backend struct {
	*framework.Backend
	sync.RWMutex
	user   UserProvider
}

const backendHelp = `
The google backend supports authenticating with google oauth2

After mounting, configure it using the "auth/google/config" path.
`
