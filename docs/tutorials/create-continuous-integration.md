# Tutorial: Creating Continuous Integration

In this tutorial we are going to cover suggestions for [Continuous Integration](https://en.wikipedia.org/wiki/Continuous_integration) (CI).
These suggestions focus on testing the graders for the autograder, but more CI may be necessary.

Extending `my-first-course` that we originally wrote during the [create-a-course.md](create-a-course.md),
we added a new implementation called `final-with-ci`.
We will reference `final-with-ci` as a guide.

## Writing Test Submissions

Within each assignment directory, include a `test-submissions` directory that contains various possible "submissions" for the assignments.
Note that a "submission" does not need to score 100% or even be complete,
it just represents something that a student may submit.

Two submissions should always exist: `solution` and `not_implemented`.
`solution` is an ideal 100% solution,
such as the [solution submission](resources/my-first-course/final-with-ci/assignment-01/sample-submissions/solution/submission.py).
`not_implemented` is the same version of the code that is initially given to the students,
such as the [not-implemented submission](resources/my-first-course/final-with-ci/assignment-01/sample-submissions/not-implemented/submission.py).
These names are not special, so feel free to change them if needed.

Additionally, a submission that fails to build can be helpful,
such as the [syntax-error submission](resources/my-first-course/final-with-ci/assignment-01/sample-submissions/syntax-error/submission.py).

As the course progresses, it is **VITALLY** important that course developers add to these submissions.
Every time a mistake is found in the grader,
the first thing a course developer should do is create a submission that exposes the bug.
Then, course developers should fix the bug and watch the new test case (test submission) pass.

## Running the Grader

We recommend writing a script that simplifies running a grader for an assignment.
In `final-with-ci`, we provide a [run_grader.sh](resources/my-first-course/final-with-ci/assignment-01/run_grader.sh) to run the grader on a sample submission.
For example, run the grader with the `not-implemented` submission as follows (within the `final-with-ci/assignment-01` directory):
```
./run_grader.sh sample-submissions/not-implemented
```

You should see an output similar to:
```
Autograder transcript for assignment: Assignment1.
Grading started at 2025-03-14 12:58 and ended at 2025-03-14 12:58.
Question 1: Add: 1 / 10
   Add() did not return a correct result on test case 'zeros'.
   Add() did not return a correct result on test case 'basic'.
   Add() did not return a correct result on test case 'negative'.

Total: 1 / 10
```

These scripts help course developers quickly verify the output on sample submissions.

## Continuous Integration

After setting up sample submissions and scripts to run the graders,
we can automatically verify the graders produce the correct results.

To do so, each sample submission requires a `test-submission.json`, such this [example from not-implemented](resources/my-first-course/final-with-ci/assignment-01/sample-submissions/not-implemented/test-submission.json).
This allows CI to capture the output of the grader and compare the results with the expected results.

With these parts in place, we can write the [CI script to check submissions](resources/my-first-course/final-with-ci/.ci/check_submissions.sh).
Run the script as follows (from the `final-with-ci` directory):
```
./.ci/check_submissions.sh
```

If everything works correctly, the script will output something similar to the following:
```
Checking assignment: '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/.ci/../assignment-01'.
Testing assignment '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/.ci/../assignment-01/assignment.json' and submission '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/assignment-01/sample-submissions/solution/test-submission.json'.
Testing assignment '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/.ci/../assignment-01/assignment.json' and submission '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/assignment-01/sample-submissions/syntax-error/test-submission.json'.
Testing assignment '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/.ci/../assignment-01/assignment.json' and submission '/home/<user>/autograder-server/docs/tutorials/resources/my-first-course/final-with-ci/assignment-01/sample-submissions/not-implemented/test-submission.json'.
Encountered 0 error(s) while testing 3 submissions.
Success
No issues found!
```

If not, the script output the test submissions that did not produce the expected results.

## Workflow

Congratulations! Now that you have CI to check your graders against sample submissions,
add it to [Github Actions](https://docs.github.com/en/actions/writing-workflows) or an equivalent workflow.
