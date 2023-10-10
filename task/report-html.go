package task

import (
    "fmt"
    "html/template"
    "strings"
)

func (this *ScoringReport) ToHTML() (string, error) {
    var builder strings.Builder;

    tmpl, err := template.New("scoring-report").Parse(reportTemplate);
    if (err != nil) {
        return "", fmt.Errorf("Could not parse scoring report template: '%w'.", err);
    }

    err = tmpl.Execute(&builder, this);
    if (err != nil) {
        return "", fmt.Errorf("Failed to execute scoring report template: '%w'.", err);
    }

    return builder.String(), nil;
}

var reportTemplate string = `
    <style>
        .autograder-scoring-report table th,
        .autograder-scoring-report table .text {
            text-align: left;
        }

        .autograder-scoring-report table .numeric {
            text-align: right;
        }

        .autograder-scoring-report table th,
        .autograder-scoring-report table td {
            padding: 5px;
        }
    </style>

    <div class='autograder autograder-scoring-report'>
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
