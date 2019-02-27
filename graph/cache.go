package graph

import (
	"errors"
	"log"
	"path/filepath"
	"strings"
)

// ItemCache caches DriveItems for a filesystem. This cache never expires so
// that local changes can persist.
type ItemCache struct {
	root *DriveItem // will be a nil pointer on start, lazily initialized
}

// Get fetches a given DriveItem in the cache, if any items along the way are
// not found, they are fetched.
func (c *ItemCache) Get(key string, auth Auth) (*DriveItem, error) {
	// lazily initialize root of filesystem
	if c.root == nil {
		root, err := GetItem("/", auth)
		if err != nil {
			log.Fatal("Could not fetch root item of filesystem!:", err)
		}
		root.auth = &auth
		c.root = root
	}
	last := c.root

	// from the root directory, traverse the chain of items till we reach our
	// target key
	key = strings.TrimSuffix(key, "/")
	split := strings.Split(key, "/")[1:] // omit leading "/"
	for i := 0; i < len(split); i++ {
		item, exists := last.Children[split[i]]
		if !exists {
			if auth.AccessToken == "" {
				return last, errors.New("Auth was empty and \"" +
					filepath.Join(last.Path(), split[i]) +
					"\" was not in cache. Could not fetch item as a result.")
			}

			// we have an auth token and can try to fetch an item's children
			children, err := last.GetChildren(auth)
			if err != nil {
				return last, err
			}
			item, exists = children[split[i]]
			if !exists {
				// this time, we know the key *really* doesn't exist
				return nil, errors.New(filepath.Join(last.Path(), split[i]) + " does not exist.")
			}
		}
		last = item
	}
	return last, nil
}

// Delete an item from the cache
func (c *ItemCache) Delete(key string) {
	// Uses empty auth, since we actually don't want to waste time fetching
	// items that are only being fetched so they can be deleted.
	parent, err := c.Get(filepath.Dir(key), Auth{})
	if err == nil {
		delete(parent.Children, filepath.Base(key))
	}
}

// Insert lets us manually insert an item to the cache (like if it was created
// locally). Overwrites a cached item if present.
func (c *ItemCache) Insert(resource string, auth Auth, item *DriveItem) error {
	parent, err := c.Get(filepath.Dir(resource), auth)
	if err != nil {
		return err
	}
	parent.Children[filepath.Base(resource)] = item
	return nil
}
