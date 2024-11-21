package networking

type SubnetAllocator interface {
	Allocate(*Network) error
	Release(*Network) error
	AllocateNext() (Network, error)
}
