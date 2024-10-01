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

type GroupToken struct {
	pulumi.CustomResourceState

	Authorization GroupTokenAuthorizationPtrOutput `pulumi:"authorization"`
	Database      pulumi.StringOutput              `pulumi:"database"`
	Expiration    pulumi.StringPtrOutput           `pulumi:"expiration"`
	ExpiresAt     pulumi.StringPtrOutput           `pulumi:"expiresAt"`
	ReadAttach    pulumi.StringArrayOutput         `pulumi:"readAttach"`
	Token         pulumi.StringOutput              `pulumi:"token"`
}

// NewGroupToken registers a new resource with the given unique name, arguments, and options.
func NewGroupToken(ctx *pulumi.Context,
	name string, args *GroupTokenArgs, opts ...pulumi.ResourceOption) (*GroupToken, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Database == nil {
		return nil, errors.New("invalid value for required argument 'Database'")
	}
	secrets := pulumi.AdditionalSecretOutputs([]string{
		"token",
	})
	opts = append(opts, secrets)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource GroupToken
	err := ctx.RegisterResource("turso:index:GroupToken", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetGroupToken gets an existing GroupToken resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetGroupToken(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *GroupTokenState, opts ...pulumi.ResourceOption) (*GroupToken, error) {
	var resource GroupToken
	err := ctx.ReadResource("turso:index:GroupToken", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering GroupToken resources.
type groupTokenState struct {
}

type GroupTokenState struct {
}

func (GroupTokenState) ElementType() reflect.Type {
	return reflect.TypeOf((*groupTokenState)(nil)).Elem()
}

type groupTokenArgs struct {
	Authorization *GroupTokenAuthorization `pulumi:"authorization"`
	Database      string                   `pulumi:"database"`
	Expiration    *string                  `pulumi:"expiration"`
	ReadAttach    []string                 `pulumi:"readAttach"`
}

// The set of arguments for constructing a GroupToken resource.
type GroupTokenArgs struct {
	Authorization GroupTokenAuthorizationPtrInput
	Database      pulumi.StringInput
	Expiration    pulumi.StringPtrInput
	ReadAttach    pulumi.StringArrayInput
}

func (GroupTokenArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*groupTokenArgs)(nil)).Elem()
}

type GroupTokenInput interface {
	pulumi.Input

	ToGroupTokenOutput() GroupTokenOutput
	ToGroupTokenOutputWithContext(ctx context.Context) GroupTokenOutput
}

func (*GroupToken) ElementType() reflect.Type {
	return reflect.TypeOf((**GroupToken)(nil)).Elem()
}

func (i *GroupToken) ToGroupTokenOutput() GroupTokenOutput {
	return i.ToGroupTokenOutputWithContext(context.Background())
}

func (i *GroupToken) ToGroupTokenOutputWithContext(ctx context.Context) GroupTokenOutput {
	return pulumi.ToOutputWithContext(ctx, i).(GroupTokenOutput)
}

// GroupTokenArrayInput is an input type that accepts GroupTokenArray and GroupTokenArrayOutput values.
// You can construct a concrete instance of `GroupTokenArrayInput` via:
//
//	GroupTokenArray{ GroupTokenArgs{...} }
type GroupTokenArrayInput interface {
	pulumi.Input

	ToGroupTokenArrayOutput() GroupTokenArrayOutput
	ToGroupTokenArrayOutputWithContext(context.Context) GroupTokenArrayOutput
}

type GroupTokenArray []GroupTokenInput

func (GroupTokenArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*GroupToken)(nil)).Elem()
}

func (i GroupTokenArray) ToGroupTokenArrayOutput() GroupTokenArrayOutput {
	return i.ToGroupTokenArrayOutputWithContext(context.Background())
}

func (i GroupTokenArray) ToGroupTokenArrayOutputWithContext(ctx context.Context) GroupTokenArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(GroupTokenArrayOutput)
}

// GroupTokenMapInput is an input type that accepts GroupTokenMap and GroupTokenMapOutput values.
// You can construct a concrete instance of `GroupTokenMapInput` via:
//
//	GroupTokenMap{ "key": GroupTokenArgs{...} }
type GroupTokenMapInput interface {
	pulumi.Input

	ToGroupTokenMapOutput() GroupTokenMapOutput
	ToGroupTokenMapOutputWithContext(context.Context) GroupTokenMapOutput
}

