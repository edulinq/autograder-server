package main

const EXPECTED_COURSE_USER_LIST = `{
    "users": [
        {
            "type": "course",
            "email": "course-admin@test.edulinq.org",
            "name": "course-admin",
            "role": "admin",
            "lms-id": "lms-course-admin@test.edulinq.org"
        },
        {
            "type": "course",
            "email": "course-grader@test.edulinq.org",
            "name": "course-grader",
            "role": "grader",
            "lms-id": "lms-course-grader@test.edulinq.org"
        },
        {
            "type": "course",
            "email": "course-other@test.edulinq.org",
            "name": "course-other",
            "role": "other",
            "lms-id": "lms-course-other@test.edulinq.org"
        },
        {
            "type": "course",
            "email": "course-owner@test.edulinq.org",
            "name": "course-owner",
            "role": "owner",
            "lms-id": "lms-course-owner@test.edulinq.org"
        },
        {
            "type": "course",
            "email": "course-student@test.edulinq.org",
            "name": "course-student",
            "role": "student",
            "lms-id": "lms-course-student@test.edulinq.org"
        }
    ]
}
`

const EXPECTED_COURSE_USER_LIST_TABLE = `email	name	course-role	LMSID
course-admin@test.edulinq.org	course-admin	admin	lms-course-admin@test.edulinq.org
course-grader@test.edulinq.org	course-grader	grader	lms-course-grader@test.edulinq.org
course-other@test.edulinq.org	course-other	other	lms-course-other@test.edulinq.org
course-owner@test.edulinq.org	course-owner	owner	lms-course-owner@test.edulinq.org
course-student@test.edulinq.org	course-student	student	lms-course-student@test.edulinq.org
`
