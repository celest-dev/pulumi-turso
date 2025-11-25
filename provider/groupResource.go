package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Group struct{}

type GroupArgs struct {
	Name             string   `pulumi:"name"`
	PrimaryLocation  string   `pulumi:"primaryLocation"`
	ReplicaLocations []string `pulumi:"replicaLocations,optional"`
	Extensions       []string `pulumi:"extensions,optional"`
}

type GroupState struct {
	DeleteProtection bool     `pulumi:"deleteProtection" json:"delete_protection"`
	Locations        []string `pulumi:"locations" json:"locations"`
	Name             string   `pulumi:"name" json:"name"`
	Primary          string   `pulumi:"primary" json:"primary"`
	UUID             string   `pulumi:"uuid" json:"uuid"`
}

var (
	_ infer.CustomCreate[GroupArgs, GroupState] = (*Group)(nil)
	_ infer.CustomRead[GroupArgs, GroupState]   = (*Group)(nil)
	_ infer.CustomUpdate[GroupArgs, GroupState] = (*Group)(nil)
	_ infer.CustomDelete[GroupState]            = (*Group)(nil)
	_ infer.CustomDiff[GroupArgs, GroupState]   = (*Group)(nil)
)

func (*Group) Create(ctx context.Context, req infer.CreateRequest[GroupArgs]) (infer.CreateResponse[GroupState], error) {
	input := req.Inputs
	preview := req.DryRun
	p.GetLogger(ctx).Infof("creating group %s (preview=%v)", req.Name, preview)

	if preview {
		return infer.CreateResponse[GroupState]{
			ID: input.Name,
			Output: GroupState{
				Name: input.Name,
			},
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var extension tursoclient.OptExtensions
	if len(input.Extensions) == 1 && input.Extensions[0] == "all" {
		extension = tursoclient.NewOptExtensions(tursoclient.Extensions{
			Type:        tursoclient.Extensions0Extensions,
			Extensions0: tursoclient.Extensions0All,
		})
	} else if len(input.Extensions) > 0 {
		enabled := make([]tursoclient.Extensions1Item, len(input.Extensions))
		for i, ext := range input.Extensions {
			enabled[i] = tursoclient.Extensions1Item(ext)
		}
		extension = tursoclient.NewOptExtensions(tursoclient.Extensions{
			Type:                 tursoclient.Extensions1ItemArrayExtensions,
			Extensions1ItemArray: enabled,
		})
	}
	createReq := tursoclient.NewGroup{
		Name:       input.Name,
		Location:   input.PrimaryLocation,
		Extensions: extension,
	}
	res, err := client.CreateGroup(ctx, &createReq, tursoclient.CreateGroupParams{
		OrganizationSlug: config.OrganizationSlug,
	})
	if err != nil {
		return infer.CreateResponse[GroupState]{}, fmt.Errorf("failed to create group: %w\n%v", err, res)
	}
	_, ok := res.(*tursoclient.CreateGroupOK)
	if !ok {
		return infer.CreateResponse[GroupState]{}, fmt.Errorf("failed to create group: unexpected response from server (%T): %v", res, res)
	}

	for _, location := range input.ReplicaLocations {
		if location == input.PrimaryLocation {
			continue
		}
		res, err := client.AddLocationToGroup(ctx, tursoclient.AddLocationToGroupParams{
			OrganizationSlug: config.OrganizationSlug,
			GroupName:        input.Name,
			Location:         location,
		})
		if err != nil {
			return infer.CreateResponse[GroupState]{}, fmt.Errorf("failed to add location to group: %w", err)
		}
		if _, ok := res.(*tursoclient.AddLocationToGroupOK); !ok {
			return infer.CreateResponse[GroupState]{}, fmt.Errorf("unexpected response from server (%T): %v", res, res)
		}
	}

	state, err := config.readGroupResource(ctx, input.Name)
	if err != nil {
		return infer.CreateResponse[GroupState]{}, fmt.Errorf("failed to read group: %w", err)
	}

	return infer.CreateResponse[GroupState]{ID: state.Name, Output: state}, nil
}

func (*Group) Read(ctx context.Context, req infer.ReadRequest[GroupArgs, GroupState]) (infer.ReadResponse[GroupArgs, GroupState], error) {
	p.GetLogger(ctx).Infof("reading group %s", req.ID)

	config := infer.GetConfig[Config](ctx)
	normalizedState, err := config.readGroupResource(ctx, req.ID)
	if err != nil {
		return infer.ReadResponse[GroupArgs, GroupState]{}, fmt.Errorf("failed to read group: %w", err)
	}

	return infer.ReadResponse[GroupArgs, GroupState]{
		ID:     req.ID,
		Inputs: req.Inputs,
		State:  normalizedState,
	}, nil
}

func (*Group) Update(ctx context.Context, req infer.UpdateRequest[GroupArgs, GroupState]) (infer.UpdateResponse[GroupState], error) {
	panic("updating groups is not supported")
}

func (*Group) Delete(ctx context.Context, req infer.DeleteRequest[GroupState]) (infer.DeleteResponse, error) {
	p.GetLogger(ctx).Infof("deleting group %s", req.ID)

	config := infer.GetConfig[Config](ctx)
	client := config.client

	_, err := client.DeleteGroup(ctx, tursoclient.DeleteGroupParams{
		OrganizationSlug: config.OrganizationSlug,
		GroupName:        req.ID,
	})
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete group: %w", err)
	}

	return infer.DeleteResponse{}, nil
}

func (*Group) Diff(ctx context.Context, req infer.DiffRequest[GroupArgs, GroupState]) (infer.DiffResponse, error) {
	olds := req.State
	news := req.Inputs
	diff := map[string]p.PropertyDiff{}
	deleteBeforeReplace := false
	if olds.Name != news.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.Primary != news.PrimaryLocation {
		diff["primaryLocation"] = p.PropertyDiff{Kind: p.UpdateReplace}
		deleteBeforeReplace = true
	}
	return infer.DiffResponse{
		DeleteBeforeReplace: deleteBeforeReplace,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
}

func (config Config) readGroupResource(ctx context.Context, name string) (GroupState, error) {
	db, err := config.readGroup(ctx, name)
	if err != nil {
		return GroupState{}, err
	}

	return GroupState{
		DeleteProtection: db.DeleteProtection.Value,
		Name:             db.Name.Value,
		Locations:        db.GetLocations(),
		Primary:          db.Primary.Value,
		UUID:             db.UUID.Value,
	}, nil
}

func (config Config) readGroup(ctx context.Context, name string) (tursoclient.Group, error) {
	resp, err := config.client.GetGroup(ctx, tursoclient.GetGroupParams{
		OrganizationSlug: config.OrganizationSlug,
		GroupName:        name,
	})
	if err != nil {
		return tursoclient.Group{}, fmt.Errorf("client error: %w", err)
	}
	dbData, ok := resp.(*tursoclient.GetGroupOK)
	if !ok {
		return tursoclient.Group{}, fmt.Errorf("unexpected response from server (%T): %v", dbData, dbData)
	}
	db := dbData.Group.Value
	p.GetLogger(ctx).Debugf("read group: %+v", db)
	return db, nil
}
