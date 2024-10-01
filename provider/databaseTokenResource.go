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
	Token     string     `pulumi:"token" json:"jwt" provider:"secret"`
	ExpiresAt *time.Time `pulumi:"expiresAt,optional" json:"expiresAt"`
}

var (
	_ infer.CustomCreate[DatabaseTokenArgs, DatabaseTokenState] = DatabaseToken{}
	_ infer.CustomDiff[DatabaseTokenArgs, DatabaseTokenState]   = DatabaseToken{}
)

func (DatabaseToken) Create(ctx context.Context, name string, input DatabaseTokenArgs, preview bool) (string, DatabaseTokenState, error) {
	if preview {
		return "", DatabaseTokenState{
			DatabaseTokenArgs: input,
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var expiration tursoclient.OptString
	var expiresAt *time.Time
	if input.Expiration != nil {
		expirationDuration, err := time.ParseDuration(*input.Expiration)
		if err != nil {
			return "", DatabaseTokenState{}, fmt.Errorf("error parsing expiration duration: %w", err)
		}
		expiration = tursoclient.NewOptString(expirationDuration.String())
		exp := time.Now().Add(expirationDuration)
		expiresAt = &exp
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
			OrganizationName: config.OrganizationName,
			DatabaseName:     input.Database,
			Expiration:       expiration,
			Authorization:    authorization,
		})
	if err != nil {
		return "", DatabaseTokenState{}, fmt.Errorf("error creating database token: %w", err)
	}
	switch token := token.(type) {
	case *tursoclient.CreateDatabaseTokenOK:
		return input.Database, DatabaseTokenState{
			DatabaseTokenArgs: input,
			ExpiresAt:         expiresAt,
			Token:             token.Jwt.Value,
		}, nil
	default:
		return "", DatabaseTokenState{}, fmt.Errorf("unexpected response creating database token: %T", token)
	}
}

func (DatabaseToken) Diff(ctx context.Context, id string, olds DatabaseTokenState, news DatabaseTokenArgs) (p.DiffResponse, error) {
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
	if olds.ExpiresAt != nil && olds.ExpiresAt.Before(time.Now()) {
		diff["expiresAt"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	return p.DiffResponse{
		DeleteBeforeReplace: false,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
}
