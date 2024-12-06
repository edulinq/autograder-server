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

const EXPECTED_COURSE_USER_LIST_TABLE = `email	lms-id	name	role	type
course-admin@test.edulinq.org	lms-course-admin@test.edulinq.org	course-admin	admin	course
course-grader@test.edulinq.org	lms-course-grader@test.edulinq.org	course-grader	grader	course
course-other@test.edulinq.org	lms-course-other@test.edulinq.org	course-other	other	course
course-owner@test.edulinq.org	lms-course-owner@test.edulinq.org	course-owner	owner	course
course-student@test.edulinq.org	lms-course-student@test.edulinq.org	course-student	student	course
`

const EXPECTED_COURSE_ASSIGNMENTS_LIST_TABLE = `id	name
hw0	Homework 0
`
