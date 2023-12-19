import java.io.PrintWriter;
import java.io.StringWriter;
import java.util.ArrayList;
import java.util.List;

public class Grader {
    private static final String ASSIGNMENT_NAME = "java";

    public static void main(String[] srgs) {
        // Score all the assignment tasks.
        List<QuestionScore> scores = new ArrayList<QuestionScore>();
        scores.add(testAdd());

        // Output JSON.
        outputJSON(scores);
    }

    private static void outputJSON(List<QuestionScore> scores) {
        List<String> questions = new ArrayList<String>();
        for (QuestionScore score : scores) {
            questions.add(score.toJSON());
        }

        String json = String.format(
                "{\"name\": \"%s\", \"questions\": [%s]}",
                ASSIGNMENT_NAME, String.join(", ", questions));

        System.out.println(json);
    }

    private static QuestionScore testAdd() {
        QuestionScore score = new QuestionScore("Task 1: add()", 10);

        AddTestCase[] testCases = new AddTestCase[]{
            new AddTestCase(1, 2, 3, "basic"),
            new AddTestCase(0, 2, 2, "one zero"),
            new AddTestCase(0, 0, 0, "all zero"),
            new AddTestCase(-1, 2, 1, "one negative"),
            new AddTestCase(-1, -2, -3, "all negative"),
        };

        try {
            testAddTestCases(testCases, score);
        } catch (Exception ex) {
            String message = "Failed to score add(), caught exception: " + ex + "\n";
            message += "---\n";
            message += getStringStackTrace(ex) + "\n";
            message += "---\n";

            score.score = 0;
            score.message = message;
        }

        return score;
    }

    private static void testAddTestCases(AddTestCase[] testCases, QuestionScore score) {
        score.score = score.max_points;

        List<String> messages = new ArrayList<String>();

        for (AddTestCase testCase : testCases) {
            int result = Assignment.add(testCase.a, testCase.b);
            if (result != testCase.expected) {
                messages.add(String.format("Missed test case '%s'.", testCase.feedback));
                score.score -= 2;
            }
        }

        score.message = String.join("\n", messages);
    }

    private static String getStringStackTrace(Throwable ex) {
        StringWriter writer = new StringWriter();
        ex.printStackTrace(new PrintWriter(writer));
        return writer.toString();
    }

    public static class AddTestCase {
        public int a;
        public int b;
        public int expected;
        public String feedback;

        public AddTestCase(int a, int b, int expected, String feedback) {
            this.a = a;
            this.b = b;
            this.expected = expected;
            this.feedback = feedback;
        }
    }

    public static class QuestionScore {
        public String name;
        public int max_points;
        public int score;
        public String message;

        public QuestionScore(String name, int max_points) {
            this.name = name;
            this.max_points = max_points;
        }

        public String toJSON() {
            String escapedMessage = message.replace("\"", "\\\"").replace("\n", "\\n").replace("\t", "\\t");
            return String.format(
                    "{\"name\": \"%s\", \"max_points\": %d, \"score\": %d, \"message\": \"%s\"}",
                    name, max_points, score, escapedMessage);
        }
    }
}
