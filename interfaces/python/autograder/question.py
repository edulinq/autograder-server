"""
A single question (test case) for an assignment.
"""

import abc
import functools
import traceback

import autograder.utils

DEFAULT_TIMEOUT_SEC = 60

class Question(object):
    """
    Questions are grade-able portions of an assignment.
    They can also be thought of as "test cases".
    Note that all scoring is in ints.
    """

    def __init__(self, name, max_points, timeout = DEFAULT_TIMEOUT_SEC):
        self.name = name

        self.max_points = max_points
        self._timeout = timeout

        # Scoring artifacts.
        self.score = 0
        self.message = ''

    def grade(self, submission, additional_data = {}, show_exceptions = False):
        """
        Invoke the scoring method using a timeout and cleanup.
        Return the score.
        """

        helper = functools.partial(self._score_helper, submission,
                additional_data = additional_data)

        try:
            success, value = autograder.utils.invoke_with_timeout(self._timeout, helper)
        except Exception:
            if (show_exceptions):
                traceback.print_exc()

            self.fail("Raised an exception: " + traceback.format_exc())
            return 0

        if (not success):
            if (value is None):
                self.fail("Timeout (%d seconds)." % (self._timeout))
            else:
                self.fail("Error during execution: " + value)

            return 0

        # Because we use the helper method, we can only get None back if there was an error.
        if (value is None):
            self.fail("Error running scoring.")
            return 0

        self.score = value[0]
        self.message = value[1]

        return self.score

    def _score_helper(self, submission, additional_data = {}):
        """
        Score the question, but make sure to return the score and message so
        multiprocessing can properly pass them back.
        """

        self.score_question(submission, **additional_data)
        return (self.score, self.message)

    def check_not_implemented(self, value):
        if (value is None):
            self.fail("None returned.")
            return True

        if (isinstance(value, type(NotImplemented))):
            self.fail("NotImplemented returned.")
            return True

        return False

    def fail(self, message):
        """
        Immediately fail this question, no partial credit.
        """

        self.score = 0
        self.message = message

    def full_credit(self):
        self.score = self.max_points

    def add_message(self, message, score = 0):
        if (self.message != ''):
            self.message += "\n"
        self.message += message

        self.score += score

    @abc.abstractmethod
    def score_question(self, submission, **kwargs):
        """
        Assign an actual score to this question.
        The implementer has full access to instance variables.
        However, only self.score and self.message can be modified.
        """

        pass

    def scoring_report(self, prefix = ''):
        """
        Get a string that represents the scoring for this question.
        """

        if ((prefix != '') and (not prefix.endswith(' '))):
            prefix += ' '

        lines = ["%s%s: %d / %d" % (prefix, self.name, self.score, self.max_points)]
        if (self.message != ''):
            for line in self.message.split("\n"):
                lines.append("   " + line)

        return "\n".join(lines)

    def __eq__(self, other):
        if (not isinstance(other, Question)):
            return False

        return (
            (self.name == other.name)
            and (self.max_points == other.max_points)
            and (self._timeout == other._timeout)
            and (self.score == other.score)
            and (self.message == other.message))

    def to_dict(self):
        """
        Convert to all simple structures that can be later converted to JSON.
        """

        return {
            'name': self.name,
            'max_points': self.max_points,
            'timeout': self._timeout,
            'score': self.score,
            'message': self.message,
        }

    @staticmethod
    def from_dict(data):
        """
        Partner to to_dict().
        Questions constructed with this will not have an implementation for score_question().
        """

        question = Question(data['name'], data['max_points'], data['timeout'])
        question.score = data['score']
        question.message = data['message']

        return question
