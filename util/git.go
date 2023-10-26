package util

import (
    "bytes"
    "fmt"
    "os/exec"

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

func GitGetCommitHash(path string) (string, error) {
    repo, err := GitGetRepo(path);
    if (err != nil) {
        return "", fmt.Errorf("Unable to open repo (%s'): '%w'.", path, err);
    }

    head, err := repo.Head()
    if (err != nil) {
        return "", fmt.Errorf("Unable to get repo's head (%s'): '%w'.", path, err);
    }

    return head.Hash().String(), nil;
}

// Due to a long standing issue in go-get, this operation is slow and should generally be avoided.
// https://github.com/go-git/go-git/issues/181
func GitRepoIsDirty(path string) (bool, error) {
    repo, err := GitGetRepo(path);
    if (err != nil) {
        return false, fmt.Errorf("Unable to open repo (%s'): '%w'.", path, err);
    }

    worktree, err := repo.Worktree();
    if (err != nil) {
        return false, fmt.Errorf("Unable to get repo's working tree (%s'): '%w'.", path, err);
    }

    status, err := worktree.Status();
    if (err != nil) {
        return false, fmt.Errorf("Unable to get repo's status (%s'): '%w'.", path, err);
    }

    return !status.IsClean(), nil;
}

// A fast, hacky version of checking for a dirty repo.
// This function will call out to the command-line git (which it assumes is installed) to check the status.
func GitRepoIsDirtyHack(path string) (bool, error) {
    var stdout bytes.Buffer;

    cmd := exec.Command("git", "status", "--porcelain");
    cmd.Stdout = &stdout;

    err := cmd.Run();
    if (err != nil) {
        return false, fmt.Errorf("Failed to run command-line git: '%w'.", err);
    }

    return (stdout.Len() > 0), nil;
}
