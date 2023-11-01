package report

import (
    "fmt"
    "html/template"
    "strings"
)

func (this *CourseScoringReport) ToHTML() (string, error) {
    title := fmt.Sprintf("Course Scoring Report for %s", this.CourseName);
    templateHTML := fmt.Sprintf(outterShell, title, style, courseReportTemplate);

    tmpl, err := template.New("course-scoring-report").Parse(templateHTML);
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

func (this *AssignmentScoringReport) ToHTML(inline bool) (string, error) {
    templateHTML := assignmentReportTemplate;
    if (!inline) {
        title := fmt.Sprintf("Assignment Scoring Report for %s", this.AssignmentName);
        templateHTML = fmt.Sprintf(outterShell, title, style, assignmentReportTemplate);
    }

    tmpl, err := template.New("assignment-scoring-report").Parse(templateHTML);
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

func (this *AssignmentScoringReport) ToInlineHTML() (template.HTML, error) {
    html, err := this.ToHTML(true);
    if (err != nil) {
        return template.HTML(""), fmt.Errorf("Failed to generate HTML scoring report for assignment '%s': '%w'.",
                this.AssignmentName, err);
    }

    return template.HTML(html), nil;
}

// Replacements: [title, head, body]
var outterShell string = `
    <html>
        <head>
            <meta charset="utf-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0">

            <title>%s</title>

            %s
        </head>
        <body>
            %s
        </body>
    </html>
`

var courseReportTemplate string = `
    <div class='autograder autograder-course-scoring-report'>
        <div class='ag-header'>
            <h1>Course: {{ .CourseName }}</h1>
        </div>
        <div class='ag-body'>
            {{ range .Assignments }}
                {{ if eq .NumberOfSubmissions 0 }} {{ continue }} {{ end }}

                <div>
                    {{ .ToInlineHTML }}
                </div>
            {{ end }}
        </div>
    </div>
`

var assignmentReportTemplate string = `
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

var style string = `
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
            padding-right: 10px;
        }

        .autograder-assignment-scoring-report table td.text {
            padding-right: 15px;
        }

        .autograder-assignment-scoring-report table tr:last-child {
            font-style: italic;
        }
    </style>
`
