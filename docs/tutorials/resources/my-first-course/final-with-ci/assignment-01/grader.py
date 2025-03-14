#!/usr/bin/env python3

"""
Grader for a sample assignment.

We expect the student's code to provide the add() function,
which should take two ints and return the sum of those two ints.

This grader can be invoked using the autograder Python interface (e.g., `autograder.cli.grading.grade-dir`),
or directly:
```
./grader.py -s my-submission-directory
```
"""

import sys

import autograder.assignment
import autograder.cmd.gradeassignment
import autograder.question

# Create a class to represent the assignment.
class Assignment1(autograder.assignment.Assignment):
    def __init__(self, **kwargs):
        super().__init__(questions = [
            Add(10, "Question 1: Add"),
        ], **kwargs)

# Create a question for each thing we want to test.
class Add(autograder.question.Question):
    # Make a method to score the student's submission (provided in the |submission| argument).
    def score_question(self, submission):
        # Setup test cases.
        # [(a, b, case name), ...]
        test_cases = [
            (0, 0, 'zeros'),
            (1, 1, 'basic'),
            (-1, 1, 'negative'),
        ]

        # It can be useful to set the deduction automatically.
        # Then if you add more test cases, you don't need to adjust point values.
        deduction = -(max(1, self.max_points // len(test_cases)))

        # Start everyone with full credit (you can also do the opposite).
        self.full_credit()

        for (a, b, name) in test_cases:
            # Reference the student's function using the __all__ lookup (it can also be referenced by package name).
            actual = submission.__all__.add(a, b)
            expected = a + b

            if (actual != expected):
                self.add_message("Add() did not return a correct result on test case '%s'." % (name), add_score = deduction)

def main():
    # The autograder library provides a default CLI for us.
    return autograder.cmd.gradeassignment.main()

if (__name__ == '__main__'):
    sys.exit(main())
