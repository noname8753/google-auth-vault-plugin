package google

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathUsersList() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "users/?$",

			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathUserList,
			},
	
			HelpSynopsis:    pathUserReadHelpSyn,
			HelpDescription: pathUserHelpDesc,
			DisplayAttrs: &framework.DisplayAttributes{
				Navigation: true,
			},
		},
	}
}

func (b *backend) pathUsers() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: `users/(?P<name>.+)`,
			Fields: map[string]*framework.FieldSchema{
				"name": &framework.FieldSchema{
					Type:        framework.TypeString,
					Description: "Name of the user.",
				},
				"policies": &framework.FieldSchema{
					Type:        framework.TypeCommaStringSlice,
					Description: "List of policies associated with the user.",
				},
			},
	
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.DeleteOperation: b.pathUserDelete,
				logical.ReadOperation:   b.pathUserRead,
				logical.UpdateOperation: b.pathUserWrite,
			},
	
			HelpSynopsis:    pathUserHelpSyn,
			HelpDescription: pathUserHelpDesc,
			DisplayAttrs: &framework.DisplayAttributes{
				Action:   "Create",
			},
		},
	}
}


func (b *backend) User(ctx context.Context, s logical.Storage, n string) (*UserEntry, error) {
	entry, err := s.Get(ctx, "user/"+n)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result UserEntry
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (b *backend) pathUserDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	if len(name) == 0 {
		return logical.ErrorResponse("Error empty name"), nil
	}

	err := req.Storage.Delete(ctx, "user/"+name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathUserRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	if len(name) == 0 {
		return logical.ErrorResponse("Error empty name"), nil
	}

	user, err := b.User(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"policies": user.Policies,
		},
	}, nil
}

func (b *backend) pathUserWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name := d.Get("name").(string)
	if len(name) == 0 {
		return logical.ErrorResponse("Error empty name"), nil
	}


	policies := d.Get("policies").([]string)

	// Store it
	entry, err := logical.StorageEntryJSON("user/"+name, &UserEntry{
		Policies: policies,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathUserList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	users, err := req.Storage.List(ctx, "user/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(users), nil
}


type UserEntry struct {
	Policies []string
}

const pathUserReadHelpSyn = `
List users 
vault list auth/google/users

vault read auth/google/users/someuser@someemail.com

Key         Value
---         -----
policies    [admin]

`

const pathUserHelpSyn = `
Map username/email to policy.
vault write auth/google/users/someuser@someemail.com policies=default
or multiple:
vault write auth/google/users/someuser@someemail.com policies=default,admin,specialaccess
`

const pathUserHelpDesc = `
This endpoint allows you to create, read, update, and delete configuration

Deleting a user will not revoke their existing auth.

vault list auth/token/accessors


`