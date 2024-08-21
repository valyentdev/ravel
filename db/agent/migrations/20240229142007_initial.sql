-- migrate:up
CREATE TABLE instances (
    "id" text primary key not null,
    "reservation_id" text not null references reservations(id),
    "desired_status" text not null,
    "restarts" integer not null,
    "machine_id" text not null,
    "fleet_id" text not null,
    "node_id" text not null,
    "namespace" text not null,
    "config" jsonb not null,
    "destroyed" boolean not null
);

-- migrate:down
DROP TABLE instances;