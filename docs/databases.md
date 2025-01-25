# Ravel databases

In its cluster configuation, Ravel rely on two kind of databases to store and share its state:
- [Corrosion](https://github.com/superfly/corrosion) to propagate the shared state accross the cluster.
- [Postgres](https://www.postgresql.org/) to store the state of the Ravel server in a strongly consistent way.

##  Corrosion

Corrosion is a distributed SQLite database that is used to share the state of the Ravel cluster accross the nodes. It is used to store the state of the machines, gateways... and to propagate the state changes to the other nodes.

You can find a example configuration in the [config.toml](../examples/corrosion-config.toml) file.

To install corrosion, you can build it from the source code or use the pre-built binaries available on the [releases page](https://github.com/superfly/corrosion) or use our [release](https://github.com/valyentdev/corrosion/releases/tag/2024-10-21).

Then you can start the corrosion instance with the following command for example:

```bash
$ corrosion agent -c /etc/ravel/corrosion-config.toml
```

### Migrations

To migrate a corrosion node you can use the following command:

```bash
$ ravel corrosion migrate -c /path/to/ravel/config.toml
```

## Postgres

The Ravel server store its state in a Postgres database.

To run the migrations you can use the following command:

```bash
$ ravel db migrate --to latest
$ ravel db status
```

Output:
```bash
Current version: 3
Total migrations: 3

Migrations:
[✓] 1  initial
[✓] 2  gateways
[✓] 3  gateway_name
```