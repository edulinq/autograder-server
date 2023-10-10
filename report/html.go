package report

import (
    "fmt"
    "html/template"
    "strings"

    "github.com/rs/zerolog/log"
)

func (this *CourseScoringReport) ToHTML() (string, error) {
    tmpl, err := template.New("course-scoring-report").Parse(courseReportTemplate);
    if (err != nil) {
        return "", fmt.Errorf("Could not parse course scoring report template: '%w'.", err);
    }

    var builder strings.Builder;
    err = tmpl.Execute(&builder, this);
    if (err != nil) {
        return "", fmt.Errorf("Failed to execute course scoring report template: '%w'.", err);
    }

    return builder.String(), nil;
}

func (this *AssignmentScoringReport) ToHTML() (string, error) {
    tmpl, err := template.New("assignment-scoring-report").Parse(assignmentReportTemplate);
    if (err != nil) {
        return "", fmt.Errorf("Could not parse assignment scoring report template: '%w'.", err);
    }

    var builder strings.Builder;
    err = tmpl.Execute(&builder, this);
    if (err != nil) {
        return "", fmt.Errorf("Failed to execute assignment scoring report template: '%w'.", err);
    }

    return builder.String(), nil;
}

func (this *AssignmentScoringReport) MustToHTML() template.HTML {
    html, err := this.ToHTML();
    if (err != nil) {
        log.Fatal().Err(err).Str("assignment", this.AssignmentName).Msg("Failed to generate HTML assignment scoring report.");
    }

    return template.HTML(html);
}

var courseReportTemplate string = `
    <div class='autograder autograder-course-scoring-report'>
        <div class='ag-header'>
            <h2>Course: {{ .CourseName }}</h2>
        </div>
        <div class='ag-body'>
            {{ range .Assignments }}
                {{ if eq .NumberOfSubmissions 0 }} {{ continue }} {{ end }}

                <div>
                    {{ .MustToHTML }}
                </div>
            {{ end }}
        </div>
    </div>
`

var assignmentReportTemplate string = `
    <style>
        .autograder-assignment-scoring-report table th,
        .autograder-assignment-scoring-report table .text {
            text-align: left;
        }

        .autograder-assignment-scoring-report table .numeric {
            text-align: right;
        }

        .autograder-assignment-scoring-report table th,
        .autograder-assignment-scoring-report table td {
            padding: 5px;
        }
    </style>

    <div class='autograder autograder-assignment-scoring-report'>
        <div class='ag-header'>
            <h2>Assignment: {{ .AssignmentName }}</h2>
            <p>Number of Submissions: {{ .NumberOfSubmissions }}</p>
            <p>Latest Submission: {{ .LatestSubmissionString }}</p>
        </div>
        <div class='ag-body'>
            <table>
                <thead>
                    <tr>
                        <th>Question</th>
                        <th>Mean</th>
                        <th>Median</th>
                        <th>Min</th>
                        <th>Max</th>
                        <th>StdDev</th>
                    </tr>
                </thead>
                <tbody>
                    {{ range .Questions }}
                        <tr>
                            <td class='text'>{{ .QuestionName }}</td>
                            <td class='numeric'>{{ .MeanString }}</td>
                            <td class='numeric'>{{ .MedianString }}</td>
                            <td class='numeric'>{{ .MinString }}</td>
                            <td class='numeric'>{{ .MaxString }}</td>
                            <td class='numeric'>{{ .StdDevString }}</td>
                        </tr>
                    {{ end }}
                </tbody>
            </table>
        </div>
    </div>
`
