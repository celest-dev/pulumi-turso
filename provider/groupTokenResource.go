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

var _ infer.Annotated = (*GroupToken)(nil)

func (t *GroupToken) Annotate(a infer.Annotator) {
	a.Describe(&t, "An authentication token for a Turso group. Tokens can be used to connect to any database in the group with specific permissions.")
}

type GroupTokenArgs struct {
	Group         string                   `pulumi:"group"`
	Authorization *GroupTokenAuthorization `pulumi:"authorization,optional"`
	ReadAttach    []string                 `pulumi:"readAttach,optional"`
	Expiration    *string                  `pulumi:"expiration,optional"`
}

var _ infer.Annotated = (*GroupTokenArgs)(nil)

func (args *GroupTokenArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.Group, "The name of the group to create the token for.")
	a.Describe(&args.Authorization, "The authorization level for the token. Defaults to full-access if not specified.")
	a.Describe(&args.ReadAttach, "List of databases that can be attached for read operations using this token.")
	a.Describe(&args.Expiration, "The duration until the token expires (e.g., '24h', '7d', '30d'). If not specified, the token will not expire.")
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

var _ infer.Annotated = (*GroupTokenState)(nil)

func (state *GroupTokenState) Annotate(a infer.Annotator) {
	a.Describe(&state.Token, "The JWT token used to authenticate with databases in the group.")
	a.Describe(&state.ExpiresAt, "The RFC3339 timestamp when the token expires, if an expiration was set.")
}

var (
	_ infer.CustomCreate[GroupTokenArgs, GroupTokenState] = (*GroupToken)(nil)
	_ infer.CustomRead[GroupTokenArgs, GroupTokenState]   = (*GroupToken)(nil)
	_ infer.CustomDiff[GroupTokenArgs, GroupTokenState]   = (*GroupToken)(nil)
)

func (*GroupToken) Create(ctx context.Context, req infer.CreateRequest[GroupTokenArgs]) (infer.CreateResponse[GroupTokenState], error) {
	input := req.Inputs
	preview := req.DryRun
	if preview {
		return infer.CreateResponse[GroupTokenState]{
			ID: req.Name,
			Output: GroupTokenState{
				GroupTokenArgs: input,
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
			return infer.CreateResponse[GroupTokenState]{}, fmt.Errorf("error parsing expiration duration: %w", err)
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
			OrganizationSlug: config.OrganizationSlug,
			GroupName:        input.Group,
			Expiration:       expiration,
			Authorization:    authorization,
		})
	if err != nil {
		return infer.CreateResponse[GroupTokenState]{}, fmt.Errorf("error creating group token: %w", err)
	}
	switch token := token.(type) {
	case *tursoclient.CreateGroupTokenOK:
		return infer.CreateResponse[GroupTokenState]{
			ID: req.Name,
			Output: GroupTokenState{
				GroupTokenArgs: input,
				ExpiresAt:      expiresAt,
				Token:          token.Jwt.Value,
			},
		}, nil
	default:
		return infer.CreateResponse[GroupTokenState]{}, fmt.Errorf("unexpected response creating group token: %T", token)
	}
}

func (*GroupToken) Read(ctx context.Context, req infer.ReadRequest[GroupTokenArgs, GroupTokenState]) (infer.ReadResponse[GroupTokenArgs, GroupTokenState], error) {
	return infer.ReadResponse[GroupTokenArgs, GroupTokenState](req), nil
}

func (*GroupToken) Diff(ctx context.Context, req infer.DiffRequest[GroupTokenArgs, GroupTokenState]) (infer.DiffResponse, error) {
	olds := req.State
	news := req.Inputs
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
