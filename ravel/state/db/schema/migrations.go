package schema

func getMigrations() []Migration {
	return []Migration{
		{
			Name: "initial",
			Up:   initialUp,
			Down: initialDown,
		},
		{
			Name: "gateways",
			Up:   gatewaysUp,
			Down: gatewaysDown,
		},
		{
			Name: "gateway_name",
			Up:   gwNameUp,
			Down: gwNameDown,
		},
		{
			Name: "metadata",
			Up:   metadataUp,
			Down: metadataDown,
		},
		{
			Name: "secrets",
			Up:   secretsUp,
			Down: secretsDown,
		},
	}
}
