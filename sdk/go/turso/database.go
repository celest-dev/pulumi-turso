// Code generated by pulumi-language-go DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package turso

import (
	"context"
	"reflect"

	"errors"
	"github.com/celest-dev/pulumi-turso/sdk/go/turso/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Database struct {
	pulumi.CustomResourceState

	AllowAttach   pulumi.BoolOutput              `pulumi:"allowAttach"`
	Archived      pulumi.BoolOutput              `pulumi:"archived"`
	BlockReads    pulumi.BoolOutput              `pulumi:"blockReads"`
	BlockWrites   pulumi.BoolOutput              `pulumi:"blockWrites"`
	DbId          pulumi.StringOutput            `pulumi:"dbId"`
	Group         pulumi.StringOutput            `pulumi:"group"`
	Hostname      pulumi.StringOutput            `pulumi:"hostname"`
	Instances     DatabaseInstanceStateMapOutput `pulumi:"instances"`
	IsSchema      pulumi.BoolOutput              `pulumi:"isSchema"`
	Name          pulumi.StringOutput            `pulumi:"name"`
	PrimaryRegion pulumi.StringOutput            `pulumi:"primaryRegion"`
	Regions       pulumi.StringArrayOutput       `pulumi:"regions"`
	Schema        pulumi.StringOutput            `pulumi:"schema"`
	SizeLimit     pulumi.StringOutput            `pulumi:"sizeLimit"`
	Type          pulumi.StringOutput            `pulumi:"type"`
	Version       pulumi.StringOutput            `pulumi:"version"`
}

// NewDatabase registers a new resource with the given unique name, arguments, and options.
func NewDatabase(ctx *pulumi.Context,
	name string, args *DatabaseArgs, opts ...pulumi.ResourceOption) (*Database, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Group == nil {
		return nil, errors.New("invalid value for required argument 'Group'")
	}
	if args.Name == nil {
		return nil, errors.New("invalid value for required argument 'Name'")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource Database
	err := ctx.RegisterResource("turso:index:Database", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetDatabase gets an existing Database resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetDatabase(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *DatabaseState, opts ...pulumi.ResourceOption) (*Database, error) {
	var resource Database
	err := ctx.ReadResource("turso:index:Database", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Database resources.
type databaseState struct {
}

type DatabaseState struct {
}

func (DatabaseState) ElementType() reflect.Type {
	return reflect.TypeOf((*databaseState)(nil)).Elem()
}

type databaseArgs struct {
	AllowAttach *bool             `pulumi:"allowAttach"`
	BlockReads  *bool             `pulumi:"blockReads"`
	BlockWrites *bool             `pulumi:"blockWrites"`
	Group       string            `pulumi:"group"`
	IsSchema    *bool             `pulumi:"isSchema"`
	Name        string            `pulumi:"name"`
	Schema      *string           `pulumi:"schema"`
	Seed        *DatabaseSeedArgs `pulumi:"seed"`
	SizeLimit   *string           `pulumi:"sizeLimit"`
}

// The set of arguments for constructing a Database resource.
type DatabaseArgs struct {
	AllowAttach pulumi.BoolPtrInput
	BlockReads  pulumi.BoolPtrInput
	BlockWrites pulumi.BoolPtrInput
	Group       pulumi.StringInput
	IsSchema    pulumi.BoolPtrInput
	Name        pulumi.StringInput
	Schema      pulumi.StringPtrInput
	Seed        DatabaseSeedArgsPtrInput
	SizeLimit   pulumi.StringPtrInput
}

func (DatabaseArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*databaseArgs)(nil)).Elem()
}

type DatabaseInput interface {
	pulumi.Input

	ToDatabaseOutput() DatabaseOutput
	ToDatabaseOutputWithContext(ctx context.Context) DatabaseOutput
}

func (*Database) ElementType() reflect.Type {
	return reflect.TypeOf((**Database)(nil)).Elem()
}

func (i *Database) ToDatabaseOutput() DatabaseOutput {
	return i.ToDatabaseOutputWithContext(context.Background())
}

func (i *Database) ToDatabaseOutputWithContext(ctx context.Context) DatabaseOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DatabaseOutput)
}

// DatabaseArrayInput is an input type that accepts DatabaseArray and DatabaseArrayOutput values.
// You can construct a concrete instance of `DatabaseArrayInput` via:
//
//	DatabaseArray{ DatabaseArgs{...} }
type DatabaseArrayInput interface {
	pulumi.Input

	ToDatabaseArrayOutput() DatabaseArrayOutput
	ToDatabaseArrayOutputWithContext(context.Context) DatabaseArrayOutput
}

type DatabaseArray []DatabaseInput

func (DatabaseArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Database)(nil)).Elem()
}

func (i DatabaseArray) ToDatabaseArrayOutput() DatabaseArrayOutput {
	return i.ToDatabaseArrayOutputWithContext(context.Background())
}

func (i DatabaseArray) ToDatabaseArrayOutputWithContext(ctx context.Context) DatabaseArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DatabaseArrayOutput)
}

