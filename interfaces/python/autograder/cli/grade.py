import argparse
import inspect
import json
import os
import sys

import autograder.assignment
import autograder.code
import autograder.utils

def _load_assignments(path):
    module = autograder.code.sanitize_and_import_path(path)
    assignments = []

    for name in dir(module):
        obj = getattr(module, name)

        if (not inspect.isclass(obj)):
            continue

        if (obj == autograder.assignment.Assignment):
            continue

        if (issubclass(obj, (autograder.assignment.Assignment,))):
            assignments.append(obj)

    return assignments

def _fetch_assignment(path):
    assignments = _load_assignments(path)

    if (len(assignments) == 0):
        raise ValueError(("Assignment file (%s) does not contain any instances of" +
            " autograder.assignment.Assignment.") % (path))
    elif (len(assignments) > 1):
        # TODO(eriq): Add in assignment-class to disambiguate files with multiple assignments.
        raise ValueError(("Assignment file (%s) contains more than one (%d) instances of" +
            " autograder.assignment.Assignment.") % (path, len(assignments)))

    return assignments[0]

def run(args):
    assignment_class = _fetch_assignment(args.assignment)
    submission = autograder.utils.prepare_submission(args.submission)

    assignment = assignment_class(submission_dir = args.submission)
    assignment.grade(submission)
    print(assignment.report())

    if (args.outpath is not None):
        # TEST
        print('---')
        print(args.outpath)
        print('---')

        os.makedirs(os.path.dirname(os.path.abspath(args.outpath)), exist_ok = True)
        with open(args.outpath, 'w') as file:
            json.dump(assignment.to_dict(), file, indent = 4)

    return 0

def _load_args():
    parser = argparse.ArgumentParser(description =
        "Grade an assignment with the given submission.")

    parser.add_argument('-a', '--assignment',
        action = 'store', type = str, required = True,
        help = 'The path to a Python file containing an Assignment subclass.')

    parser.add_argument('-s', '--submission',
        action = 'store', type = str, required = True,
        help = 'The path to a submission to use for grading.')

    parser.add_argument('-o', '--outpath',
        action = 'store', type = str, required = False, default = None,
        help = 'The path to a output the JSON result.')

    return parser.parse_args()

def main():
    return run(_load_args())

if (__name__ == '__main__'):
    sys.exit(main())
