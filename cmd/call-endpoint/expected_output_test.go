package main

const EXPECTED_SERVER_USER_LIST_TABLE = `courses	email	name	role	type
{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"admin"},"course101":{"id":"course101","name":"Course 101","role":"admin"}}	course-admin@test.edulinq.org	course-admin	user	server
{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"grader"},"course101":{"id":"course101","name":"Course 101","role":"grader"}}	course-grader@test.edulinq.org	course-grader	user	server
{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"other"},"course101":{"id":"course101","name":"Course 101","role":"other"}}	course-other@test.edulinq.org	course-other	user	server
{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"owner"},"course101":{"id":"course101","name":"Course 101","role":"owner"}}	course-owner@test.edulinq.org	course-owner	user	server
{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"student"},"course101":{"id":"course101","name":"Course 101","role":"student"}}	course-student@test.edulinq.org	course-student	user	server
{}	root	root	root	server
{}	server-admin@test.edulinq.org	server-admin	admin	server
{}	server-creator@test.edulinq.org	server-creator	creator	server
{}	server-owner@test.edulinq.org	server-owner	owner	server
{}	server-user@test.edulinq.org	server-user	user	server
`

const EXPECTED_COURSES_USERS_LIST_TABLE = `email	lms-id	name	role	type
course-admin@test.edulinq.org	lms-course-admin@test.edulinq.org	course-admin	admin	course
course-grader@test.edulinq.org	lms-course-grader@test.edulinq.org	course-grader	grader	course
course-other@test.edulinq.org	lms-course-other@test.edulinq.org	course-other	other	course
course-owner@test.edulinq.org	lms-course-owner@test.edulinq.org	course-owner	owner	course
course-student@test.edulinq.org	lms-course-student@test.edulinq.org	course-student	student	course
`

const EXPECTED_COURSES_ASSIGNMENTS_TABLE = `id	name
hw0	Homework 0
`
const EXPECTED_COURSES_USERS_GET_TABLE = `found	user
true	{"email":"course-student@test.edulinq.org","lms-id":"lms-course-student@test.edulinq.org","name":"course-student","role":"student","type":"course"}
`

const EXPECTED_FETCH_COURSE_SCORES_TABLE = `course-admin@test.edulinq.org	course-grader@test.edulinq.org	course-other@test.edulinq.org	course-owner@test.edulinq.org	course-student@test.edulinq.org
<nil>	<nil>	<nil>	<nil>	{"assignment-id":"hw0","course-id":"course101","grading_start_time":1697406273000,"id":"course101::hw0::course-student@test.edulinq.org::1697406272","max_points":2,"message":"","score":2,"short-id":"1697406272","user":"course-student@test.edulinq.org"}
`
const EXPECTED_FETCH_COURSE_ATTEMPTS_TABLE = `course-admin@test.edulinq.org	course-grader@test.edulinq.org	course-other@test.edulinq.org	course-owner@test.edulinq.org	course-student@test.edulinq.org
<nil>	<nil>	<nil>	<nil>	{"info":{"additional-info":null,"assignment-id":"hw0","course-id":"course101","grading_end_time":1697406273000,"grading_start_time":1697406273000,"id":"course101::hw0::course-student@test.edulinq.org::1697406272","max_points":2,"message":"","name":"HW0","questions":[{"grading_end_time":1697406273000,"grading_start_time":1697406273000,"max_points":1,"message":"","name":"Q1","score":1},{"grading_end_time":1697406273000,"grading_start_time":1697406273000,"max_points":1,"message":"","name":"Q2","score":1},{"grading_end_time":1697406273000,"grading_start_time":1697406273000,"max_points":0,"message":"Style is clean!","name":"Style","score":0}],"score":2,"short-id":"1697406272","user":"course-student@test.edulinq.org"},"input-files-gzip":{"submission.py":"H4sICAAAAAAA/3N1Ym1pc3Npb24ucHkASklNU0grzUsuyczPM9TQtOJSUFBQKEotKS3KUwgpKk3l4kJWYaRRlpiDqqgsMUdBW8GQCxAAAP//PpwmbkkAAAA="},"output-files-gzip":{"result.json":"H4sICAAAAAAA/3Jlc3VsdC5qc29uAMSSvQqDMBSFd5/iNrNDDPUnPkHX0qFDKRI0SEBja1JoEd+9mIptg1YyNdPl3HPO5YN0HgAAkqzmKAW0O2Lkv6TrjSstGqlQCicjDa+bpq/cPhhj06Zm9+zSCKmHfGAtVd60fEavuVKsNI12X9myQsgyU5q1OtPC3A0iSuMtTggJKVkIcFnM2id376/Skb/ThW50oQPdQT8q/hMQLwDa+gegKQWhIK84kxtn3sSFNyLxm9dM5/ETr97BlnGhn3q99wwAAP//PProBisDAAA="},"stderr":"","stdout":"Autograder transcript for assignment: HW0.\nGrading started at 2023-11-11 22:13 and ended at 2023-11-11 22:13.\nQ1: 1 / 1\nQ2: 1 / 1\nStyle: 0 / 0\n   Style is clean!\n\nTotal: 2 / 2\n"}
`

