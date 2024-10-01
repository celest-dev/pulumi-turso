package main

import (
	"time"

	"github.com/celest-dev/pulumi-turso/sdk/go/turso"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		database, err := turso.NewDatabase(ctx, "database", &turso.DatabaseArgs{
			Group: pulumi.String("my-group"),
			Name:  pulumi.String("my-database"),
		})
		if err != nil {
			return err
		}
		ctx.Export("databaseOutput", database)

		databaseToken, err := turso.NewDatabaseToken(ctx, "databaseToken", &turso.DatabaseTokenArgs{
			Database:      pulumi.String("my-database"),
			Authorization: turso.DatabaseTokenAuthorization_Full_Access,
			Expiration:    pulumi.String(time.Duration(24 * time.Hour).String()),
		})
		if err != nil {
			return err
		}
		ctx.Export("databaseTokenOutput", databaseToken)

		return nil
	})
}
