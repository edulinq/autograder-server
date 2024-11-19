from autograder.assignment import Assignment
from autograder.question import Question
from autograder.style import Style

class HW0(Assignment):
    def __init__(self, **kwargs):
        super().__init__(questions = [
            Q1(1),
            Q2(1),
            Style(kwargs.get('input_dir'), max_points = 0),
        ], **kwargs)

class Q1(Question):
    def score_question(self, submission):
        result = submission.__all__.function1()
        self.check_not_implemented(result)

        if (not result):
            self.fail("function1() should return True.")

        self.full_credit()

class Q2(Question):
    def score_question(self, submission):
        result = submission.__all__.function2(0)
        self.check_not_implemented(result)

        if (result != 1):
            self.fail("function2(0) should return 1.")

        self.full_credit()
