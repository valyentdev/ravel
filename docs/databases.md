# Ravel databases

Ravel uses [Postgres](https://www.postgresql.org/) to store the state of the Ravel server in a strongly consistent way.

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
