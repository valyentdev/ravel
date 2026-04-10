package schema

const gwNameUp = `
ALTER TABLE gateways RENAME COLUMN name TO old_name;
ALTER TABLE gateways ADD COLUMN name text not null default '';
UPDATE gateways SET name = old_name || '-' || namespace;
ALTER TABLE gateways ADD CONSTRAINT unique_gateway_name UNIQUE (name);
ALTER TABLE gateways DROP COLUMN old_name;
`

const gwNameDown = `` // This migration is not reversible

const UniqueGatewayNameConstraint = "unique_gateway_name"
