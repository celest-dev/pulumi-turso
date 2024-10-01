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

type DatabaseArgs struct {
	Group       string            `pulumi:"group"`
	Name        string            `pulumi:"name"`
	AllowAttach *bool             `pulumi:"allowAttach,optional"`
	BlockReads  *bool             `pulumi:"blockReads,optional"`
	BlockWrites *bool             `pulumi:"blockWrites,optional"`
	SizeLimit   *string           `pulumi:"sizeLimit,optional"`
	IsSchema    *bool             `pulumi:"isSchema,optional"`
	Schema      *string           `pulumi:"schema,optional"`
	Seed        *DatabaseSeedArgs `pulumi:"seed,optional"`
}

type DatabaseSeedArgs struct {
	Type      DatabaseSeedType `pulumi:"type"`
	Name      *string          `pulumi:"name,optional"`
	Timestamp *time.Time       `pulumi:"timestamp,optional"`
	URL       *string          `pulumi:"url,optional"`
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
	AllowAttach   bool     `pulumi:"allowAttach" json:"allow_attach"`
	Archived      bool     `pulumi:"archived" json:"archived"`
	BlockReads    bool     `pulumi:"blockReads" json:"block_reads"`
	BlockWrites   bool     `pulumi:"blockWrites" json:"block_writes"`
	DbId          string   `pulumi:"dbId" json:"db_id"`
	Group         string   `pulumi:"group" json:"group"`
	Hostname      string   `pulumi:"hostname" json:"hostname"`
	IsSchema      bool     `pulumi:"isSchema" json:"is_schema"`
	Name          string   `pulumi:"name" json:"name"`
	PrimaryRegion string   `pulumi:"primaryRegion" json:"primary_region"`
	Regions       []string `pulumi:"regions" json:"regions"`
	Schema        string   `pulumi:"schema" json:"schema"`
	SizeLimit     string   `pulumi:"sizeLimit" json:"size_limit"`
	Type          string   `pulumi:"type" json:"type"`
	Version       string   `pulumi:"version" json:"version"`

	Instances map[string]DatabaseInstanceState `pulumi:"instances" json:"instances"`
}

type DatabaseInstanceState struct {
	Hostname string `pulumi:"hostname" json:"hostname"`
	Name     string `pulumi:"name" json:"name"`
	Region   string `pulumi:"region" json:"region"`
	Type     string `pulumi:"type" json:"type"`
	UUID     string `pulumi:"uuid" json:"uuid"`
}

var (
	_ infer.CustomCreate[DatabaseArgs, DatabaseState] = Database{}
	_ infer.CustomRead[DatabaseArgs, DatabaseState]   = Database{}
	_ infer.CustomUpdate[DatabaseArgs, DatabaseState] = Database{}
	_ infer.CustomDelete[DatabaseState]               = Database{}
	_ infer.CustomDiff[DatabaseArgs, DatabaseState]   = Database{}
)

