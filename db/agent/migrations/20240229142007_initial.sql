-- migrate:up
CREATE TABLE instances (
    "id" text primary key not null,
    "reservation_id" text not null references reservations(id),
    "prepared" boolean not null,
    "desired_status" text not null,
    "restarts" integer not null,
    "machine_id" text not null,
    "fleet_id" text not null,
    "node_id" text not null,
    "namespace" text not null,
    "config" jsonb not null,
    "created_at" timestamp not null,
    "destroyed" boolean not null,
    "destroyed_at" timestamp
);

-- migrate:down
DROP TABLE instances;