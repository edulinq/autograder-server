package util

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
)

// {AbsPath: *sync.Mutex}.
var fileLocks sync.Map;

// {AbsPath: ModTime, ...}
type FileCache map[string]int64;

// Check to see if the given files have changed since the cache was written.
// Note that this method only looks at mod time and
// does not account for deletions from directories (if specified inside |paths|).
// A specified path that does not exist will always return in true being returned.
// |quick| will be used to determine if the function should return immediately if a difference is found,
// or complete checking (and updating) the cache.
// This method is thread safe.
func HaveFilesChanges(cachePath string, paths []string, quick bool) (bool, error) {
    cachePath = MustAbs(cachePath);

    lock, _ := fileLocks.LoadOrStore(cachePath, &sync.Mutex{});
    lock.(*sync.Mutex).Lock();
    defer lock.(*sync.Mutex).Unlock();

    cache, err := loadFileCache(cachePath);
    if (err != nil) {
        return false, fmt.Errorf("Unable to get file cache '%s': '%w'.", cachePath, err);
    }

    result, err := haveFilesChanges(cache, paths, quick);
    if (err != nil) {
        return false, err;
    }

    // If chnges were found, the cache must be rewritten
    // (even if this is quick and it is just the one file).
    if (result) {
        err := ToJSONFile(cache, cachePath);
        if (err != nil) {
            return false, fmt.Errorf("Unable to save file cache '%s': '%w'.", cachePath, err);
        }
    }

    return result, nil;
}

func loadFileCache(path string) (FileCache, error) {
    if (!PathExists(path)) {
        return make(FileCache), nil;
    }

    var cache FileCache;

    err := JSONFromFile(path, &cache);
    if (err != nil) {
        return nil, err;
    }

    return cache, nil;
}

func haveFilesChanges(cache FileCache, paths []string, quick bool) (bool, error) {
    changesFound := false;

    for _, path := range paths {
        if (changesFound && quick) {
            break;
        }

        // Need abs paths for cache consistency.
        path = MustAbs(path);

        if (!PathExists(path)) {
            changesFound = true;
            // The path may not be in the cache and could result in and extra write, but that's fine.
            delete(cache, path);

            continue;
        }

        if (IsDir(path)) {
            dirHasChanges, err := handleDir(cache, path, quick);
            if (err != nil) {
                return false, err;
            }

            changesFound = changesFound || dirHasChanges;

            continue;
        }

        modTime, err := getModTime(path);
        if (err != nil) {
            return false, fmt.Errorf("Failed to get mod time for '%s': '%w'.", path, err);
        }

        if (modTime != cache[path]) {
            changesFound = true;
            cache[path] = modTime;
        }
    }

    return changesFound, nil;
}

func handleDir(cache FileCache, dir string, quick bool) (bool, error) {
    dirents, err := os.ReadDir(dir);
    if (err != nil) {
        return false, fmt.Errorf("Could not list dir '%s': '%w'.", dir, err);
    }

    paths := make([]string, 0, len(dirents));
    for _, dirent := range dirents {
        paths = append(paths, filepath.Join(dir, dirent.Name()));
    }

    return haveFilesChanges(cache, paths, quick);
}

func getModTime(path string) (int64, error) {
    file, err := os.Open(path);
    if (err != nil) {
        return 0, err;
    }
    defer file.Close();

    stat, err := file.Stat();
    if (err != nil) {
        return 0, err;
    }

    return stat.ModTime().Unix(), nil;
}
