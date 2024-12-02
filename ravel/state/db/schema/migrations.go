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
	}
}
