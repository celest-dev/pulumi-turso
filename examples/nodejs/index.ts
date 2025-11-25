import * as pulumi from "@pulumi/pulumi";
import * as random from "@pulumi/random";
import * as turso from "@celest-dev/pulumi-turso";

const databaseTag = new random.RandomId("databaseTag", {byteLength: 4});
const databaseResource = new turso.Database("database", {
    group: "test",
    name: pulumi.interpolate`test-${databaseTag.hex}`,
});
const databaseTokenResource = new turso.DatabaseToken("databaseToken", {
    database: databaseResource.name,
    expiration: "1h",
    authorization: turso.DatabaseTokenAuthorization.Read_only,
});
export const database = {
    hostname: databaseResource.hostname,
};
export const databaseToken = {
    token: databaseTokenResource.token,
    expiration: databaseTokenResource.expiration,
};
