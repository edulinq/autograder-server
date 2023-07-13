import os
import unittest

import autograder.code

THIS_DIR = os.path.abspath(os.path.dirname(os.path.realpath(__file__)))
DATA_DIR = os.path.join(THIS_DIR, "data")

class TestImport(unittest.TestCase):
    """
    Test the utilities for importing code.
    """

    def test_extract_base(self):
        for ext in ['py', 'ipynb']:
            path = os.path.join(DATA_DIR, 'simple.' + ext)
            source_code = autograder.code.extract_code(path)
            self.assertEqual(source_code, "SOME_CONSTANT = 1")

    def test_import_base(self):
        for ext in ['py', 'ipynb']:
            # The code should be executed, causing some_int to become -1.
            path = os.path.join(DATA_DIR, 'base.' + ext)
            module = autograder.code.import_path(path)

            self.assertEqual(module.SOME_CONSTANT, 1)
            self.assertEqual(module.some_int, -1)

            path = os.path.join(DATA_DIR, 'base_with_raise.' + ext)
            try:
                module = autograder.code.import_path(path)
                self.fail("Import with raise did not fail as expected.")
            except RuntimeError:
                # Expected.
                pass

    def test_sanitize_import_base(self):
        for ext in ['py', 'ipynb']:
            for basename in ['base', 'base_with_raise']:
                path = os.path.join(DATA_DIR, basename + '.' + ext)
                module = autograder.code.sanitize_and_import_path(path)

                self.assertEqual(module.SOME_CONSTANT, 1)
                self.assertIn('random', dir(module))
                self.assertNotIn('some_int', dir(module))
                self.assertIn('some_function', dir(module))
