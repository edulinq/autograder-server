package util

import (
    "fmt"
    "sync"
)

// {AbsPath: *sync.Mutex}.
var cacheLocks sync.Map;

// Put a new value in the cache and return any previous value.
// Return: (old value, old value exists, error).
func CachePut(cachePath string, key string, value any) (any, bool, error) {
    cachePath = MustAbs(cachePath);

    lock, _ := fileLocks.LoadOrStore(cachePath, &sync.Mutex{});
    lock.(*sync.Mutex).Lock();
    defer lock.(*sync.Mutex).Unlock();

    cache, err := loadCahce(cachePath);
    if (err != nil) {
        return nil, false, fmt.Errorf("Unable to get cache '%s': '%w'.", cachePath, err);
    }

    oldValue, ok := cache[key];

    cache[key] = value;

    err = ToJSONFile(cache, cachePath);
    if (err != nil) {
        return nil, false, fmt.Errorf("Unable to save cache '%s': '%w'.", cachePath, err);
    }

    return oldValue, ok, nil;
}

// Get a value from the cache.
// Return: (value, value exists, error).
func CacheFetch(cachePath string, key string) (any, bool, error) {
    cachePath = MustAbs(cachePath);

    lock, _ := fileLocks.LoadOrStore(cachePath, &sync.Mutex{});
    lock.(*sync.Mutex).Lock();
    defer lock.(*sync.Mutex).Unlock();

    cache, err := loadCahce(cachePath);
    if (err != nil) {
        return nil, false, fmt.Errorf("Unable to get cache '%s': '%w'.", cachePath, err);
    }

    value, ok := cache[key];
    return value, ok, nil;
}

func loadCahce(path string) (map[string]any, error) {
    if (!PathExists(path)) {
        return make(map[string]any), nil;
    }

    var cache map[string]any;

    err := JSONFromFile(path, &cache);
    if (err != nil) {
        return nil, err;
    }

    return cache, nil;
}
