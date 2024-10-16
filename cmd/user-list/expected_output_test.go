package main

const EXPECTED_SERVER_USER_LIST = `{
    "users": [
        {
            "type": "server",
            "email": "course-admin@test.edulinq.org",
            "name": "course-admin",
            "role": "user",
            "courses": {
                "course-languages": {
                    "id": "course-languages",
                    "name": "Course Using Different Languages.",
                    "role": "admin"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "admin"
                }
            }
        },
        {
            "type": "server",
            "email": "course-grader@test.edulinq.org",
            "name": "course-grader",
            "role": "user",
            "courses": {
                "course-languages": {
                    "id": "course-languages",
                    "name": "Course Using Different Languages.",
                    "role": "grader"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "grader"
                }
            }
        },
        {
            "type": "server",
            "email": "course-other@test.edulinq.org",
            "name": "course-other",
            "role": "user",
            "courses": {
                "course-languages": {
                    "id": "course-languages",
                    "name": "Course Using Different Languages.",
                    "role": "other"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "other"
                }
            }
        },
        {
            "type": "server",
            "email": "course-owner@test.edulinq.org",
            "name": "course-owner",
            "role": "user",
            "courses": {
                "course-languages": {
                    "id": "course-languages",
                    "name": "Course Using Different Languages.",
                    "role": "owner"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "owner"
                }
            }
        },
        {
            "type": "server",
            "email": "course-student@test.edulinq.org",
            "name": "course-student",
            "role": "user",
            "courses": {
                "course-languages": {
                    "id": "course-languages",
                    "name": "Course Using Different Languages.",
                    "role": "student"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "student"
                }
            }
        },
        {
            "type": "server",
            "email": "root",
            "name": "root",
            "role": "root",
            "courses": {}
        },
        {
            "type": "server",
            "email": "server-admin@test.edulinq.org",
            "name": "server-admin",
            "role": "admin",
            "courses": {}
        },
        {
            "type": "server",
            "email": "server-creator@test.edulinq.org",
            "name": "server-creator",
            "role": "creator",
            "courses": {}
        },
        {
            "type": "server",
            "email": "server-owner@test.edulinq.org",
            "name": "server-owner",
            "role": "owner",
            "courses": {}
        },
        {
            "type": "server",
            "email": "server-user@test.edulinq.org",
            "name": "server-user",
            "role": "user",
            "courses": {}
        }
    ]
}
`

const EXPECTED_SERVER_USER_LIST_TABLE = `email	name	server-role	courses
course-admin@test.edulinq.org	course-admin	user	{
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "admin"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "admin"
    }
}
course-grader@test.edulinq.org	course-grader	user	{
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "grader"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "grader"
    }
}
course-other@test.edulinq.org	course-other	user	{
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "other"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "other"
    }
}
course-owner@test.edulinq.org	course-owner	user	{
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "owner"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "owner"
    }
}
course-student@test.edulinq.org	course-student	user	{
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "student"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "student"
    }
}
root	root	root	{}
server-admin@test.edulinq.org	server-admin	admin	{}
server-creator@test.edulinq.org	server-creator	creator	{}
server-owner@test.edulinq.org	server-owner	owner	{}
server-user@test.edulinq.org	server-user	user	{}
`