func (Database) Create(ctx context.Context, name string, input DatabaseArgs, preview bool) (string, DatabaseState, error) {
	p.GetLogger(ctx).Infof("creating database %s (preview=%v)", name, preview)
	if preview {
		return input.Name, DatabaseState{
			Name:        input.Name,
			Group:       input.Group,
			AllowAttach: UnwrapOrZero(input.AllowAttach),
			BlockReads:  UnwrapOrZero(input.BlockReads),
			BlockWrites: UnwrapOrZero(input.BlockWrites),
			IsSchema:    UnwrapOrZero(input.IsSchema),
			Schema:      UnwrapOrZero(input.Schema),
			SizeLimit:   UnwrapOrZero(input.SizeLimit),
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client

	var dbSeed tursoclient.OptCreateDatabaseInputSeed
	if seed := input.Seed; seed != nil {
		dbSeed = tursoclient.NewOptCreateDatabaseInputSeed(tursoclient.CreateDatabaseInputSeed{
			Type:      tursoclient.NewOptCreateDatabaseInputSeedType(tursoclient.CreateDatabaseInputSeedType(seed.Type)),
			Name:      optString(seed.Name),
			URL:       optString(seed.URL),
			Timestamp: optTime(seed.Timestamp),
		})
	}

	createReq := tursoclient.CreateDatabaseInput{
		Name:      input.Name,
		Group:     input.Group,
		Seed:      dbSeed,
		SizeLimit: optString(input.SizeLimit),
		IsSchema:  optBool(input.IsSchema),
		Schema:    optString(input.Schema),
	}
	res, err := client.CreateDatabase(ctx, &createReq, tursoclient.CreateDatabaseParams{
		OrganizationName: config.OrganizationName,
	})
	if err != nil {
		return "", DatabaseState{}, fmt.Errorf("failed to create database: %w\n%v", err, res)
	}
	_, ok := res.(*tursoclient.CreateDatabaseOK)
	if !ok {
		return "", DatabaseState{}, fmt.Errorf("error creating database. unexpected response from server (%T): %v", res, res)
	}

	updateConfigReq := tursoclient.DatabaseConfigurationInput{
		AllowAttach: optBool(input.AllowAttach),
		BlockReads:  optBool(input.BlockReads),
		BlockWrites: optBool(input.BlockWrites),
	}
	_, err = config.client.UpdateDatabaseConfiguration(ctx, &updateConfigReq, tursoclient.UpdateDatabaseConfigurationParams{
		OrganizationName: config.OrganizationName,
		DatabaseName:     input.Name,
	})
	if err != nil {
		return "", DatabaseState{}, fmt.Errorf("failed to update database configuration: %w", err)
	}

	state, err := config.readDatabaseResource(ctx, input.Name)
	if err != nil {
		return "", DatabaseState{}, fmt.Errorf("failed to read database: %w", err)
	}

	return state.Name, state, nil
}

func (Database) Read(ctx context.Context, id string, inputs DatabaseArgs, state DatabaseState) (canonicalID string, normalizedInputs DatabaseArgs, normalizedState DatabaseState, err error) {
	p.GetLogger(ctx).Infof("reading database %s", id)

	config := infer.GetConfig[Config](ctx)
	normalizedState, err = config.readDatabaseResource(ctx, id)
	if err != nil {
		return "", DatabaseArgs{}, DatabaseState{}, fmt.Errorf("failed to read database: %w", err)
	}

	return id, inputs, normalizedState, nil
}

func (Database) Update(ctx context.Context, id string, olds DatabaseState, news DatabaseArgs, preview bool) (DatabaseState, error) {
	p.GetLogger(ctx).Infof("updating database %s (preview=%v)", id, preview)

	if preview {
		return DatabaseState{
			Name:        news.Name,
			Group:       news.Group,
			AllowAttach: UnwrapOrZero(news.AllowAttach),
			BlockReads:  UnwrapOrZero(news.BlockReads),
			BlockWrites: UnwrapOrZero(news.BlockWrites),
			IsSchema:    UnwrapOrZero(news.IsSchema),
			Schema:      UnwrapOrZero(news.Schema),
			SizeLimit:   UnwrapOrZero(news.SizeLimit),
		}, nil
	}

	config := infer.GetConfig[Config](ctx)
	client := config.client
	updateReq := tursoclient.DatabaseConfigurationInput{
		AllowAttach: optBool(news.AllowAttach),
		BlockReads:  optBool(news.BlockReads),
		BlockWrites: optBool(news.BlockWrites),
		SizeLimit:   optString(news.SizeLimit),
	}
	_, err := client.UpdateDatabaseConfiguration(ctx, &updateReq, tursoclient.UpdateDatabaseConfigurationParams{
		OrganizationName: config.OrganizationName,
		DatabaseName:     id,
	})
	if err != nil {
		return DatabaseState{}, fmt.Errorf("failed to update database: %w", err)
	}

	state, err := config.readDatabaseResource(ctx, id)
	if err != nil {
		return DatabaseState{}, fmt.Errorf("failed to read database: %w", err)
	}

	return state, nil
}

func (Database) Delete(ctx context.Context, id string, props DatabaseState) error {
	p.GetLogger(ctx).Infof("deleting database %s", id)

	config := infer.GetConfig[Config](ctx)
	client := config.client

	_, err := client.DeleteDatabase(ctx, tursoclient.DeleteDatabaseParams{
		OrganizationName: config.OrganizationName,
		DatabaseName:     id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	return nil
}

func (Database) Diff(ctx context.Context, id string, olds DatabaseState, news DatabaseArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if olds.AllowAttach != UnwrapOrZero(news.AllowAttach) {
		diff["allowAttach"] = p.PropertyDiff{Kind: p.Update}
	}
	if olds.BlockReads != UnwrapOrZero(news.BlockReads) {
		diff["blockReads"] = p.PropertyDiff{Kind: p.Update}
	}
	if olds.BlockWrites != UnwrapOrZero(news.BlockWrites) {
		diff["blockWrites"] = p.PropertyDiff{Kind: p.Update}
	}
	if olds.SizeLimit != UnwrapOrZero(news.SizeLimit) {
		diff["sizeLimit"] = p.PropertyDiff{Kind: p.Update}
	}
	if olds.Group != news.Group {
		diff["group"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.Name != news.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.IsSchema != UnwrapOrZero(news.IsSchema) {
		diff["isSchema"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if olds.Schema != UnwrapOrZero(news.Schema) {
		diff["schema"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	return p.DiffResponse{
		DeleteBeforeReplace: true,
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
	}, nil
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
		OrganizationName: config.OrganizationName,
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
		AllowAttach:   db.AllowAttach.Value,
		Archived:      db.Archived.Value,
		BlockReads:    db.BlockReads.Value,
		BlockWrites:   db.BlockWrites.Value,
		DbId:          db.DbId.Value,
		Group:         db.Group.Value,
		Hostname:      db.Hostname.Value,
		IsSchema:      db.IsSchema.Value,
		Name:          db.Name.Value,
		PrimaryRegion: db.PrimaryRegion.Value,
		Regions:       db.GetRegions(),
		Schema:        db.Schema.Value,
		SizeLimit:     dbConfig.SizeLimit.Value,
		Type:          db.Type.Value,
		Version:       db.Version.Value,

		Instances: instances,
	}, nil
}

func (config Config) readDatabase(ctx context.Context, name string) (tursoclient.Database, error) {
	resp, err := config.client.GetDatabase(ctx, tursoclient.GetDatabaseParams{
		OrganizationName: config.OrganizationName,
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
		OrganizationName: config.OrganizationName,
		DatabaseName:     name,
	})
	if err != nil {
		return nil, fmt.Errorf("client error: %w", err)
	}
	p.GetLogger(ctx).Debugf("read database configuration: %+v", resp)
	return resp, nil
}
