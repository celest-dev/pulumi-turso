package provider

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/blang/semver"
	p "github.com/pulumi/pulumi-go-provider"
	integration "github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupResource(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{})
	require.NoError(t, err)

	groupName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Group",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":            groupName,
				"primaryLocation": "aws-us-east-1",
			})),
			Hook: func(inputs, output property.Map) {
				t.Logf("Outputs: %v", output)
				name := output.Get("name").AsString()
				assert.Equal(t, groupName, name)
				uuid := output.Get("uuid").AsString()
				assert.NotEmpty(t, uuid)
			},
		},
	}.Run(t, server)
}

func TestGroupResource_ChangeName(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{
		Args: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
			"organization": "celest-dev",
		})),
	})
	require.NoError(t, err)

	groupName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Group",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":            groupName,
				"primaryLocation": "aws-us-east-1",
			})),
			Hook: func(inputs, output property.Map) {
				t.Logf("Outputs: %v", output)
				name := output.Get("name").AsString()
				assert.Equal(t, groupName, name)
				uuid := output.Get("uuid").AsString()
				assert.NotEmpty(t, uuid)
			},
		},
		Updates: []integration.Operation{
			{
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":            groupName + "-updated",
					"primaryLocation": "aws-us-east-1",
				})),
				Hook: func(inputs, output property.Map) {
					t.Logf("Outputs: %v", output)
					name := output.Get("name").AsString()
					assert.Equal(t, groupName+"-updated", name)
				},
			},
		},
	}.Run(t, server)
}
