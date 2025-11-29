package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type Database struct{}

var _ infer.Annotated = (*Database)(nil)

func (d *Database) Annotate(a infer.Annotator) {
	a.Describe(&d, "A Turso database. Databases are SQLite databases that live in a group and can be replicated across multiple regions.")
}

type DatabaseArgs struct {
	Group       string            `pulumi:"group" provider:"replaceOnChanges"`
	Name        string            `pulumi:"name" provider:"replaceOnChanges"`
	BlockReads  *bool             `pulumi:"blockReads,optional"`
	BlockWrites *bool             `pulumi:"blockWrites,optional"`
	SizeLimit   *string           `pulumi:"sizeLimit,optional"`
	Seed        *DatabaseSeedArgs `pulumi:"seed,optional" provider:"replaceOnChanges"`
}

var _ infer.Annotated = (*DatabaseArgs)(nil)

func (args *DatabaseArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.Group, "The name of the group where the database belongs. The group must already exist.")
	a.Describe(&args.Name, "The name of the database. Must be unique within the organization.")
	a.Describe(&args.BlockReads, "When true, read queries to this database will be blocked.")
	a.Describe(&args.BlockWrites, "When true, write queries to this database will be blocked.")
	a.Describe(&args.SizeLimit, "The maximum size of the database in bytes. You can use units like '1gb', '500mb', etc.")
	a.Describe(&args.Seed, "Configuration for seeding the database from an existing database or dump.")
}

type DatabaseSeedArgs struct {
	Type      DatabaseSeedType `pulumi:"type"`
	Name      *string          `pulumi:"name,optional"`
	Timestamp *time.Time       `pulumi:"timestamp,optional"`
}

var _ infer.Annotated = (*DatabaseSeedArgs)(nil)

func (args *DatabaseSeedArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.Type, "The type of seed to use.")
	a.Describe(&args.Name, "The name of the database to seed from (when type is 'database') or the URL of the dump file (when type is 'dump').")
	a.Describe(&args.Timestamp, "A specific point in time to seed from. Only applies when seeding from a database.")
}

type DatabaseSeedType string

const (
	DatabaseSeedTypeDatabase DatabaseSeedType = "database"
	DatabaseSeedTypeDump     DatabaseSeedType = "dump"
)

var _ infer.Enum[DatabaseSeedType] = (*DatabaseSeedType)(nil)

func (*DatabaseSeedType) Values() []infer.EnumValue[DatabaseSeedType] {
	return []infer.EnumValue[DatabaseSeedType]{
		{Value: DatabaseSeedTypeDatabase, Name: "Database", Description: "Uses an database to seed the new database."},
		{Value: DatabaseSeedTypeDump, Name: "Dump", Description: "Uses a database dump to seed the new database."},
	}
}

type DatabaseState struct {
	BlockReads       bool     `pulumi:"blockReads" json:"block_reads"`
	BlockWrites      bool     `pulumi:"blockWrites" json:"block_writes"`
	DbId             string   `pulumi:"dbId" json:"db_id"`
	DeleteProtection bool     `pulumi:"deleteProtection" json:"delete_protection"`
	Group            string   `pulumi:"group" json:"group"`
	Hostname         string   `pulumi:"hostname" json:"hostname"`
	Name             string   `pulumi:"name" json:"name"`
	PrimaryRegion    string   `pulumi:"primaryRegion" json:"primary_region"`
	Regions          []string `pulumi:"regions" json:"regions"`
	SizeLimit        string   `pulumi:"sizeLimit" json:"size_limit"`

	Instances map[string]DatabaseInstanceState `pulumi:"instances" json:"instances"`
}

var _ infer.Annotated = (*DatabaseState)(nil)

func (s *DatabaseState) Annotate(a infer.Annotator) {
	a.Describe(&s.BlockReads, "Whether read queries to this database are blocked.")
	a.Describe(&s.BlockWrites, "Whether write queries to this database are blocked.")
	a.Describe(&s.DbId, "The unique identifier of the database.")
	a.Describe(&s.DeleteProtection, "Whether deletion protection is enabled for this database.")
	a.Describe(&s.Group, "The name of the group this database belongs to.")
	a.Describe(&s.Hostname, "The hostname to connect to this database.")
	a.Describe(&s.Name, "The name of the database.")
	a.Describe(&s.PrimaryRegion, "The primary region where the database is located.")
	a.Describe(&s.Regions, "All regions where this database has replicas.")
	a.Describe(&s.SizeLimit, "The maximum size limit of the database.")
	a.Describe(&s.Instances, "A map of database instances by region.")
}

type DatabaseInstanceState struct {
	Hostname string `pulumi:"hostname" json:"hostname"`
	Name     string `pulumi:"name" json:"name"`
	Region   string `pulumi:"region" json:"region"`
	Type     string `pulumi:"type" json:"type"`
	UUID     string `pulumi:"uuid" json:"uuid"`
}

