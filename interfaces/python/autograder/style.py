import contextlib
import os
import sys

from flake8.api import legacy as flake8

import autograder.code
import autograder.question
import autograder.utils

# For codes, see:
# flake8: https://flake8.pycqa.org/en/latest/user/error-codes.html
# pycodestyle: https://pycodestyle.pycqa.org/en/latest/intro.html#error-codes
STYLE_OPTIONS = {
    'max_line_length': 100,
    'max_doc_length': 100,
    'select': 'E,W,F',
    'show_source': True,
    'color': 'never',
    'ignore': [
        # Don't force continuation line alignment.
        'E128',

        # Allow spaces around parameter/keyword equals.
        'E251',

        # Don't force two spaces between functions/class.
        # We only want one space.
        'E302',
        'E305',

        # Allow lambdas to be assigned into a local variable.
        'E731',

        # Ignore missing newlines at the end of the file.
        # This is a result of striping incoming code.
        'W292',

        # Do not enforce spaces in blank lines.
        # This should typically be enforced,
        # but because code from notebooks is hard to give a nice line number for
        # students have difficulty finding where the issue is.
        'W293',

        # PEP-8 recommends breaking a line before a binary operator.
        # This was a recent reversal of idiomatic Python.
        # W503 enforces the old style, while W504 enforces the new style.
        'W503',
    ]
}

class Style(autograder.question.Question):
    """
    A question that can be added to assignments that checks style.
    """

    def __init__(self, path, max_points = 5, replacement_name = 'assignment.py'):
        super().__init__("Style", max_points)
        self._path = path
        self._replacement_name = replacement_name

    def score_question(self, *args, **kwargs):
        error_count, style_output = check_path(self._path,
                replace_output_path = self._replacement_name)

        if (error_count == 0):
            self.add_message("Style is clean!")
            self.full_credit()
        else:
            self.add_message(("Code has %d style issues (shown below)."
                    + " Note that line numbers will be offset because of iPython notebooks.")
                    % (error_count))
            self.score = max(0, self.max_points - error_count)

            self.add_message("--- Style Output BEGIN ---")
            for (path, lines) in style_output:
                self.add_message("\nStyle Issues for: '%s':" % path)
                self.add_message('---')
                self.add_message("\n".join(lines))
                self.add_message("---\n")
            self.add_message("--- Style Output END ---")

def check_path(path, **kwargs):
    return check_paths([path], **kwargs)

def check_paths(paths, **kwargs):
    """
    Check the style of all the listed paths (recursivley) and return a two-item tuple of:
        - The total number of style violations.
        - A list of two-item tules of:
            - The path to a file with style violations.
            - A list of strings that describe the style issues.
    """

    total_count = 0
    # [(path, lines), ...]
    total_lines = []

    for path in paths:
        if (os.path.isfile(path)):
            if (os.path.splitext(path)[1] not in autograder.code.ALLOWED_EXTENSIONS):
                continue

            count, lines = _check_file(path, **kwargs)
            lines = [(path, lines)]
        else:
            dirents = [os.path.join(path, dirent) for dirent in os.listdir(path)]
            count, lines = check_paths(dirents, **kwargs)

        if (count > 0):
            total_count += count
            total_lines += lines

    return total_count, total_lines

def _check_file(path, replace_output_path = None):
    """
    Check the style of a file and return a two-item tuple of:
        - The number of style violations.
        - A list of strings that describe the style issues.
    """

    if (not os.path.isfile(path)):
        raise ValueError("Can only check style on a file, got a directory: '%s'." % (path))

    cleanup_paths = []

    if (path.endswith('.py')):
        pass
    elif (path.endswith('.ipynb')):
        contents = autograder.code.extract_notebook_code(path)

        temp_path = autograder.utils.get_temp_path(prefix = 'style_', suffix = '_notebook')
        cleanup_paths.append(temp_path)
        with open(temp_path, 'w') as file:
            file.write(contents)

        if (replace_output_path is None):
            replace_output_path = path

        path = temp_path
    else:
        raise ValueError("Can only check style on .py or .ipynb files, got '%s'." % (path))

    path = os.path.realpath(path)

    output_path = autograder.utils.get_temp_path(prefix = 'style_', suffix = '_output')
    cleanup_paths.append(output_path)

    # argparse (used by flake8) will look for a program name on sys.argv[0].
    if (len(sys.argv) == 0):
        sys.argv = ['']

    style_guide = flake8.get_style_guide(**STYLE_OPTIONS)

    with open(output_path, 'w') as file:
        with contextlib.redirect_stdout(file):
            report = style_guide.check_files([path])

    with open(output_path, 'r') as file:
        lines = file.readlines()

    lines = [line.rstrip() for line in lines]

    if (replace_output_path is not None):
        lines = [line.replace(path, replace_output_path) for line in lines]

    for path in cleanup_paths:
        autograder.utils.remove_dirent(path)

    return (report._application.result_count, lines)
