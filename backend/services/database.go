package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sing_panel/models"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketConfig   = []byte("config")
	bucketVersions = []byte("versions")
	bucketState    = []byte("state")
	bucketSingbox  = []byte("singbox")
)

type Database struct {
	db *bolt.DB
}

func NewDatabase(path string) (*Database, error) {
	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketConfig, bucketVersions, bucketState, bucketSingbox} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create buckets: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// Get retrieves a value by bucket and key
func (d *Database) Get(bucket, key string, dest interface{}) error {
	return d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key %q not found", key)
		}
		return json.Unmarshal(data, dest)
	})
}

// Put stores a value by bucket and key
func (d *Database) Put(bucket, key string, value interface{}) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), data)
	})
}

// Delete removes a key from a bucket
func (d *Database) Delete(bucket, key string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}

// HasKey checks if a key exists
func (d *Database) HasKey(bucket, key string) bool {
	exists := false
	d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			exists = b.Get([]byte(key)) != nil
		}
		return nil
	})
	return exists
}

// UpdateKernelState updates the kernel state in place
func (d *Database) UpdateKernelState(fn func(state *models.KernelState)) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("state"))
		if b == nil {
			return fmt.Errorf("bucket state not found")
		}

		var state models.KernelState
		data := b.Get([]byte("kernel"))
		if data != nil {
			if err := json.Unmarshal(data, &state); err != nil {
				return err
			}
		}

		fn(&state)

		newData, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return b.Put([]byte("kernel"), newData)
	})
}

// ListBuckets returns all bucket names
func (d *Database) ListBuckets() ([]string, error) {
	var buckets []string
	err := d.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, string(name))
			return nil
		})
	})
	return buckets, err
}

// ListKeys returns all keys in a bucket
func (d *Database) ListKeys(bucket string) ([]string, error) {
	var keys []string
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		return b.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

// GetValue returns the raw value for a bucket/key as string
func (d *Database) GetValue(bucket, key string) (string, error) {
	var result string
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("key %q not found", key)
		}
		// Try to format as JSON
		var pretty bytes.Buffer
		if json.Indent(&pretty, data, "", "  ") == nil {
			result = pretty.String()
		} else {
			result = string(data)
		}
		return nil
	})
	return result, err
}

// PutValue stores a raw string value for a bucket/key
func (d *Database) PutValue(bucket, key, value string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		// Validate JSON
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("invalid JSON")
		}
		return b.Put([]byte(key), []byte(value))
	})
}

// DeleteKey removes a key from a bucket
func (d *Database) DeleteKey(bucket, key string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		return b.Delete([]byte(key))
	})
}