var _ infer.Annotated = (*DatabaseInstanceState)(nil)

func (s *DatabaseInstanceState) Annotate(a infer.Annotator) {
	a.Describe(&s.Hostname, "The hostname of this database instance.")
	a.Describe(&s.Name, "The name of this database instance.")
	a.Describe(&s.Region, "The region where this instance is located.")
	a.Describe(&s.Type, "The type of instance (primary or replica).")
	a.Describe(&s.UUID, "The unique identifier of this instance.")
}

var (
	_ infer.CustomCreate[DatabaseArgs, DatabaseState] = (*Database)(nil)
	_ infer.CustomRead[DatabaseArgs, DatabaseState]   = (*Database)(nil)
	_ infer.CustomUpdate[DatabaseArgs, DatabaseState] = (*Database)(nil)
	_ infer.CustomDelete[DatabaseState]               = (*Database)(nil)
)

func (*Database) Create(ctx context.Context, req infer.CreateRequest[DatabaseArgs]) (infer.CreateResponse[DatabaseState], error) {
	input := req.Inputs
	preview := req.DryRun
	p.GetLogger(ctx).Infof("creating database %s (preview=%v)", req.Name, preview)
	if preview {
		return infer.CreateResponse[DatabaseState]{
			ID: input.Name,
			Output: DatabaseState{
				Name:        input.Name,
				Group:       input.Group,
				BlockReads:  UnwrapOrZero(input.BlockReads),
				BlockWrites: UnwrapOrZero(input.BlockWrites),
				SizeLimit:   UnwrapOrZero(input.SizeLimit),
			},
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var dbSeed tursoclient.OptCreateDatabaseInputSeed
	if seed := input.Seed; seed != nil {
		dbSeed = tursoclient.NewOptCreateDatabaseInputSeed(tursoclient.CreateDatabaseInputSeed{
			Type:      tursoclient.NewOptCreateDatabaseInputSeedType(tursoclient.CreateDatabaseInputSeedType(seed.Type)),
			Name:      optString(seed.Name),
			Timestamp: optTime(seed.Timestamp),
		})
	}

	createReq := tursoclient.CreateDatabaseInput{
		Name:      input.Name,
		Group:     input.Group,
		Seed:      dbSeed,
		SizeLimit: optString(input.SizeLimit),
	}
	res, err := client.CreateDatabase(ctx, &createReq, tursoclient.CreateDatabaseParams{
		OrganizationSlug: config.OrganizationSlug,
	})
	if err != nil {
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("failed to create database: %w\n%v", err, res)
	}
	_, ok := res.(*tursoclient.CreateDatabaseOK)
	if !ok {
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("error creating database. unexpected response from server (%T): %v", res, res)
	}

	updateConfigReq := tursoclient.DatabaseConfigurationInput{
		BlockReads:  optBool(input.BlockReads),
		BlockWrites: optBool(input.BlockWrites),
	}
	_, err = config.client.UpdateDatabaseConfiguration(ctx, &updateConfigReq, tursoclient.UpdateDatabaseConfigurationParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     input.Name,
	})
	if err != nil {
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("failed to update database configuration: %w", err)
	}

	state, err := config.readDatabaseResource(ctx, input.Name)
	if err != nil {
		return infer.CreateResponse[DatabaseState]{}, fmt.Errorf("failed to read database: %w", err)
	}

	return infer.CreateResponse[DatabaseState]{ID: state.Name, Output: state}, nil
}

func (*Database) Read(ctx context.Context, req infer.ReadRequest[DatabaseArgs, DatabaseState]) (infer.ReadResponse[DatabaseArgs, DatabaseState], error) {
	p.GetLogger(ctx).Infof("reading database %s", req.ID)

	config := infer.GetConfig[Config](ctx)
	normalizedState, err := config.readDatabaseResource(ctx, req.ID)
	if err != nil {
		return infer.ReadResponse[DatabaseArgs, DatabaseState]{}, fmt.Errorf("failed to read database: %w", err)
	}

	return infer.ReadResponse[DatabaseArgs, DatabaseState]{
		ID:     req.ID,
		Inputs: req.Inputs,
		State:  normalizedState,
	}, nil
}

func (*Database) Update(ctx context.Context, req infer.UpdateRequest[DatabaseArgs, DatabaseState]) (infer.UpdateResponse[DatabaseState], error) {
	news := req.Inputs
	preview := req.DryRun
	p.GetLogger(ctx).Infof("updating database %s (preview=%v)", req.ID, preview)

	if preview {
		return infer.UpdateResponse[DatabaseState]{
			Output: DatabaseState{
				Name:        news.Name,
				Group:       news.Group,
				BlockReads:  UnwrapOrZero(news.BlockReads),
				BlockWrites: UnwrapOrZero(news.BlockWrites),
				SizeLimit:   UnwrapOrZero(news.SizeLimit),
			},
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client
	updateReq := tursoclient.DatabaseConfigurationInput{
		BlockReads:  optBool(news.BlockReads),
		BlockWrites: optBool(news.BlockWrites),
		SizeLimit:   optString(news.SizeLimit),
	}
	_, err := client.UpdateDatabaseConfiguration(ctx, &updateReq, tursoclient.UpdateDatabaseConfigurationParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     req.ID,
	})
	if err != nil {
		return infer.UpdateResponse[DatabaseState]{}, fmt.Errorf("failed to update database: %w", err)
	}

	state, err := config.readDatabaseResource(ctx, req.ID)
	if err != nil {
		return infer.UpdateResponse[DatabaseState]{}, fmt.Errorf("failed to read database: %w", err)
	}

	return infer.UpdateResponse[DatabaseState]{Output: state}, nil
}

func (*Database) Delete(ctx context.Context, req infer.DeleteRequest[DatabaseState]) (infer.DeleteResponse, error) {
	p.GetLogger(ctx).Infof("deleting database %s", req.ID)

	config := infer.GetConfig[Config](ctx)
	client := config.client

	_, err := client.DeleteDatabase(ctx, tursoclient.DeleteDatabaseParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     req.ID,
	})
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete database: %w", err)
	}

	return infer.DeleteResponse{}, nil
}

