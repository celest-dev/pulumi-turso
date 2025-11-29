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

// TestDatabaseResource_ReplaceOnChanges verifies that changing name, group, or seed
// triggers a replacement, while blockReads/blockWrites/sizeLimit are updateable.
func TestDatabaseResource_ReplaceOnChanges(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	urn := presource.NewURN("test", "provider", "", "turso:index:Database", "test")

	// Test that changing the name triggers a replace
	t.Run("name change triggers replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-db",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":  property.New("old-name"),
				"group": property.New("test-group"),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":  property.New("new-name"),
				"group": property.New("test-group"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that name change is a replace
		nameDiff, ok := diff.DetailedDiff["name"]
		assert.True(t, ok, "expected name to be in diff")
		assert.Equal(t, p.UpdateReplace, nameDiff.Kind, "expected name change to trigger replace")
	})

	// Test that changing group triggers a replace
	t.Run("group change triggers replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-db",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":  property.New("test-db"),
				"group": property.New("old-group"),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":  property.New("test-db"),
				"group": property.New("new-group"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that group change is a replace
		groupDiff, ok := diff.DetailedDiff["group"]
		assert.True(t, ok, "expected group to be in diff")
		assert.Equal(t, p.UpdateReplace, groupDiff.Kind, "expected group change to trigger replace")
	})

	// Test that changing blockReads does NOT trigger a replace (it's updateable)
	t.Run("blockReads change does not trigger replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-db",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":       property.New("test-db"),
				"group":      property.New("test-group"),
				"blockReads": property.New(false),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":       property.New("test-db"),
				"group":      property.New("test-group"),
				"blockReads": property.New(true),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that blockReads change is NOT a replace
		blockReadsDiff, ok := diff.DetailedDiff["blockReads"]
		assert.True(t, ok, "expected blockReads to be in diff")
		isReplace := blockReadsDiff.Kind == p.AddReplace || blockReadsDiff.Kind == p.DeleteReplace || blockReadsDiff.Kind == p.UpdateReplace
		assert.False(t, isReplace, "expected blockReads change to NOT trigger replace, got %v", blockReadsDiff.Kind)
	})

	// Test that changing sizeLimit does NOT trigger a replace (it's updateable)
	t.Run("sizeLimit change does not trigger replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-db",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":      property.New("test-db"),
				"group":     property.New("test-group"),
				"sizeLimit": property.New(""),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":      property.New("test-db"),
				"group":     property.New("test-group"),
				"sizeLimit": property.New("1gb"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that sizeLimit change is NOT a replace
		sizeLimitDiff, ok := diff.DetailedDiff["sizeLimit"]
		assert.True(t, ok, "expected sizeLimit to be in diff")
		isReplace := sizeLimitDiff.Kind == p.AddReplace || sizeLimitDiff.Kind == p.DeleteReplace || sizeLimitDiff.Kind == p.UpdateReplace
		assert.False(t, isReplace, "expected sizeLimit change to NOT trigger replace, got %v", sizeLimitDiff.Kind)
	})
}

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
