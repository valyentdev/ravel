package schema

const metadataUp = `
-- Add metadata column to fleets table
ALTER TABLE fleets ADD COLUMN metadata jsonb NOT NULL DEFAULT '{}';

-- Add metadata column to machines table
ALTER TABLE machines ADD COLUMN metadata jsonb NOT NULL DEFAULT '{}';

-- Create GIN indexes for efficient label filtering
CREATE INDEX fleets_metadata_labels_idx ON fleets USING gin ((metadata->'labels'));
CREATE INDEX machines_metadata_labels_idx ON machines USING gin ((metadata->'labels'));
`

const metadataDown = `
DROP INDEX IF EXISTS machines_metadata_labels_idx;
DROP INDEX IF EXISTS fleets_metadata_labels_idx;
ALTER TABLE machines DROP COLUMN metadata;
ALTER TABLE fleets DROP COLUMN metadata;
`
