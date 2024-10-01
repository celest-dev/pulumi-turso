package provider

import (
	"testing"

	"github.com/blang/semver"
	"github.com/golang-jwt/jwt/v5"
	provider "github.com/pulumi/pulumi-go-provider"
	integration "github.com/pulumi/pulumi-go-provider/integration"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseTokenResource(t *testing.T) {
	server := integration.NewServer("turso", semver.Version{Minor: 1}, Provider())
	err := server.Configure(provider.ConfigureRequest{})
	assert.NoError(t, err)

	jwtParser := jwt.NewParser(jwt.WithoutClaimsValidation())

	const dbName = "test"
	integration.LifeCycleTest{
		Resource: "turso:index:DatabaseToken",
		Create: integration.Operation{
			Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
				"database":      dbName,
				"authorization": "full-access",
			}),
			Hook: func(inputs, output presource.PropertyMap) {
				t.Logf("Outputs: %v", output)
				name := output["database"].StringValue()
				assert.Equal(t, dbName, name)
				exp := output["expiresAt"].StringValue()
				assert.Empty(t, exp)

				token := output["token"].StringValue()
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
				Inputs: presource.NewPropertyMapFromMap(map[string]interface{}{
					"database":      dbName,
					"authorization": "full-access",
					"expiration":    "1h",
				}),
				Hook: func(inputs, output presource.PropertyMap) {
					t.Logf("Outputs: %v", output)
					exp := output["expiresAt"].StringValue()
					assert.NotEmpty(t, exp)

					token := output["token"].StringValue()
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