func optString(s *string) tursoclient.OptString {
	if s == nil {
		return tursoclient.OptString{}
	}
	return tursoclient.NewOptString(*s)
}

func optBool(b *bool) tursoclient.OptBool {
	if b == nil {
		return tursoclient.OptBool{}
	}
	return tursoclient.NewOptBool(*b)
}

func optTime(t *time.Time) tursoclient.OptString {
	if t == nil {
		return tursoclient.OptString{}
	}
	return tursoclient.NewOptString(t.Format(time.RFC3339))
}

// UnwrapOrZero unwraps a value when not nil or returns the zero value.
func UnwrapOrZero[VP *V, V any](v VP) V {
	if v == nil {
		return Zero[V]()
	}
	return *v
}

// Zero returns the zero value of a type.
func Zero[T any]() T {
	var zero T
	return zero
}

func (config Config) readDatabaseResource(ctx context.Context, name string) (DatabaseState, error) {
	db, err := config.readDatabase(ctx, name)
	if err != nil {
		return DatabaseState{}, err
	}
	dbConfig, err := config.readDatabaseConfiguration(ctx, name)
	if err != nil {
		return DatabaseState{}, err
	}
	dbInstances, err := config.client.ListDatabaseInstances(ctx, tursoclient.ListDatabaseInstancesParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     name,
	})
	if err != nil {
		return DatabaseState{}, fmt.Errorf("failed to list database instances: %w", err)
	}

	instances := make(map[string]DatabaseInstanceState)
	for _, instance := range dbInstances.GetInstances() {
		instances[instance.Region.Value] = DatabaseInstanceState{
			Hostname: instance.Hostname.Value,
			Name:     instance.Name.Value,
			Region:   instance.Region.Value,
			Type:     string(instance.Type.Value),
			UUID:     instance.UUID.Value,
		}
	}

	return DatabaseState{
		BlockReads:       db.BlockReads.Value,
		BlockWrites:      db.BlockWrites.Value,
		DbId:             db.DbId.Value,
		DeleteProtection: db.DeleteProtection.Value,
		Group:            db.Group.Value,
		Hostname:         db.Hostname.Value,
		Name:             db.Name.Value,
		PrimaryRegion:    db.PrimaryRegion.Value,
		Regions:          db.GetRegions(),
		SizeLimit:        dbConfig.SizeLimit.Value,

		Instances: instances,
	}, nil
}

func (config Config) readDatabase(ctx context.Context, name string) (tursoclient.Database, error) {
	resp, err := config.client.GetDatabase(ctx, tursoclient.GetDatabaseParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     name,
	})
	if err != nil {
		return tursoclient.Database{}, fmt.Errorf("client error: %w", err)
	}
	dbData, ok := resp.(*tursoclient.GetDatabaseOK)
	if !ok {
		return tursoclient.Database{}, fmt.Errorf("unexpected response from server (%T): %v", dbData, dbData)
	}
	db := dbData.Database.Value
	p.GetLogger(ctx).Debugf("read database: %+v", db)
	return db, nil
}

func (config Config) readDatabaseConfiguration(ctx context.Context, name string) (*tursoclient.DatabaseConfigurationResponse, error) {
	resp, err := config.client.GetDatabaseConfiguration(ctx, tursoclient.GetDatabaseConfigurationParams{
		OrganizationSlug: config.OrganizationSlug,
		DatabaseName:     name,
	})
	if err != nil {
		return nil, fmt.Errorf("client error: %w", err)
	}
	p.GetLogger(ctx).Debugf("read database configuration: %+v", resp)
	return resp, nil
}
