package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Group struct{}

var _ infer.Annotated = (*Group)(nil)

func (g *Group) Annotate(a infer.Annotator) {
	a.Describe(&g, "A Turso group. Groups are collections of databases that share the same replication configuration and can be managed together.")
}

type GroupArgs struct {
	Name             string   `pulumi:"name" provider:"replaceOnChanges"`
	PrimaryLocation  string   `pulumi:"primaryLocation" provider:"replaceOnChanges"`
	ReplicaLocations []string `pulumi:"replicaLocations,optional"`
	Extensions       []string `pulumi:"extensions,optional"`
}

var _ infer.Annotated = (*GroupArgs)(nil)

func (args *GroupArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.Name, "The name of the group. Must be unique within the organization.")
	a.Describe(&args.PrimaryLocation, "The primary location (region) for the group. This is where the primary database instance will be created.")
	a.Describe(&args.ReplicaLocations, "Additional locations where database replicas will be created.")
	a.Describe(&args.Extensions, "SQLite extensions to enable for databases in this group. Use 'all' to enable all available extensions, or specify individual extension names.")
}

type GroupState struct {
	DeleteProtection bool     `pulumi:"deleteProtection" json:"delete_protection"`
	Locations        []string `pulumi:"locations" json:"locations"`
	Name             string   `pulumi:"name" json:"name"`
	Primary          string   `pulumi:"primary" json:"primary"`
	UUID             string   `pulumi:"uuid" json:"uuid"`
}

var _ infer.Annotated = (*GroupState)(nil)

func (s *GroupState) Annotate(a infer.Annotator) {
	a.Describe(&s.DeleteProtection, "Whether deletion protection is enabled for this group.")
	a.Describe(&s.Locations, "All locations where this group has database instances.")
	a.Describe(&s.Name, "The name of the group.")
	a.Describe(&s.Primary, "The primary location of the group.")
	a.Describe(&s.UUID, "The unique identifier of the group.")
}

var (
	_ infer.CustomCreate[GroupArgs, GroupState] = (*Group)(nil)
	_ infer.CustomRead[GroupArgs, GroupState]   = (*Group)(nil)
	_ infer.CustomDelete[GroupState]            = (*Group)(nil)
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
