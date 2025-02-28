package main

const EXPECTED_COURSES_ASSIGNMENTS_GET_TABLE = `id	name
hw0	Homework 0
`

const EXPECTED_COURSES_ASSIGNMENTS_LIST_TABLE = EXPECTED_COURSES_ASSIGNMENTS_GET_TABLE + `hw1	Homework 1
`

const EXPECTED_FETCH_USER_HISTORY_TABLE = `found-user	history
true	[{"assignment-id":"hw0","course-id":"course101","grading_start_time":1697406256000,"id":"course101::hw0::course-student@test.edulinq.org::1697406256","max_points":2,"message":"","score":0,"short-id":"1697406256","user":"course-student@test.edulinq.org"},{"assignment-id":"hw0","course-id":"course101","grading_start_time":1697406266000,"id":"course101::hw0::course-student@test.edulinq.org::1697406265","max_points":2,"message":"","score":1,"short-id":"1697406265","user":"course-student@test.edulinq.org"},{"assignment-id":"hw0","course-id":"course101","grading_start_time":1697406273000,"id":"course101::hw0::course-student@test.edulinq.org::1697406272","max_points":2,"message":"","score":2,"short-id":"1697406272","user":"course-student@test.edulinq.org"}]
`

const EXPECTED_COURSES_USERS_GET_TABLE = `found	user
true	{"email":"course-student@test.edulinq.org","lms-id":"lms-course-student@test.edulinq.org","name":"course-student","role":"student","type":"course"}
`
