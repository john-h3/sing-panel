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
	bucketConfig    = []byte("config")
	bucketState     = []byte("state")
	bucketSingbox   = []byte("singbox")
	bucketInstances = []byte("instances")
)

// preservedBuckets are never replaced by ImportAll: they hold machine-local
// runtime state (state) and this panel's own managed instances (instances).
var preservedBuckets = map[string]bool{
	"state":     true,
	"instances": true,
}

// The app_config key in the config bucket. Its dashboards field holds
// machine-local service URLs (e.g. "http://<this-host>:9090/ui"), so it is
// excluded from syncing and from the configuration fingerprint: every panel
// keeps its own dashboards setting.
const (
	appConfigKey             = "app_config"
	appConfigDashboardsField = "dashboards"
)

// stripAppConfigDashboards removes the dashboards field from a serialized
// app_config value, leaving all other fields untouched.
func stripAppConfigDashboards(raw string) string {
	var conf map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &conf); err != nil {
		return raw
	}
	delete(conf, appConfigDashboardsField)
	out, err := json.Marshal(conf)
	if err != nil {
		return raw
	}
	return string(out)
}

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
		for _, b := range [][]byte{bucketConfig, bucketState, bucketSingbox, bucketInstances} {
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

// Close returns the database file size in bytes
func (d *Database) DBFileSize() int64 {
	info, err := os.Stat(d.db.Path())
	if err != nil {
		return 0
	}
	return info.Size()
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

// DeleteBucket removes a bucket. Only empty buckets can be deleted, so that
// functional data is protected from accidental removal.
func (d *Database) DeleteBucket(bucket string) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucket)
		}
		if k, _ := b.Cursor().First(); k != nil {
			return fmt.Errorf("bucket %q is not empty, delete keys first", bucket)
		}
		return tx.DeleteBucket([]byte(bucket))
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

// ExportAll returns the entire database as bucket -> key -> value.
func (d *Database) ExportAll() (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	err := d.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			keys := make(map[string]string)
			if err := b.ForEach(func(k, v []byte) error {
				keys[string(k)] = string(v)
				return nil
			}); err != nil {
				return err
			}
			result[string(name)] = keys
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ImportAll replaces the whole database with the given data. Existing
// buckets/keys that are not present in data are removed, except for
// preservedBuckets (local runtime state and managed instances), which are
// kept as-is. The local dashboards setting is always kept: the imported
// app_config never overwrites it.
func (d *Database) ImportAll(data map[string]map[string]string) error {
	merged := make(map[string]map[string]string, len(data))
	for name, keys := range data {
		merged[name] = keys
	}
	if raw, ok := merged["config"][appConfigKey]; ok {
		merged["config"] = d.keepLocalDashboards(merged["config"], raw)
	}
	return d.db.Update(func(tx *bolt.Tx) error {
		var buckets [][]byte
		if err := tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			if !preservedBuckets[string(name)] {
				buckets = append(buckets, name)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, name := range buckets {
			if err := tx.DeleteBucket(name); err != nil {
				return err
			}
		}
		for name, keys := range merged {
			if preservedBuckets[name] {
				continue
			}
			b, err := tx.CreateBucket([]byte(name))
			if err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
			for k, v := range keys {
				if err := b.Put([]byte(k), []byte(v)); err != nil {
					return fmt.Errorf("put %q/%q: %w", name, k, err)
				}
			}
		}
		return nil
	})
}

// keepLocalDashboards removes the dashboards field from the imported
// app_config value and merges this panel's own dashboards setting back in.
func (d *Database) keepLocalDashboards(configKeys map[string]string, raw string) map[string]string {
	imported := stripAppConfigDashboards(raw)
	var local map[string]json.RawMessage
	var localDashboards json.RawMessage
	if err := d.Get("config", appConfigKey, &local); err == nil {
		localDashboards = local[appConfigDashboardsField]
	}
	if localDashboards == nil {
		configKeys[appConfigKey] = imported
		return configKeys
	}
	var conf map[string]json.RawMessage
	if err := json.Unmarshal([]byte(imported), &conf); err != nil {
		return configKeys
	}
	conf[appConfigDashboardsField] = localDashboards
	if out, err := json.Marshal(conf); err == nil {
		configKeys[appConfigKey] = string(out)
	}
	return configKeys
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
