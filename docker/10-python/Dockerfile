FROM autograder.base

# TODO(eriq): Install from pypi.
COPY autograder-*.tar.gz /autograder/build/
RUN pip3 install /autograder/build/autograder-*.tar.gz

COPY scripts /autograder/scripts

# TODO(eriq): Entrypoint?

WORKDIR /autograder
CMD ["/autograder/scripts/run_grader.sh"]
