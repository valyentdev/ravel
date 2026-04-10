package schema

const secretsUp = `
-- Create secrets table for storing sensitive data
CREATE TABLE secrets (
    "id" text PRIMARY KEY,
    "name" text NOT NULL,
    "namespace" text NOT NULL REFERENCES namespaces(name) ON DELETE CASCADE,
    "value" text NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT timezone('utc', now()),
    "updated_at" timestamp NOT NULL DEFAULT timezone('utc', now())
);

-- Ensure unique secret names within a namespace
CREATE UNIQUE INDEX secrets_name_namespace_idx ON secrets(namespace, name);

-- Index for efficient lookups by namespace
CREATE INDEX secrets_namespace_idx ON secrets(namespace);
`

const secretsDown = `
DROP INDEX IF EXISTS secrets_namespace_idx;
DROP INDEX IF EXISTS secrets_name_namespace_idx;
DROP TABLE IF EXISTS secrets;
`
