#include <cstdio>
#include <iostream>
#include <string>
#include <vector>

#include "assignment.h"

const char* ASSIGNMENT_NAME = "cpp";
const int BUFFER_LENGTH = 1024;

struct AddTestCase {
    int a;
    int b;
    int expected;
    std::string feedback;
};

class QuestionScore {
    public:
        std::string name;
        int max_points;
        int score;
        std::string message;

        QuestionScore(std::string a_name, int a_max_points)
            : name(a_name), max_points(a_max_points)
        {}

        char* toJSON();
};

char* QuestionScore::toJSON() {
    char* output = new char[BUFFER_LENGTH];
    snprintf(output, BUFFER_LENGTH, R"({"name": "%s", "max_points": %d, "score": %d, "message": "%s"})", &(name[0]), max_points, score, &(message[0]));
    return output;
}

void testAddTestCases(std::vector<AddTestCase> &testCases, QuestionScore &score) {
    score.score = score.max_points;

    for (int i = 0; i < testCases.size(); i++) {
        AddTestCase testCase = testCases.at(i);

        int result = add(testCase.a, testCase.b);
        if (result != testCase.expected) {
            score.message.append("Missed test case '" + testCase.feedback + "'. ");
            score.score -= 2;
        }
    }
}

QuestionScore testAdd() {
    QuestionScore score("Task 1: add()", 10);

    std::vector<AddTestCase> testCases = {
        AddTestCase{1, 2, 3, "basic"},
        AddTestCase{0, 2, 2, "one zero"},
        AddTestCase{0, 0, 0, "all zero"},
        AddTestCase{-1, 2, 1, "one negative"},
        AddTestCase{-1, -2, -3, "all negative"},
    };

    try {
        testAddTestCases(testCases, score);
    } catch (const std::exception& ex) {
        score.score = 0;
        score.message.append("Failed to score add(), caught exception: ").append(ex.what()).append("\n");
    }

    return score;
}

int main() {
    char* question = testAdd().toJSON();

    char* output = new char[BUFFER_LENGTH];
    snprintf(output, BUFFER_LENGTH, R"({"name": "%s", "questions": [%s]})", ASSIGNMENT_NAME, question);

    std::cout << output << std::endl;

    return 0;
}
