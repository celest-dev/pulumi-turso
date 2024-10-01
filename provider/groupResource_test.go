package provider

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/blang/semver"
	provider "github.com/pulumi/pulumi-go-provider"
	integration "github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
)

func TestGroupResource(t *testing.T) {
	server := integration.NewServer("turso", semver.Version{Minor: 1}, Provider())
	err := server.Configure(provider.ConfigureRequest{})
	assert.NoError(t, err)

	groupName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Group",
		Create: integration.Operation{
			Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":            groupName,
				"primaryLocation": "sjc",
			}),
			Hook: func(inputs, output presource.PropertyMap) {
				t.Logf("Outputs: %v", output)
				name := output["name"].StringValue()
				assert.Equal(t, groupName, name)
				uuid := output["uuid"].StringValue()
				assert.NotEmpty(t, uuid)
				archived := output["archived"].BoolValue()
				assert.False(t, archived)
			},
		},
	}.Run(t, server)
}

func TestGroupResource_ChangeName(t *testing.T) {
	server := integration.NewServer("turso", semver.Version{Minor: 1}, Provider())
	err := server.Configure(provider.ConfigureRequest{
		Args: presource.NewPropertyMapFromMap(map[string]interface{}{
			"organization": "celest-dev",
		}),
	})
	assert.NoError(t, err)

	groupName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Group",
		Create: integration.Operation{
			Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":            groupName,
				"primaryLocation": "sjc",
			}),
			Hook: func(inputs, output presource.PropertyMap) {
				t.Logf("Outputs: %v", output)
				name := output["name"].StringValue()
				assert.Equal(t, groupName, name)
				uuid := output["uuid"].StringValue()
				assert.NotEmpty(t, uuid)
				archived := output["archived"].BoolValue()
				assert.False(t, archived)
			},
		},
		Updates: []integration.Operation{
			{
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":            groupName + "-updated",
					"primaryLocation": "sjc",
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					name := output["name"].StringValue()
					assert.Equal(t, groupName+"-updated", name)
				},
			},
		},
	}.Run(t, server)
}
