package initd

const InitdPort = 64242

type WaitResult struct {
	ExitCode int `json:"exit_code"`
}

type SignalOptions struct {
	Signal int
}

type Status struct {
	Ok bool `json:"ok"`
}
