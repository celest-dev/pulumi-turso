package provider

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type DatabaseToken struct{}

type DatabaseTokenArgs struct {
	Database      string                      `pulumi:"database"`
	Authorization *DatabaseTokenAuthorization `pulumi:"authorization,optional"`
	ReadAttach    []string                    `pulumi:"readAttach,optional"`
	Expiration    *string                     `pulumi:"expiration,optional"`
}

type DatabaseTokenAuthorization tursoclient.CreateDatabaseTokenAuthorization

var _ infer.Enum[DatabaseTokenAuthorization] = (*DatabaseTokenAuthorization)(nil)

func (*DatabaseTokenAuthorization) Values() []infer.EnumValue[DatabaseTokenAuthorization] {
	return []infer.EnumValue[DatabaseTokenAuthorization]{
		{Value: DatabaseTokenAuthorization(tursoclient.CreateDatabaseTokenAuthorizationFullAccess), Name: "Full Access", Description: "Full access to the database"},
		{Value: DatabaseTokenAuthorization(tursoclient.CreateDatabaseTokenAuthorizationReadOnly), Name: "Read Only", Description: "Read only access to the database"},
	}
}

type DatabaseTokenState struct {
	DatabaseTokenArgs
	Token     string `pulumi:"token" json:"jwt" provider:"secret"`
	ExpiresAt string `pulumi:"expiresAt,optional" json:"expiresAt,omitempty"`
}

var (
	_ infer.CustomCreate[DatabaseTokenArgs, DatabaseTokenState] = (*DatabaseToken)(nil)
	_ infer.CustomRead[DatabaseTokenArgs, DatabaseTokenState]   = (*DatabaseToken)(nil)
	_ infer.CustomDiff[DatabaseTokenArgs, DatabaseTokenState]   = (*DatabaseToken)(nil)
)

func (*DatabaseToken) Create(ctx context.Context, req infer.CreateRequest[DatabaseTokenArgs]) (infer.CreateResponse[DatabaseTokenState], error) {
	input := req.Inputs
	preview := req.DryRun
	if preview {
		return infer.CreateResponse[DatabaseTokenState]{
			ID: req.Name,
			Output: DatabaseTokenState{
				DatabaseTokenArgs: input,
			},
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var expiration tursoclient.OptString
	var expiresAt string
	if input.Expiration != nil {
		expirationDuration, err := time.ParseDuration(*input.Expiration)
		if err != nil {
			return infer.CreateResponse[DatabaseTokenState]{}, fmt.Errorf("error parsing expiration duration: %w", err)
		}
		expiration = tursoclient.NewOptString(expirationDuration.String())
		expiresAt = time.Now().Add(expirationDuration).Format(time.RFC3339)
	}
	var authorization tursoclient.OptCreateDatabaseTokenAuthorization
	if input.Authorization != nil {
		authorization = tursoclient.NewOptCreateDatabaseTokenAuthorization(tursoclient.CreateDatabaseTokenAuthorization(*input.Authorization))
	}

	token, err := client.CreateDatabaseToken(ctx,
		tursoclient.NewOptCreateTokenInput(tursoclient.CreateTokenInput{
			Permissions: tursoclient.NewOptCreateTokenInputPermissions(tursoclient.CreateTokenInputPermissions{
				ReadAttach: tursoclient.NewOptCreateTokenInputPermissionsReadAttach(tursoclient.CreateTokenInputPermissionsReadAttach{
					Databases: input.ReadAttach,
				}),
			}),
		}),
		tursoclient.CreateDatabaseTokenParams{
			OrganizationSlug: config.OrganizationSlug,
			DatabaseName:     input.Database,
			Expiration:       expiration,
			Authorization:    authorization,
		})
	if err != nil {
		return infer.CreateResponse[DatabaseTokenState]{}, fmt.Errorf("error creating database token: %w", err)
	}
	switch token := token.(type) {
	case *tursoclient.CreateDatabaseTokenOK:
		return infer.CreateResponse[DatabaseTokenState]{
			ID: req.Name,
			Output: DatabaseTokenState{
				DatabaseTokenArgs: input,
				ExpiresAt:         expiresAt,
				Token:             token.Jwt.Value,
			},
		}, nil
	default:
		return infer.CreateResponse[DatabaseTokenState]{}, fmt.Errorf("unexpected response creating database token: %T", token)
	}
}

func (*DatabaseToken) Read(ctx context.Context, req infer.ReadRequest[DatabaseTokenArgs, DatabaseTokenState]) (infer.ReadResponse[DatabaseTokenArgs, DatabaseTokenState], error) {
	return infer.ReadResponse[DatabaseTokenArgs, DatabaseTokenState](req), nil
}

func (*DatabaseToken) Diff(ctx context.Context, req infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]) (infer.DiffResponse, error) {
	olds := req.State
	news := req.Inputs
	diff := map[string]p.PropertyDiff{}
	if olds.Database != news.Database {
		diff["database"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.Authorization != news.Authorization {
		diff["authorization"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if !slices.Equal(olds.ReadAttach, news.ReadAttach) {
		diff["readAttach"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.Expiration != news.Expiration {
		diff["expiration"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.ExpiresAt != "" {
		oldExp, err := time.Parse(time.RFC3339, olds.ExpiresAt)
		if err != nil {
			return infer.DiffResponse{}, fmt.Errorf("error parsing old expiration time: %w", err)
		}
		if time.Now().After(oldExp) {
			diff["expiresAt"] = p.PropertyDiff{Kind: p.UpdateReplace}
		}
	}
	return infer.DiffResponse{
		DeleteBeforeReplace: false,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
}
