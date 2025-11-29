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

// TestGroupResource_ReplaceOnChanges verifies that changing name or primaryLocation
// triggers a replacement, not an update.
func TestGroupResource_ReplaceOnChanges(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	urn := presource.NewURN("test", "provider", "", "turso:index:Group", "test")

	// Test that changing the name triggers a replace
	t.Run("name change triggers replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-group",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":    property.New("old-name"),
				"primary": property.New("aws-us-east-1"),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":            property.New("new-name"),
				"primaryLocation": property.New("aws-us-east-1"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that name change is a replace
		nameDiff, ok := diff.DetailedDiff["name"]
		assert.True(t, ok, "expected name to be in diff")
		assert.Equal(t, p.UpdateReplace, nameDiff.Kind, "expected name change to trigger replace")
	})

	// Test that changing primaryLocation triggers a replace
	t.Run("primaryLocation change triggers replace", func(t *testing.T) {
		diff, err := server.Diff(p.DiffRequest{
			ID:  "test-group",
			Urn: urn,
			State: property.NewMap(map[string]property.Value{
				"name":            property.New("test-group"),
				"primaryLocation": property.New("aws-us-east-1"),
			}),
			Inputs: property.NewMap(map[string]property.Value{
				"name":            property.New("test-group"),
				"primaryLocation": property.New("aws-us-west-2"),
			}),
		})
		require.NoError(t, err)
		assert.True(t, diff.HasChanges)

		// Check that primaryLocation change is a replace (could be UpdateReplace or AddReplace)
		locationDiff, ok := diff.DetailedDiff["primaryLocation"]
		assert.True(t, ok, "expected primaryLocation to be in diff")
		isReplace := locationDiff.Kind == p.UpdateReplace || locationDiff.Kind == p.AddReplace
		assert.True(t, isReplace, "expected primaryLocation change to trigger replace, got %v", locationDiff.Kind)
	})
}

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
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
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

	err = server.Configure(p.ConfigureRequest{})
	require.NoError(t, err)

	groupName := fmt.Sprintf("test-%d", rand.IntN(100000))
	integration.LifeCycleTest{
		Resource: "turso:index:Group",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
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
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
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