const EXPECTED_FETCH_USER_ATTEMPT_TABLE = `found-submission	found-user	grading-result
true	true	{"info":{"additional-info":null,"assignment-id":"hw0","course-id":"course101","grading_end_time":1697406256000,"grading_start_time":1697406256000,"id":"course101::hw0::course-student@test.edulinq.org::1697406256","max_points":2,"message":"","name":"HW0","questions":[{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q1","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q2","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":0,"message":"Style is clean!","name":"Style","score":0}],"score":0,"short-id":"1697406256","user":"course-student@test.edulinq.org"},"input-files-gzip":{"submission.py":"H4sICAAAAAAA/3N1Ym1pc3Npb24ucHkASklNU0grzUsuyczPM9TQtOJSUFBQKEotKS3KU/DLL/HMLchJzU3NK0lN4eJCVmukUZaYg1c5IAAA//8hrjgTWgAAAA=="},"output-files-gzip":{"result.json":"H4sICAAAAAAA/3Jlc3VsdC5qc29uAMySwWrDMAyG73kKzecy4rRrkz7BdhmMHXYYI5haBEMsd7YKG6XvPuIlGTFsWW7NSUj/J+UDnzMAAEHKotiDuH/Jxeq79X7CwMZREHt4ja3uO4/VhHuSPTZOrPqoj84Qd7xMhuHgfMflKYQhqCZufHT8YI8tWiRGDR755An1bXqn8UobaurAynPNJv6P3FbVbpOXhVxX6Y0BQNJDXEzyYsxfVrPexdV6r+e9J/EF1s/82eKf4qngP8TjUjABDi0qulnsu13iuynufnxj9dY/+7k7ZZkEf9m/yy7ZVwAAAP//do3jyV0DAAA="},"stderr":"Dummy Stderr\n","stdout":"Autograder transcript for assignment: HW0.\nGrading started at 2023-11-11 22:13 and ended at 2023-11-11 22:13.\nQ1: 0 / 1\n   NotImplemented returned.\nQ2: 0 / 1\n   NotImplemented returned.\nStyle: 0 / 0\n   Style is clean!\n\nTotal: 0 / 2\n"}
`

const EXPECTED_FETCH_USER_ATTEMPTS_TABLE = `found-user	grading-results
true	{"info":{"additional-info":null,"assignment-id":"hw0","course-id":"course101","grading_end_time":1697406256000,"grading_start_time":1697406256000,"id":"course101::hw0::course-student@test.edulinq.org::1697406256","max_points":2,"message":"","name":"HW0","questions":[{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q1","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q2","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":0,"message":"Style is clean!","name":"Style","score":0}],"score":0,"short-id":"1697406256","user":"course-student@test.edulinq.org"},"input-files-gzip":{"submission.py":"H4sICAAAAAAA/3N1Ym1pc3Npb24ucHkASklNU0grzUsuyczPM9TQtOJSUFBQKEotKS3KU/DLL/HMLchJzU3NK0lN4eJCVmukUZaYg1c5IAAA//8hrjgTWgAAAA=="},"output-files-gzip":{"result.json":"H4sICAAAAAAA/3Jlc3VsdC5qc29uAMySwWrDMAyG73kKzecy4rRrkz7BdhmMHXYYI5haBEMsd7YKG6XvPuIlGTFsWW7NSUj/J+UDnzMAAEHKotiDuH/Jxeq79X7CwMZREHt4ja3uO4/VhHuSPTZOrPqoj84Qd7xMhuHgfMflKYQhqCZufHT8YI8tWiRGDR755An1bXqn8UobaurAynPNJv6P3FbVbpOXhVxX6Y0BQNJDXEzyYsxfVrPexdV6r+e9J/EF1s/82eKf4qngP8TjUjABDi0qulnsu13iuynufnxj9dY/+7k7ZZkEf9m/yy7ZVwAAAP//do3jyV0DAAA="},"stderr":"Dummy Stderr\n","stdout":"Autograder transcript for assignment: HW0.\nGrading started at 2023-11-11 22:13 and ended at 2023-11-11 22:13.\nQ1: 0 / 1\n   NotImplemented returned.\nQ2: 0 / 1\n   NotImplemented returned.\nStyle: 0 / 0\n   Style is clean!\n\nTotal: 0 / 2\n"}
`

const EXPECTED_FETCH_USER_HISTORY_TABLE = `found-user	history
true	{"assignment-id":"hw0","course-id":"course101","grading_start_time":1697406256000,"id":"course101::hw0::course-student@test.edulinq.org::1697406256","max_points":2,"message":"","score":0,"short-id":"1697406256","user":"course-student@test.edulinq.org"}
`

