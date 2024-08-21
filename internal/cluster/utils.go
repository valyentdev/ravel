package cluster

type scannable interface {
	Scan(dest ...interface{}) error
}
