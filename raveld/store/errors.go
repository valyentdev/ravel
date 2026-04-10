package store

import "errors"

// ErrBucketNotFound is returned when a required database bucket is missing
// This indicates the database was not properly initialized
var ErrBucketNotFound = errors.New("bucket not found: database not initialized")
