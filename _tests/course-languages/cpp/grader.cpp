#include <iostream>
#include <string>
#include <vector>

#include "json.hpp"

#include "assignment.h"

const std::string ASSIGNMENT_NAME = "cpp-simple";
const int INDENT = 4;

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

        nlohmann::json toJSON();
};

nlohmann::json QuestionScore::toJSON() {
    nlohmann::json value = {
        {"name", name},
        {"max_points", max_points},
        {"score", score},
        {"message", message},
    };

    return value;
}

void testAddTestCases(std::vector<AddTestCase> &testCases, QuestionScore &score) {
    score.score = score.max_points;

    for (int i = 0; i < testCases.size(); i++) {
        AddTestCase testCase = testCases.at(i);

        int result = add(testCase.a, testCase.b);
        if (result != testCase.expected) {
            score.message.append("Missed test case '" + testCase.feedback + "'.\n");
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
    nlohmann::json questions = {
        testAdd().toJSON(),
    };

    nlohmann::json output = {
        {"name", ASSIGNMENT_NAME},
        {"questions", questions},
    };

    std::cout << output.dump(INDENT) << std::endl;

    return 0;
}