const EXPECTED_PEEK_TABLE = `found-submission	found-user	submission-result
true	true	{"additional-info":null,"assignment-id":"hw0","course-id":"course101","grading_end_time":1697406256000,"grading_start_time":1697406256000,"id":"course101::hw0::course-student@test.edulinq.org::1697406256","max_points":2,"message":"","name":"HW0","questions":[{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q1","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":1,"message":"NotImplemented returned.","name":"Q2","score":0},{"grading_end_time":1697406256000,"grading_start_time":1697406256000,"max_points":0,"message":"Style is clean!","name":"Style","score":0}],"score":0,"short-id":"1697406256","user":"course-student@test.edulinq.org"}
`

const EXPECTED_SUBMISSIONS_REMOVE_TABLE = `found-submission	found-user
true	true
`

const EXPECTED_LMS_USER_GET_TABLE = `found-autograder-user	found-lms-user	user
true	true	{"email":"course-student@test.edulinq.org","lms-id":"lms-course-student@test.edulinq.org","name":"course-student","role":"student","type":"course"}
`

const EXPECTED_DESCRIBE_TABLE = `courses/admin/update	courses/assignments/get	courses/assignments/list	courses/assignments/submissions/fetch/course/attempts	courses/assignments/submissions/fetch/course/scores	courses/assignments/submissions/fetch/user/attempt	courses/assignments/submissions/fetch/user/attempts	courses/assignments/submissions/fetch/user/history	courses/assignments/submissions/fetch/user/peek	courses/assignments/submissions/remove	courses/assignments/submissions/submit	courses/upsert/filespec	courses/upsert/zip	courses/users/drop	courses/users/enroll	courses/users/get	courses/users/list	lms/upload/scores	lms/user/get	logs/query	metadata/describe	users/auth	users/get	users/list	users/password/change	users/password/reset	users/remove	users/tokens/create	users/tokens/delete	users/upsert
{"request-type":"*admin.UpdateRequest","response-type":"*admin.UpdateResponse"}	{"request-type":"*assignments.GetRequest","response-type":"*assignments.GetResponse"}	{"request-type":"*assignments.ListRequest","response-type":"*assignments.ListResponse"}	{"request-type":"*submissions.FetchCourseAttemptsRequest","response-type":"*submissions.FetchCourseAttemptsResponse"}	{"request-type":"*submissions.FetchCourseScoresRequest","response-type":"*submissions.FetchCourseScoresResponse"}	{"request-type":"*submissions.FetchUserAttemptRequest","response-type":"*submissions.FetchUserAttemptResponse"}	{"request-type":"*submissions.FetchUserAttemptsRequest","response-type":"*submissions.FetchUserAttemptsResponse"}	{"request-type":"*submissions.FetchUserHistoryRequest","response-type":"*submissions.FetchUserHistoryResponse"}	{"request-type":"*submissions.FetchUserPeekRequest","response-type":"*submissions.FetchUserPeekResponse"}	{"request-type":"*submissions.RemoveRequest","response-type":"*submissions.RemoveResponse"}	{"request-type":"*submissions.SubmitRequest","response-type":"*submissions.SubmitResponse"}	{"request-type":"*upsert.FileSpecRequest","response-type":"*upsert.UpsertResponse"}	{"request-type":"*upsert.ZipFileRequest","response-type":"*upsert.UpsertResponse"}	{"request-type":"*users.DropRequest","response-type":"*users.DropResponse"}	{"request-type":"*users.EnrollRequest","response-type":"*users.EnrollResponse"}	{"request-type":"*users.GetRequest","response-type":"*users.GetResponse"}	{"request-type":"*users.ListRequest","response-type":"*users.ListResponse"}	{"request-type":"*lms.UploadScoresRequest","response-type":"*lms.UploadScoresResponse"}	{"request-type":"*lms.UserGetRequest","response-type":"*lms.UserGetResponse"}	{"request-type":"*logs.QueryRequest","response-type":"*logs.QueryResponse"}	{"request-type":"*metadata.DescribeRequest","response-type":"*metadata.DescribeResponse"}	{"request-type":"*users.AuthRequest","response-type":"*users.AuthResponse"}	{"request-type":"*users.GetRequest","response-type":"*users.GetResponse"}	{"request-type":"*users.ListRequest","response-type":"*users.ListResponse"}	{"request-type":"*users.PasswordChangeRequest","response-type":"*users.PasswordChangeResponse"}	{"request-type":"*users.PasswordResetRequest","response-type":"*users.PasswordResetResponse"}	{"request-type":"*users.RemoveRequest","response-type":"*users.RemoveResponse"}	{"request-type":"*users.TokensCreateRequest","response-type":"*users.TokensCreateResponse"}	{"request-type":"*users.TokensDeleteRequest","response-type":"*users.TokensDeleteResponse"}	{"request-type":"*users.UpsertRequest","response-type":"*users.UpsertResponse"}
`

const EXPECTED_USERS_GET_TABLE = `found	user
true	{"courses":{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"student"},"course101":{"id":"course101","name":"Course 101","role":"student"}},"email":"course-student@test.edulinq.org","name":"course-student","role":"user","type":"server"}
`
