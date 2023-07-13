import os
import unittest

import autograder.utils

THIS_DIR = os.path.abspath(os.path.dirname(os.path.realpath(__file__)))
DATA_DIR = os.path.join(THIS_DIR, "data")

class TestImport(unittest.TestCase):
    """
    Test the utilities for preparing submissions.
    """

    def test_simple(self):
        for ext in ['py', 'ipynb']:
            path = os.path.join(DATA_DIR, 'simple.' + ext)
            submission = autograder.utils.prepare_submission(path)

            self.assertIn('__all__', dir(submission))
            self.assertIn('SOME_CONSTANT', dir(submission.__all__))
            self.assertEquals(submission.__all__.SOME_CONSTANT, 1)

            self.assertIn('simple', dir(submission))
            self.assertIn('SOME_CONSTANT', dir(submission.simple))
            self.assertEquals(submission.simple.SOME_CONSTANT, 1)

    def test_nested(self):
        for ext in ['py', 'ipynb']:
            path = os.path.join(DATA_DIR, 'nested')
            submission = autograder.utils.prepare_submission(path)

            self.assertIn('__all__', dir(submission))
            self.assertIn('SOME_CONSTANT', dir(submission.__all__))
            self.assertEquals(submission.__all__.SOME_CONSTANT, 1)

            self.assertIn('nested1', dir(submission))
            self.assertIn('nested2', dir(submission.nested1))
            self.assertIn('nested', dir(submission.nested1.nested2))
            self.assertIn('SOME_CONSTANT', dir(submission.nested1.nested2.nested))
            self.assertEquals(submission.nested1.nested2.nested.SOME_CONSTANT, 1)
