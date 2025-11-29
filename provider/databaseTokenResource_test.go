package provider

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/golang-jwt/jwt/v5"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	integration "github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseTokenDiff_AuthorizationPointerComparison(t *testing.T) {
	t.Parallel()

	token := &DatabaseToken{}

	// Test case 1: Both nil - no diff
	t.Run("both nil", func(t *testing.T) {
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:      "test-db",
					Authorization: nil,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:      "test-db",
				Authorization: nil,
			},
		})
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	// Test case 2: Same value, different pointers - no diff
	t.Run("same value different pointers", func(t *testing.T) {
		authOld := DatabaseTokenAuthorization("full-access")
		authNew := DatabaseTokenAuthorization("full-access")
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:      "test-db",
					Authorization: &authOld,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:      "test-db",
				Authorization: &authNew,
			},
		})
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	// Test case 3: Different values - should diff
	t.Run("different values", func(t *testing.T) {
		authOld := DatabaseTokenAuthorization("full-access")
		authNew := DatabaseTokenAuthorization("read-only")
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:      "test-db",
					Authorization: &authOld,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:      "test-db",
				Authorization: &authNew,
			},
		})
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Contains(t, resp.DetailedDiff, "authorization")
	})

	// Test case 4: Old nil, new has value - should diff
	t.Run("old nil new has value", func(t *testing.T) {
		authNew := DatabaseTokenAuthorization("full-access")
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:      "test-db",
					Authorization: nil,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:      "test-db",
				Authorization: &authNew,
			},
		})
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Contains(t, resp.DetailedDiff, "authorization")
	})

	// Test case 5: Old has value, new nil - should diff
	t.Run("old has value new nil", func(t *testing.T) {
		authOld := DatabaseTokenAuthorization("full-access")
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:      "test-db",
					Authorization: &authOld,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:      "test-db",
				Authorization: nil,
			},
		})
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Contains(t, resp.DetailedDiff, "authorization")
	})
}

func TestDatabaseTokenDiff_ExpirationPointerComparison(t *testing.T) {
	t.Parallel()

	token := &DatabaseToken{}

	// Test case: Same expiration value, different pointers - no diff
	t.Run("same value different pointers", func(t *testing.T) {
		expOld := "1h"
		expNew := "1h"
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:   "test-db",
					Expiration: &expOld,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:   "test-db",
				Expiration: &expNew,
			},
		})
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	// Test case: Different expiration values - should diff
	t.Run("different values", func(t *testing.T) {
		expOld := "1h"
		expNew := "2h"
		resp, err := token.Diff(context.Background(), infer.DiffRequest[DatabaseTokenArgs, DatabaseTokenState]{
			State: DatabaseTokenState{
				DatabaseTokenArgs: DatabaseTokenArgs{
					Database:   "test-db",
					Expiration: &expOld,
				},
			},
			Inputs: DatabaseTokenArgs{
				Database:   "test-db",
				Expiration: &expNew,
			},
		})
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Contains(t, resp.DetailedDiff, "expiration")
	})
}

func TestDatabaseTokenResource(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{})
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	jwtParser := jwt.NewParser(jwt.WithoutClaimsValidation())

	const dbName = "test"
	integration.LifeCycleTest{
		Resource: "turso:index:DatabaseToken",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
				"database":      dbName,
				"authorization": "full-access",
			})),
			Hook: func(inputs, output property.Map) {
				t.Logf("Outputs: %v", output)
				name := output.Get("database").AsString()
				assert.Equal(t, dbName, name)
				exp := output.Get("expiresAt").AsString()
				assert.Empty(t, exp)

				token := output.Get("token").AsString()
				assert.NotEmpty(t, token)
				claims := jwt.MapClaims{}
				_, _, err := jwtParser.ParseUnverified(token, &claims)
				assert.NoError(t, err)
				parsedExp, err := claims.GetExpirationTime()
				assert.NoError(t, err)
				assert.Empty(t, parsedExp)
			},
		},
		Updates: []integration.Operation{
			{
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
					"database":      dbName,
					"authorization": "full-access",
					"expiration":    "1h",
				})),
				Hook: func(inputs, output property.Map) {
					t.Logf("Outputs: %v", output)
					exp := output.Get("expiresAt").AsString()
					assert.NotEmpty(t, exp)

					token := output.Get("token").AsString()
					assert.NotEmpty(t, token)
					claims := jwt.MapClaims{}
					_, _, err := jwtParser.ParseUnverified(token, &claims)
					assert.NoError(t, err)
					parsedExp, err := claims.GetExpirationTime()
					assert.NoError(t, err)
					assert.NotEmpty(t, parsedExp)
				},
			},
		},
	}.Run(t, server)
}

// TestDatabaseTokenResource_NoUnnecessaryReplace verifies that re-running with the same
// authorization value doesn't trigger an unnecessary replacement.
// This is a regression test for the pointer comparison bug.
func TestDatabaseTokenResource_NoUnnecessaryReplace(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{})
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	const dbName = "test"
	var originalToken string

	integration.LifeCycleTest{
		Resource: "turso:index:DatabaseToken",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
				"database":      dbName,
				"authorization": "read-only",
			})),
			Hook: func(inputs, output property.Map) {
				originalToken = output.Get("token").AsString()
				assert.NotEmpty(t, originalToken)
			},
		},
		Updates: []integration.Operation{
			{
				// Same inputs - should NOT trigger a replacement
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]any{
					"database":      dbName,
					"authorization": "read-only",
				})),
				Hook: func(inputs, output property.Map) {
					// The token should be the same since no replacement occurred
					currentToken := output.Get("token").AsString()
					assert.Equal(t, originalToken, currentToken, "Token should not have changed when inputs are identical")
				},
			},
		},
	}.Run(t, server)
}
