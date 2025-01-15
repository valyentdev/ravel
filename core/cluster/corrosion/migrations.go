package corrosion

import (
	"context"
	"errors"
	"fmt"

	"github.com/valyentdev/corroclient"
)

var schema = []string{
	// Gateways
	`CREATE table gateways (
    	id text not null primary key default '',
    	namespace text not null default '',
    	fleet_id text not null default '',
    	name text not null default '',
    	protocol text not null default 'https',
    	target_port int not null default 80
	);`,

	// Instances
	`CREATE TABLE instances (
    	id text not null default '',
    	node text not null default '',
    	namespace text not null default '',
    	machine_id text not null default '',
    	machine_version text not null default '',
    	status text not null default '',
    	created_at integer not null default 0,
    	updated_at integer not null default 0,
    	local_ipv4 text not null default '',
    	events text not null default '[]',
    	enable_machine_gateway integer not null default 0,
    	primary key (id, machine_id)
	);`,
	`CREATE index instances_machine_id on instances(machine_id);`,
	`CREATE index instances_node on instances(node);`,

	// Machines
	`CREATE TABLE machines (
		id text primary key not null default "",
		namespace text not null default "",
		fleet_id text not null default "",
		instance_id text not null default "",
		machine_version text not null default "",
		node text not null default "",
		region text not null default "",
		created_at integer not null default 0,
		updated_at integer not null default 0,
		destroyed_at integer
	);`,
	`CREATE index machines_node on machines(node);`,
	`CREATE index machines_instance_id on machines(instance_id);`,
	`CREATE index machines_fleet_id on machines(fleet_id);`,
	`CREATE index machines_namespace on machines(namespace);`,
	`CREATE index machines_region on machines(region);`,

	// MachineVersions
	`CREATE TABLE machine_versions (
    	id text primary key not null default "",
    	namespace text not null default "",
    	machine_id text not null default "",
    	config text not null default "{}",
    	resources text not null default "{}"
);`,
	// Nodes
	`CREATE TABLE nodes (
		id text primary key not null default '',
		address text not null default '',
		agent_port integer not null default 0,
		http_proxy_port integer not null default 0,
		region text not null default '',
		heartbeated_at integer not null default 0
	);`,
	`CREATE index nodes_address on nodes(address);`,
	`CREATE index nodes_region on nodes(region);`,
}

func RunCorrosionMigrations(ctx context.Context, config corroclient.Config) error {

	client := corroclient.NewCorroClient(config)

	result, err := client.UpdateSchema(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to run update schema: %w", err)
	}

	errs := result.Errors()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
