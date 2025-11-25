package provider

import (
	"testing"

	"github.com/blang/semver"
	"github.com/golang-jwt/jwt/v5"
	p "github.com/pulumi/pulumi-go-provider"
	integration "github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseTokenResource(t *testing.T) {
	t.Parallel()

	server, err := integration.NewServer(t.Context(),
		"turso",
		semver.Version{Minor: 1},
		integration.WithProvider(Provider()),
	)
	require.NoError(t, err)

	err = server.Configure(p.ConfigureRequest{})
	require.NoError(t, err)

	jwtParser := jwt.NewParser(jwt.WithoutClaimsValidation())

	const dbName = "test"
	integration.LifeCycleTest{
		Resource: "turso:index:DatabaseToken",
		Create: integration.Operation{
			Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
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
				Inputs: presource.FromResourcePropertyMap(presource.NewPropertyMapFromMap(map[string]interface{}{
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