type GroupTokenMap map[string]GroupTokenInput

func (GroupTokenMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*GroupToken)(nil)).Elem()
}

func (i GroupTokenMap) ToGroupTokenMapOutput() GroupTokenMapOutput {
	return i.ToGroupTokenMapOutputWithContext(context.Background())
}

func (i GroupTokenMap) ToGroupTokenMapOutputWithContext(ctx context.Context) GroupTokenMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(GroupTokenMapOutput)
}

type GroupTokenOutput struct{ *pulumi.OutputState }

func (GroupTokenOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**GroupToken)(nil)).Elem()
}

func (o GroupTokenOutput) ToGroupTokenOutput() GroupTokenOutput {
	return o
}

func (o GroupTokenOutput) ToGroupTokenOutputWithContext(ctx context.Context) GroupTokenOutput {
	return o
}

func (o GroupTokenOutput) Authorization() GroupTokenAuthorizationPtrOutput {
	return o.ApplyT(func(v *GroupToken) GroupTokenAuthorizationPtrOutput { return v.Authorization }).(GroupTokenAuthorizationPtrOutput)
}

func (o GroupTokenOutput) Database() pulumi.StringOutput {
	return o.ApplyT(func(v *GroupToken) pulumi.StringOutput { return v.Database }).(pulumi.StringOutput)
}

func (o GroupTokenOutput) Expiration() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *GroupToken) pulumi.StringPtrOutput { return v.Expiration }).(pulumi.StringPtrOutput)
}

func (o GroupTokenOutput) ExpiresAt() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *GroupToken) pulumi.StringPtrOutput { return v.ExpiresAt }).(pulumi.StringPtrOutput)
}

func (o GroupTokenOutput) ReadAttach() pulumi.StringArrayOutput {
	return o.ApplyT(func(v *GroupToken) pulumi.StringArrayOutput { return v.ReadAttach }).(pulumi.StringArrayOutput)
}

func (o GroupTokenOutput) Token() pulumi.StringOutput {
	return o.ApplyT(func(v *GroupToken) pulumi.StringOutput { return v.Token }).(pulumi.StringOutput)
}

type GroupTokenArrayOutput struct{ *pulumi.OutputState }

func (GroupTokenArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*GroupToken)(nil)).Elem()
}

func (o GroupTokenArrayOutput) ToGroupTokenArrayOutput() GroupTokenArrayOutput {
	return o
}

func (o GroupTokenArrayOutput) ToGroupTokenArrayOutputWithContext(ctx context.Context) GroupTokenArrayOutput {
	return o
}

func (o GroupTokenArrayOutput) Index(i pulumi.IntInput) GroupTokenOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *GroupToken {
		return vs[0].([]*GroupToken)[vs[1].(int)]
	}).(GroupTokenOutput)
}

type GroupTokenMapOutput struct{ *pulumi.OutputState }

func (GroupTokenMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*GroupToken)(nil)).Elem()
}

func (o GroupTokenMapOutput) ToGroupTokenMapOutput() GroupTokenMapOutput {
	return o
}

func (o GroupTokenMapOutput) ToGroupTokenMapOutputWithContext(ctx context.Context) GroupTokenMapOutput {
	return o
}

func (o GroupTokenMapOutput) MapIndex(k pulumi.StringInput) GroupTokenOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *GroupToken {
		return vs[0].(map[string]*GroupToken)[vs[1].(string)]
	}).(GroupTokenOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*GroupTokenInput)(nil)).Elem(), &GroupToken{})
	pulumi.RegisterInputType(reflect.TypeOf((*GroupTokenArrayInput)(nil)).Elem(), GroupTokenArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*GroupTokenMapInput)(nil)).Elem(), GroupTokenMap{})
	pulumi.RegisterOutputType(GroupTokenOutput{})
	pulumi.RegisterOutputType(GroupTokenArrayOutput{})
	pulumi.RegisterOutputType(GroupTokenMapOutput{})
}
