package util

import (
    "fmt"

    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
)

func GitEnsureRepo(url string, path string, update bool, ref string) (*git.Repository, error) {
    if (IsFile(path)) {
        return nil, fmt.Errorf("Cannot ensure git repo, path exists and is a file: '%s'.", path);
    }

    var err error;
    var repo *git.Repository;

    if (!IsDir(path)) {
        repo, err = GitClone(url, path);
        if (err != nil) {
            return nil, err;
        }
    }

    if (repo == nil) {
        repo, err = GitGetRepo(path);
        if (err != nil) {
            return nil, err;
        }
    }

    if (update) {
        _, err = GitUpdateRepo(repo);
        if (err != nil) {
            return nil, err;
        }
    }

    if (ref != "") {
        err = GitCheckoutRepo(repo, ref);
        if (err != nil) {
            return nil, err;
        }
    }

    return repo, nil;
}

func GitGetRepo(path string) (*git.Repository, error) {
    repo, err := git.PlainOpen(path);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to open repo at '%s': '%w'.", path, err);
    }

    return repo, nil;
}

func GitClone(url string, path string) (*git.Repository, error) {
    repo, err := git.PlainClone(path, false, &git.CloneOptions{
        URL: url,
    });

    if (err != nil) {
        return nil, fmt.Errorf("Failed to clone git repo '%s' into '%s': '%w'.", url, path, err);
    }

    return repo, nil;
}

func GitCheckoutRepo(repo *git.Repository, ref string) error {
    tree, err := repo.Worktree();
    if (err != nil) {
        return err;
    }

    resolvedHash, err := repo.ResolveRevision(plumbing.Revision(ref));
    if (err != nil) {
        return fmt.Errorf("Failed to resolve ref '%s': '%w'.", ref, err);
    }

    options := &git.CheckoutOptions{
        Hash: *resolvedHash,
    };

    err = tree.Checkout(options);
    if (err != nil) {
        return fmt.Errorf("Failed to checkout ref '%s': '%w'.", ref, err);
    }

    return nil;
}

// Return true if an update happened.
func GitUpdateRepo(repo *git.Repository) (bool, error) {
    tree, err := repo.Worktree();
    if (err != nil) {
        return false, err;
    }

    err = tree.Pull(&git.PullOptions{});

    if (err == nil) {
        return true, nil;
    }

    if (err == git.NoErrAlreadyUpToDate) {
        return false, nil;
    }

    return false, err;
}
