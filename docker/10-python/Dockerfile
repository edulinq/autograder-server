FROM autograder.base

# TODO(eriq): Install from pypi.
COPY autograder-*.tar.gz /autograder/build/
RUN pip3 install /autograder/build/autograder-*.tar.gz

WORKDIR /autograder/work
CMD ["python3", "-m", "autograder.cli.grade", "--assignment", "/autograder/work/assignment.py", "--submission", "/autograder/submission"]
