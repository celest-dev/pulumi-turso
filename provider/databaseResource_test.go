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

func TestDatabaseResource(t *testing.T) {
	server := integration.NewServer("turso", semver.Version{Minor: 1}, Provider())
	err := server.Configure(provider.ConfigureRequest{
		Args: presource.NewPropertyMapFromMap(map[string]interface{}{
			"organization": "celest-dev",
		}),
	})
	assert.NoError(t, err)

	dbName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Database",
		Create: integration.Operation{
			Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":  dbName,
				"group": "test",
			}),
			Hook: func(inputs, output presource.PropertyMap) {
				t.Logf("Outputs: %v", output)
				name := output["name"].StringValue()
				assert.Equal(t, dbName, name)
				group := output["group"].StringValue()
				assert.Equal(t, "test", group)
				dbId := output["dbId"].StringValue()
				assert.NotEmpty(t, dbId)
				allowAttach := output["allowAttach"].BoolValue()
				assert.False(t, allowAttach)
				blockReads := output["blockReads"].BoolValue()
				assert.False(t, blockReads)
				blockWrites := output["blockWrites"].BoolValue()
				assert.False(t, blockWrites)
				sizeLimit := output["sizeLimit"].StringValue()
				assert.Empty(t, sizeLimit)
			},
		},
		Updates: []integration.Operation{
			{
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":        dbName,
					"group":       "test",
					"allowAttach": true,
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					allowAttach := output["allowAttach"].BoolValue()
					assert.True(t, allowAttach)
				},
			},
			{
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":       dbName,
					"group":      "test",
					"blockReads": true,
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					blockReads := output["blockReads"].BoolValue()
					assert.True(t, blockReads)
				},
			},
			{
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":        dbName,
					"group":       "test",
					"blockWrites": true,
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					blockWrites := output["blockWrites"].BoolValue()
					assert.True(t, blockWrites)
				},
			},
			{
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":      dbName,
					"group":     "test",
					"sizeLimit": "1gb",
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					sizeLimit := output["sizeLimit"].StringValue()
					assert.Equal(t, "1gb", sizeLimit)
				},
			},
		},
	}.Run(t, server)
}
