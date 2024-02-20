package util

import (
    "bytes"
    "fmt"
    "os/exec"

    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/transport"
    "github.com/go-git/go-git/v5/plumbing/transport/http"

    "github.com/edulinq/autograder/log"
)

func GitEnsureRepo(url string, path string, update bool, ref string, user string, pass string) (*git.Repository, error) {
    if (IsFile(path)) {
        return nil, fmt.Errorf("Cannot ensure git repo, path exists and is a file: '%s'.", path);
    }

    var err error;
    var repo *git.Repository;

    var auth transport.AuthMethod = nil;
    if ((user != "") || (pass != "")) {
        auth = &http.BasicAuth{
            Username: user,
            Password: pass,
        };
    }

    if (!IsDir(path)) {
        repo, err = GitClone(url, path, auth);
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
        _, err = GitUpdateRepo(repo, auth);
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

func GitClone(url string, path string, auth transport.AuthMethod) (*git.Repository, error) {
    log.Trace("Cloning git repo.", log.NewAttr("url", url), log.NewAttr("path", path));

    options := &git.CloneOptions{
        URL: url,
        RecurseSubmodules: 3,
        Auth: auth,
    };

    repo, err := git.PlainClone(path, false, options);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to clone git repo '%s' into '%s': '%w'.", url, path, err);
    }

    err = setupTrackingBranches(repo, auth);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to setup tracking branches on repo '%s': '%w'.", url, err);
    }

    return repo, nil;
}

// Go through all remotes and setup any tracked branches.
func setupTrackingBranches(repo *git.Repository, auth transport.AuthMethod) error {
    remotes, err := repo.Remotes();
    if (err != nil) {
        return fmt.Errorf("Failed to get remotes: '%w'.", err);
    }

    options := &git.ListOptions{
        Auth: auth,
    };

    for _, remote := range remotes {
        refs, err := remote.List(options);
        if (err != nil) {
            return fmt.Errorf("Failed to get objects for remote '%s': '%w'.", remote, err);
        }

        for _, ref := range refs {
            // Skip symbolic references.
            if (ref.Type() != plumbing.HashReference) {
                continue;
            }

            // Skip already tracking branches.
            if (ref.Target() != "") {
                continue;
            }

            // Only do things that look like branches.
            if (!ref.Name().IsBranch()) {
                continue;
            }

            // Skip branches that already exist.
            existingBranch, _ := repo.Branch(ref.Name().Short());
            if (existingBranch != nil) {
                continue;
            }

            branch := config.Branch{
                Name: ref.Name().Short(),
                Remote: remote.Config().Name,
                Merge: ref.Name(),
            }

            err = repo.CreateBranch(&branch);
            if (err != nil) {
                return fmt.Errorf("Failed to create local tracking branch of '%s' for remote '%s': '%w'.",
                        ref.Name(), remote, err);
            }

            remoteReferenceName := plumbing.NewRemoteReferenceName("origin", ref.Name().Short());
            newReference := plumbing.NewSymbolicReference(ref.Name() , remoteReferenceName)
            err = repo.Storer.SetReference(newReference)
            if (err != nil) {
                return fmt.Errorf("Failed to create tracking link of '%s' for remote '%s': '%w'.",
                        ref.Name(), remote, err);
            }
        }
    }

    return nil;
}

func GitCheckoutRepo(repo *git.Repository, ref string) error {
    log.Trace("Checking out git repo.", log.NewAttr("ref", ref));

    tree, err := repo.Worktree();
    if (err != nil) {
        return err;
    }

    options, err := getCheckOptions(repo, ref);
    if (err != nil) {
        return err;
    }

    options.Force = true;

    err = tree.Checkout(options);
    if (err != nil) {
        return fmt.Errorf("Failed to checkout ref '%s': '%w'.", ref, err);
    }

    return nil;
}

// Resolve a refernce into checkout options.
func getCheckOptions(repo *git.Repository, ref string) (*git.CheckoutOptions, error) {
    options := &git.CheckoutOptions{}

    branch, err := repo.Branch(ref);
    if (err == nil) {
        options.Branch = plumbing.NewBranchReferenceName(branch.Name);
        return options, nil;
    }

    if (err != git.ErrBranchNotFound) {
        return nil, fmt.Errorf("Failed to check if ref is a branch: '%w'.", err);
    }

    tag, err := repo.Tag(ref);
    if (err == nil) {
        options.Hash = tag.Hash();
        return options, nil;
    }

    if (err != git.ErrTagNotFound) {
        return nil, fmt.Errorf("Failed to check if ref is a tag: '%w'.", err);
    }

    resolvedHash, err := repo.ResolveRevision(plumbing.Revision(ref));
    if (err != nil) {
        return nil, fmt.Errorf("Failed to resolve ref '%s': '%w'.", ref, err);
    }

    options.Hash = *resolvedHash;
    return options, nil;
}

// Return true if an update happened.
func GitUpdateRepo(repo *git.Repository, auth transport.AuthMethod) (bool, error) {
    log.Trace("Pulling git repo.");

    tree, err := repo.Worktree();
    if (err != nil) {
        return false, err;
    }

    options := &git.PullOptions{
        Auth: auth,
    };

    err = tree.Pull(options);
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
