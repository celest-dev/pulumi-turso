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

type GroupToken struct{}

type GroupTokenArgs struct {
	Group         string                   `pulumi:"database"`
	Authorization *GroupTokenAuthorization `pulumi:"authorization,optional"`
	ReadAttach    []string                 `pulumi:"readAttach,optional"`
	Expiration    *string                  `pulumi:"expiration,optional"`
}

type GroupTokenAuthorization tursoclient.CreateGroupTokenAuthorization

var _ infer.Enum[GroupTokenAuthorization] = (*GroupTokenAuthorization)(nil)

func (*GroupTokenAuthorization) Values() []infer.EnumValue[GroupTokenAuthorization] {
	return []infer.EnumValue[GroupTokenAuthorization]{
		{Value: GroupTokenAuthorization(tursoclient.CreateGroupTokenAuthorizationFullAccess), Name: "Full Access", Description: "Full access to the database"},
		{Value: GroupTokenAuthorization(tursoclient.CreateGroupTokenAuthorizationReadOnly), Name: "Read Only", Description: "Read only access to the database"},
	}
}

type GroupTokenState struct {
	GroupTokenArgs
	Token     string `pulumi:"token" json:"jwt" provider:"secret"`
	ExpiresAt string `pulumi:"expiresAt,optional" json:"expiresAt,omitempty"`
}

var (
	_ infer.CustomCreate[GroupTokenArgs, GroupTokenState] = GroupToken{}
	_ infer.CustomRead[GroupTokenArgs, GroupTokenState]   = GroupToken{}
	_ infer.CustomDiff[GroupTokenArgs, GroupTokenState]   = GroupToken{}
)

func (GroupToken) Create(ctx context.Context, name string, input GroupTokenArgs, preview bool) (string, GroupTokenState, error) {
	if preview {
		return "", GroupTokenState{
			GroupTokenArgs: input,
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var expiration tursoclient.OptString
	var expiresAt string
	if input.Expiration != nil {
		expirationDuration, err := time.ParseDuration(*input.Expiration)
		if err != nil {
			return "", GroupTokenState{}, fmt.Errorf("error parsing expiration duration: %w", err)
		}
		expiration = tursoclient.NewOptString(expirationDuration.String())
		expiresAt = time.Now().Add(expirationDuration).Format(time.RFC3339)
	}
	var authorization tursoclient.OptCreateGroupTokenAuthorization
	if input.Authorization != nil {
		authorization = tursoclient.NewOptCreateGroupTokenAuthorization(tursoclient.CreateGroupTokenAuthorization(*input.Authorization))
	}

	token, err := client.CreateGroupToken(ctx,
		tursoclient.NewOptCreateTokenInput(tursoclient.CreateTokenInput{
			Permissions: tursoclient.NewOptCreateTokenInputPermissions(tursoclient.CreateTokenInputPermissions{
				ReadAttach: tursoclient.NewOptCreateTokenInputPermissionsReadAttach(tursoclient.CreateTokenInputPermissionsReadAttach{
					Databases: input.ReadAttach,
				}),
			}),
		}),
		tursoclient.CreateGroupTokenParams{
			OrganizationName: config.OrganizationName,
			GroupName:        input.Group,
			Expiration:       expiration,
			Authorization:    authorization,
		})
	if err != nil {
		return "", GroupTokenState{}, fmt.Errorf("error creating group token: %w", err)
	}
	switch token := token.(type) {
	case *tursoclient.CreateGroupTokenOK:
		return input.Group, GroupTokenState{
			GroupTokenArgs: input,
			ExpiresAt:      expiresAt,
			Token:          token.Jwt.Value,
		}, nil
	default:
		return "", GroupTokenState{}, fmt.Errorf("unexpected response creating group token: %T", token)
	}
}

func (GroupToken) Read(ctx context.Context, id string, inputs GroupTokenArgs, state GroupTokenState) (canonicalID string, normalizedInputs GroupTokenArgs, normalizedState GroupTokenState, err error) {
	return id, inputs, state, nil
}

func (GroupToken) Diff(ctx context.Context, id string, olds GroupTokenState, news GroupTokenArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if olds.Group != news.Group {
		diff["group"] = p.PropertyDiff{Kind: p.UpdateReplace}
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
			return p.DiffResponse{}, fmt.Errorf("error parsing old expiration time: %w", err)
		}
		if time.Now().After(oldExp) {
			diff["expiresAt"] = p.PropertyDiff{Kind: p.UpdateReplace}
		}
	}
	return p.DiffResponse{
		DeleteBeforeReplace: false,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
}
