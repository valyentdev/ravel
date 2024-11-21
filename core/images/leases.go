package images

import (
	"sync"
)

type ImagesUsage struct {
	leases map[string]int
	lock   sync.Mutex
}

func NewImagesUsage() *ImagesUsage {
	return &ImagesUsage{
		leases: make(map[string]int),
	}
}

// Lock should be held when calling this function
func (iu *ImagesUsage) Usage() map[string]struct{} {
	images := make(map[string]struct{})
	for ref, _ := range iu.leases {
		images[ref] = struct{}{}
	}
	return images
}

func (iu *ImagesUsage) IsUsed(ref string) bool {
	_, ok := iu.leases[ref]
	return ok
}

func (iu *ImagesUsage) Lock() {
	iu.lock.Lock()
}
func (iu *ImagesUsage) Unlock() {
	iu.lock.Unlock()
}

// UseImage increments the lease count for the image with the given ref.
// Lock should be held when calling this function
func (iu *ImagesUsage) UseImage(ref string) {
	_, ok := iu.leases[ref]
	if !ok {
		iu.leases[ref] = 1
		return
	}
	iu.leases[ref]++
}

// ReleaseImage releases the lease on the image with the given ref.
// Lock should be held when calling this function
func (iu *ImagesUsage) ReleaseImage(ref string) {
	lease, ok := iu.leases[ref]
	if !ok {
		return
	}

	if lease == 1 {
		delete(iu.leases, ref)
		return
	}
}
