package superbolt

import "go.etcd.io/bbolt"

type Tx struct {
	tx      *bbolt.Tx
	options *Options
}

func (tx *Tx) Bucket(name []byte) (*Bucket, error) {
	bucket := tx.tx.Bucket(name)
	if bucket == nil {
		return nil, ErrBucketNotFound
	}
	return bucketFromBBolt(bucket, tx.options), nil
}

func (tx *Tx) CreateBucket(name []byte) (*Bucket, error) {
	bucket, err := tx.tx.CreateBucket(name)
	if err != nil {
		return nil, err
	}
	return bucketFromBBolt(bucket, tx.options), nil
}

func (tx *Tx) CreateBucketIfNotExists(name []byte) (*Bucket, error) {
	bucket, err := tx.tx.CreateBucketIfNotExists(name)
	if err != nil {
		return nil, err
	}
	return bucketFromBBolt(bucket, tx.options), nil
}

func (tx *Tx) DeleteBucket(name []byte) error {
	return tx.tx.DeleteBucket(name)
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}
