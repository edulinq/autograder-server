package main

const EXPECTED_SERVER_USER_LIST_TABLE = `email	name	server-role	courses
course-admin@test.edulinq.org	course-admin	user	{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"admin"},"course101":{"id":"course101","name":"Course 101","role":"admin"}}
course-grader@test.edulinq.org	course-grader	user	{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"grader"},"course101":{"id":"course101","name":"Course 101","role":"grader"}}
course-other@test.edulinq.org	course-other	user	{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"other"},"course101":{"id":"course101","name":"Course 101","role":"other"}}
course-owner@test.edulinq.org	course-owner	user	{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"owner"},"course101":{"id":"course101","name":"Course 101","role":"owner"}}
course-student@test.edulinq.org	course-student	user	{"course-languages":{"id":"course-languages","name":"Course Using Different Languages.","role":"student"},"course101":{"id":"course101","name":"Course 101","role":"student"}}
root	root	root	{}
server-admin@test.edulinq.org	server-admin	admin	{}
server-creator@test.edulinq.org	server-creator	creator	{}
server-owner@test.edulinq.org	server-owner	owner	{}
server-user@test.edulinq.org	server-user	user	{}
`

const EXPECTED_COURSE_USER_LIST_TABLE = `email	name	role	lms-id
course-admin@test.edulinq.org	course-admin	admin	lms-course-admin@test.edulinq.org
course-grader@test.edulinq.org	course-grader	grader	lms-course-grader@test.edulinq.org
course-other@test.edulinq.org	course-other	other	lms-course-other@test.edulinq.org
course-owner@test.edulinq.org	course-owner	owner	lms-course-owner@test.edulinq.org
course-student@test.edulinq.org	course-student	student	lms-course-student@test.edulinq.org
`