// DatabaseMapInput is an input type that accepts DatabaseMap and DatabaseMapOutput values.
// You can construct a concrete instance of `DatabaseMapInput` via:
//
//	DatabaseMap{ "key": DatabaseArgs{...} }
type DatabaseMapInput interface {
	pulumi.Input

	ToDatabaseMapOutput() DatabaseMapOutput
	ToDatabaseMapOutputWithContext(context.Context) DatabaseMapOutput
}

type DatabaseMap map[string]DatabaseInput

func (DatabaseMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Database)(nil)).Elem()
}

func (i DatabaseMap) ToDatabaseMapOutput() DatabaseMapOutput {
	return i.ToDatabaseMapOutputWithContext(context.Background())
}

func (i DatabaseMap) ToDatabaseMapOutputWithContext(ctx context.Context) DatabaseMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DatabaseMapOutput)
}

type DatabaseOutput struct{ *pulumi.OutputState }

func (DatabaseOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Database)(nil)).Elem()
}

func (o DatabaseOutput) ToDatabaseOutput() DatabaseOutput {
	return o
}

func (o DatabaseOutput) ToDatabaseOutputWithContext(ctx context.Context) DatabaseOutput {
	return o
}

func (o DatabaseOutput) AllowAttach() pulumi.BoolOutput {
	return o.ApplyT(func(v *Database) pulumi.BoolOutput { return v.AllowAttach }).(pulumi.BoolOutput)
}

func (o DatabaseOutput) Archived() pulumi.BoolOutput {
	return o.ApplyT(func(v *Database) pulumi.BoolOutput { return v.Archived }).(pulumi.BoolOutput)
}

func (o DatabaseOutput) BlockReads() pulumi.BoolOutput {
	return o.ApplyT(func(v *Database) pulumi.BoolOutput { return v.BlockReads }).(pulumi.BoolOutput)
}

func (o DatabaseOutput) BlockWrites() pulumi.BoolOutput {
	return o.ApplyT(func(v *Database) pulumi.BoolOutput { return v.BlockWrites }).(pulumi.BoolOutput)
}

func (o DatabaseOutput) DbId() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.DbId }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Group() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Group }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Hostname() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Hostname }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Instances() DatabaseInstanceStateMapOutput {
	return o.ApplyT(func(v *Database) DatabaseInstanceStateMapOutput { return v.Instances }).(DatabaseInstanceStateMapOutput)
}

func (o DatabaseOutput) IsSchema() pulumi.BoolOutput {
	return o.ApplyT(func(v *Database) pulumi.BoolOutput { return v.IsSchema }).(pulumi.BoolOutput)
}

func (o DatabaseOutput) Name() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Name }).(pulumi.StringOutput)
}

func (o DatabaseOutput) PrimaryRegion() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.PrimaryRegion }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Regions() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *Database) pulumi.StringArrayOutput { return v.Regions }).(pulumi.StringArrayOutput)
}

func (o DatabaseOutput) Schema() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Schema }).(pulumi.StringOutput)
}

func (o DatabaseOutput) SizeLimit() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.SizeLimit }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Type() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Type }).(pulumi.StringOutput)
}

func (o DatabaseOutput) Version() pulumi.StringOutput {
	return o.ApplyT(func(v *Database) pulumi.StringOutput { return v.Version }).(pulumi.StringOutput)
}

type DatabaseArrayOutput struct{ *pulumi.OutputState }

func (DatabaseArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Database)(nil)).Elem()
}

func (o DatabaseArrayOutput) ToDatabaseArrayOutput() DatabaseArrayOutput {
	return o
}

func (o DatabaseArrayOutput) ToDatabaseArrayOutputWithContext(ctx context.Context) DatabaseArrayOutput {
	return o
}

func (o DatabaseArrayOutput) Index(i pulumi.IntInput) DatabaseOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *Database {
		return vs[0].([]*Database)[vs[1].(int)]
	}).(DatabaseOutput)
}

type DatabaseMapOutput struct{ *pulumi.OutputState }

func (DatabaseMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Database)(nil)).Elem()
}

func (o DatabaseMapOutput) ToDatabaseMapOutput() DatabaseMapOutput {
	return o
}

func (o DatabaseMapOutput) ToDatabaseMapOutputWithContext(ctx context.Context) DatabaseMapOutput {
	return o
}

func (o DatabaseMapOutput) MapIndex(k pulumi.StringInput) DatabaseOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *Database {
		return vs[0].(map[string]*Database)[vs[1].(string)]
	}).(DatabaseOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*DatabaseInput)(nil)).Elem(), &Database{})
	pulumi.RegisterInputType(reflect.TypeOf((*DatabaseArrayInput)(nil)).Elem(), DatabaseArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*DatabaseMapInput)(nil)).Elem(), DatabaseMap{})
	pulumi.RegisterOutputType(DatabaseOutput{})
	pulumi.RegisterOutputType(DatabaseArrayOutput{})
	pulumi.RegisterOutputType(DatabaseMapOutput{})
}
