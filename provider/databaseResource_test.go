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

func TestDatabaseResource(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{})
	require.NoError(t, err)

	dbName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Database",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
				"name":  dbName,
				"group": "test",
			})),
			Hook: func(inputs, output property.Map) {
				t.Logf("Outputs: %v", output)
				name := output.Get("name").AsString()
				assert.Equal(t, dbName, name)
				group := output.Get("group").AsString()
				assert.Equal(t, "test", group)
				dbId := output.Get("dbId").AsString()
				assert.NotEmpty(t, dbId)
				blockReads := output.Get("blockReads").AsBool()
				assert.False(t, blockReads)
				blockWrites := output.Get("blockWrites").AsBool()
				assert.False(t, blockWrites)
				sizeLimit := output.Get("sizeLimit").AsString()
				assert.Empty(t, sizeLimit)
			},
		},
		Updates: []integration.Operation{
			{
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":       dbName,
					"group":      "test",
					"blockReads": true,
				})),
				Hook: func(inputs, output property.Map) {
					t.Logf("Outputs: %v", output)
					blockReads := output.Get("blockReads").AsBool()
					assert.True(t, blockReads)
				},
			},
			{
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":        dbName,
					"group":       "test",
					"blockWrites": true,
				})),
				Hook: func(inputs, output property.Map) {
					t.Logf("Outputs: %v", output)
					blockWrites := output.Get("blockWrites").AsBool()
					assert.True(t, blockWrites)
				},
			},
			{
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
					"name":      dbName,
					"group":     "test",
					"sizeLimit": "1gb",
				})),
				Hook: func(inputs, output property.Map) {
					t.Logf("Outputs: %v", output)
					sizeLimit := output.Get("sizeLimit").AsString()
					assert.Equal(t, "1gb", sizeLimit)
				},
			},
		},
	}.Run(t, server)
}
