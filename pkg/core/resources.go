package core

type Resources struct {
	Cpus   int `json:"cpus"`   // in MHz
	Memory int `json:"memory"` // in MB
}

func (r *Resources) Sub(other Resources) Resources {
	new := Resources{
		Cpus:   r.Cpus - other.Cpus,
		Memory: r.Memory - other.Memory,
	}
	return new
}

// Add returns a new Resources object which is the sum of the resources.
func (r *Resources) Add(other Resources) Resources {
	new := Resources{
		Cpus:   r.Cpus + other.Cpus,
		Memory: r.Memory + other.Memory,
	}
	return new
}

// GT returns true if the resources are greater than the other resources.
func (r *Resources) GT(other Resources) bool {
	return r.Cpus > other.Cpus && r.Memory > other.Memory
}
