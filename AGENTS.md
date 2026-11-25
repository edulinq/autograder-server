# Agent Instructions

You are assisting a developer who is responsible for everything you produce.
Reviewer time is this project's scarcest resource.
If you cannot follow a rule here, stop and tell the developer.
If you cannot ask, leave that step undone and report it.

## Read First

 - [EduLinq Contributing Guide](https://github.com/edulinq/.github/blob/main/.github/CONTRIBUTING.md) (not in this repo, fetch it),
   especially the sections on PRs, new contributors, and AI tools.
 - [Development Notes](docs/development.md) and the [package hierarchy](docs/packages.md).
 - The full discussion of any issue being addressed, including related PRs (especially rejected ones).

## Before Coding

 - Tell the developer the current behavior, the intended behavior, and the root cause.
   Flag anything you are unsure of before implementing.
 - Fix the root cause at the layer that owns it, not where the symptom appears.
 - Do not invent features.
   If no issue asks for the change (and it is not a clear bug), suggest that the developer ask first.
 - Branch from an up-to-date `main`.
 - Read sibling files (including tests) and match them.

## Evidence of Correctness

 - Initialize the submodule (`git submodule update --init`), the project does not build without it.
   Do not commit changes to it.
 - Run the full checks, see [Running Tests](README.md#running-tests).
   Without Docker, use `AUTOGRADER__DOCKER__DISABLE=true ./.ci/run_all.sh`.
 - Use the Go version from the [CI configuration](.github/workflows/main.yml), since some expected values depend on it.
 - If a test fails, check whether it also fails on `main` (use a separate worktree).
   Never change an expected value just to make a test pass.
 - When fixing a bug, write the test first and watch it fail before fixing the bug.
 - Test the underlying behavior, not just the endpoint that exposes it.
 - Tests should be deterministic (no luck or repeated runs) and match exact output.
 - Prefer adding cases to an existing test over adding new test functions.
   Cover each distinct case once, redundant tests are a maintenance burden.
 - If a test is impractical (e.g., some race conditions), explain the exact sequence that fails
   and report the results of `go test -race` on the relevant packages.
 - Only claim what you actually ran, and report skipped tests.

## API Changes

 - Follow the [procedure for adding/editing an endpoint](docs/development.md#development-procedure-for-addingediting-an-api-endpoint), in order.
   The Python interface comes after the server PR is conditionally accepted.
 - When the API description test fails, show the diff to the developer.
   Only run `scripts/update_api_description.sh` once they confirm that the change is intended.
 - For externally visible changes, include a minimal before/after example in the PR description.

## Style

Run `scripts/format_code.sh`, then check for what it will not catch:

 - Separate adjacent conditionals with a blank line.
 - Put parentheses around the operands of binary logical operators, e.g., `(a == nil) || b.isValid()`.
 - Put spaces around binary operators, restructure the expression if the formatter removes them.
 - Name receivers `this`, and avoid one letter variable names (except `i`, `j`, and generic types).
 - Keep lists of routes in lexicographic order.
 - Comments are complete sentences or Title Case.
   They describe the code as it is now, not what it used to be or the bug that was fixed.
 - Do not remove existing comments that are still accurate.

## Before Opening a PR

 1. Read the full diff against `main`.
    Remove anything out of scope, including unrelated files and commits.
 2. If you can, have an independent subagent (with fresh context) review the commits as a skeptical maintainer.
    Give it the issue, the diff, this file, and the contributing guide.
    Ask if the change is wanted, correct, minimal, tested, and consistent with existing code.
    Fix what it finds, and keep the review private.
 3. Do not open a PR, post a comment, or request a review without the developer's explicit approval.
 4. Keep the PR description short: what changed, why, and the evidence.
    Please end it with one line naming the harness and model used, e.g., `Assisted by: <harness>, <model>`.
    Do the same for any comment you substantially wrote.

## After a Review

 - Follow the contributing guide's rules for addressing comments and requesting another pass.
 - Do not post summaries of your changes.
 - Do not ask the reviewer to do the design work.
   Help the developer reason through the options and make an attempt.
