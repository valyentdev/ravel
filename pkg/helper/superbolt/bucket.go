package superbolt

import (
	"errors"

	"go.etcd.io/bbolt"
)

func bucketFromBBolt(b *bbolt.Bucket, opt *Options) *Bucket {
	return &Bucket{
		bucket: b,
		opt:    opt,
	}
}

type Bucket struct {
	bucket *bbolt.Bucket
	opt    *Options
}

var ErrKeyNotFound = errors.New("key not found")

func (b *Bucket) Get(key []byte, dest any) error {
	data := b.bucket.Get(key)
	if data == nil {
		return ErrKeyNotFound
	}
	return b.opt.Unmarshal(data, dest)
}

func (b *Bucket) Create(key []byte, value any) error {
	v := b.bucket.Get(key)
	if v != nil {
		return ErrKeyExists
	}

	return b.Put(key, value)
}

func (b *Bucket) Put(key []byte, value any) error {
	data, err := b.opt.Marshal(value)
	if err != nil {
		return err
	}
	return b.bucket.Put(key, data)
}

func (b *Bucket) Bucket(name []byte) (*Bucket, error) {
	bucket := b.bucket.Bucket(name)
	if bucket == nil {
		return nil, ErrBucketNotFound
	}

	return bucketFromBBolt(bucket, b.opt), nil
}
func (b *Bucket) CreateBucket(key []byte) (*Bucket, error) {
	bucket, err := b.bucket.CreateBucket(key)
	if err != nil {
		return nil, err
	}
	return bucketFromBBolt(bucket, b.opt), nil
}
func (b *Bucket) CreateBucketIfNotExists(key []byte) (*Bucket, error) {
	bucket, err := b.bucket.CreateBucketIfNotExists(key)
	if err != nil {
		return nil, err
	}
	return bucketFromBBolt(bucket, b.opt), nil
}
func (b *Bucket) Cursor() *bbolt.Cursor {
	return b.bucket.Cursor()
}

func (b *Bucket) Delete(key []byte) error {
	return b.bucket.Delete(key)
}
func (b *Bucket) DeleteBucket(key []byte) error {
	return b.bucket.DeleteBucket(key)
}

func (b *Bucket) ForEachBucket(fn func(key []byte, bucket *Bucket) error) error {
	return b.bucket.ForEachBucket(func(k []byte) error {
		bucket := b.bucket.Bucket(k)
		if bucket == nil {
			return errors.New("bucket not found")
		}
		return fn(k, bucketFromBBolt(bucket, b.opt))
	})
}
