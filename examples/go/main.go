package main

import (
	"fmt"

	"github.com/celest-dev/pulumi-turso/sdk/go/turso"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		databaseTag, err := random.NewRandomId(ctx, "databaseTag", &random.RandomIdArgs{
			ByteLength: pulumi.Int(4),
		})
		if err != nil {
			return err
		}
		databaseResource, err := turso.NewDatabase(ctx, "database", &turso.DatabaseArgs{
			Group: pulumi.String("test"),
			Name: databaseTag.Hex.ApplyT(func(hex string) (string, error) {
				return fmt.Sprintf("test-%v", hex), nil
			}).(pulumi.StringOutput),
		})
		if err != nil {
			return err
		}
		databaseTokenResource, err := turso.NewDatabaseToken(ctx, "databaseToken", &turso.DatabaseTokenArgs{
			Database:      databaseResource.Name,
			Expiration:    pulumi.String("1h"),
			Authorization: turso.DatabaseTokenAuthorization_Read_Only,
		})
		if err != nil {
			return err
		}
		ctx.Export("database", pulumi.StringMap{
			"value": databaseResource.Name,
		})
		ctx.Export("databaseToken", pulumi.StringMap{
			"value": databaseTokenResource.Token,
		})
		return nil
	})
}
