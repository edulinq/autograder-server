import atexit
import multiprocessing
import os
import shutil
import sys
import tempfile
import time
import traceback
import types
import uuid

import autograder.code

REAP_TIME_SEC = 5

ALL_SUBMISSION_KEY = '__all__'

class Mock(object):
    def __init__(self):
        self.item_history = list()
        self.attribute_history = list()
        self.call_history = list()

    def __repr__(self):
        return "Mock -- Item History: %s, Attribute History: %s, Call History: %s" % (
            str(self.item_history), str(self.attribute_history), str(self.call_history))

    def __call__(self, *args, **kwargs):
        self.call_history.append((args, kwargs))
        return self

    def __getitem__(self, name):
        self.item_history.append(name)
        return self

    def __getattr__(self, name):
        self.attribute_history.append(name)
        return self

def _invoke_helper(result, function):
    value = None
    error = None

    try:
        value = function()
    except Exception as ex:
        error = (ex, traceback.format_exc())

    sys.stdout.flush()

    result.put((value, error))
    result.close()

# Return: (success, function return value)
# On timeout, success will be false and the value will be None.
# On error, success will be false and value will be the string stacktrace.
# On successful completion, success will be true and value may be None (if nothing was returned).
def invoke_with_timeout(timeout, function):
    if (not sys.platform.startswith('linux')):
        # Mac and Windows have some pickling issues with multiprocessing.
        # Just run them without a timeout.
        # Any autograder will be run on a Linux machine and will be safe.
        start_time = time.time()
        value = function()
        runtime = time.time() - start_time

        if (runtime > timeout):
            return (False, None)

        return (True, value)

    result = multiprocessing.Queue(1)

    # Note that we use processes instead of threads so they can be more completely killed.
    process = multiprocessing.Process(target = _invoke_helper, args = (result, function))
    process.start()

    # Wait for at most the timeout.
    process.join(timeout)

    # Check to see if the process is still running.
    if process.is_alive():
        # Kill the long-running process.
        process.terminate()

        # Try to reap the process once before just giving up on it.
        process.join(REAP_TIME_SEC)

        return (False, None)

    # Check to see if the process explicitly existed (like via sys.exit()).
    if (result.empty()):
        return (False, 'Code explicitly exited (like via sys.exit()).')

    value, error = result.get()

    if (error is not None):
        exception, stacktrace = error
        return (False, stacktrace)

    return (True, value)

def prepare_submission(path, raise_on_collision = False):
    """
    Get a submission from a path, prepare it for grading,
    and return a submission namespace that contains all parsed entities.

    Directories will be fully recursively descended and each python or notebook will be prepared.

    Entries that begin with two underscores will not be included.

    The following things will be done to each file:
        1) The code will be parsed and sanitized.
        2) The sanitized code will be imported as a module into a namespace.
        3) The namespace will be made available in the returned namespace
            in a structure that matches its package structure.
            E.g. "./a/b/c.py" will be available from `prepare_submission('.').a.b.c`,
            and "foo.py" will be available from `prepare_submission('foo.py').foo`.
        4) All entries in the module will be put in the ALL_SUBMISSION_KEY attribute of the
            returned namespace.
            If raise_on_collision is True, an error will be raised if a key already exists.
    """

    submission = {}

    if (os.path.isfile(path)):
        _prepare_submission_file(submission, path, [], raise_on_collision)
    else:
        _prepare_submission_dir(submission, path, [], raise_on_collision)

    return _dict_to_namespace(submission)

def _prepare_submission_dir(submission, path, prefix, raise_on_collision):
    if (not os.path.isdir(path)):
        raise ValueError("Preparation target must be a dir: '%s'." % path)

    for dirent in os.listdir(path):
        dirent_path = os.path.join(path, dirent)

        if (os.path.isfile(dirent_path)):
            if (os.path.splitext(dirent)[1] not in autograder.code.ALLOWED_EXTENSIONS):
                continue

            _prepare_submission_file(submission, dirent_path, prefix, raise_on_collision)
        else:
            _prepare_submission_dir(submission, dirent_path, prefix + [dirent], raise_on_collision)

def _prepare_submission_file(submission, path, prefix, raise_on_collision):
    if (not os.path.isfile(path)):
        raise ValueError("Preparation target must be a file: '%s'." % path)

    basename = os.path.splitext(os.path.basename(path))[0]

    defs = autograder.code.sanitize_and_import_path(path, as_dict = True)

    for (name, value) in defs.items():
        if (name.startswith('__')):
            continue

        if ((name in submission) and (raise_on_collision)):
            raise ValueError("Name collision ('%s') when importing all keys for '%s'." % (
                name, path))

        if (ALL_SUBMISSION_KEY not in submission):
            submission[ALL_SUBMISSION_KEY] = {}

        submission[ALL_SUBMISSION_KEY][name] = defs[name]

    # Place the defs into a structure that matches the path.
    context = submission
    for part in prefix + [basename]:
        if ((part in context) and (not isinstance(context[part], dict))):
            raise ValueError("Name collision ('%s') when importing all keys for '%s'." % (
                part, path))

        if (part not in context):
            context[part] = {}

        context = context[part]

    context.update(defs)

def _dict_to_namespace(root):
    if (not isinstance(root, dict)):
        return root

    for (key, value) in root.items():
        root[key] = _dict_to_namespace(value)

    return types.SimpleNamespace(**root)

def get_temp_path(prefix = '', suffix = '', rm = True):
    """
    Get a path to a valid temp dirent.
    If rm is True, then the dirent will be attempted to be deleted on exit
    (no error will occur if the path is not there).
    """

    path = None
    while ((path is None) or os.path.exists(path)):
        path = os.path.join(tempfile.gettempdir(), prefix + str(uuid.uuid4()) + suffix)

    if (rm):
        atexit.register(remove_dirent, path)

    return path

def remove_dirent(path):
    if (not os.path.exists(path)):
        return

    if (os.path.isfile(path) or os.path.islink(path)):
        os.remove(path)
    elif (os.path.isdir(path)):
        shutil.rmtree(path)
    else:
        raise ValueError("Unknown type of dirent: '%s'." % (path))
