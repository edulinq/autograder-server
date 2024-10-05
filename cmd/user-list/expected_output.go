package main

const expectedServerUserList = `{
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
                "course-with-lms": {
                    "id": "course-with-lms",
                    "name": "Course With LMS",
                    "role": "admin"
                },
                "course-without-source": {
                    "id": "course-without-source",
                    "name": "Course Without Source",
                    "role": "admin"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "admin"
                },
                "course101-with-zero-limit": {
                    "id": "course101-with-zero-limit",
                    "name": "Course 101 - With Zero Limit",
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
                "course-with-lms": {
                    "id": "course-with-lms",
                    "name": "Course With LMS",
                    "role": "grader"
                },
                "course-without-source": {
                    "id": "course-without-source",
                    "name": "Course Without Source",
                    "role": "grader"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "grader"
                },
                "course101-with-zero-limit": {
                    "id": "course101-with-zero-limit",
                    "name": "Course 101 - With Zero Limit",
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
                "course-with-lms": {
                    "id": "course-with-lms",
                    "name": "Course With LMS",
                    "role": "other"
                },
                "course-without-source": {
                    "id": "course-without-source",
                    "name": "Course Without Source",
                    "role": "other"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "other"
                },
                "course101-with-zero-limit": {
                    "id": "course101-with-zero-limit",
                    "name": "Course 101 - With Zero Limit",
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
                "course-with-lms": {
                    "id": "course-with-lms",
                    "name": "Course With LMS",
                    "role": "owner"
                },
                "course-without-source": {
                    "id": "course-without-source",
                    "name": "Course Without Source",
                    "role": "owner"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "owner"
                },
                "course101-with-zero-limit": {
                    "id": "course101-with-zero-limit",
                    "name": "Course 101 - With Zero Limit",
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
                "course-with-lms": {
                    "id": "course-with-lms",
                    "name": "Course With LMS",
                    "role": "student"
                },
                "course-without-source": {
                    "id": "course-without-source",
                    "name": "Course Without Source",
                    "role": "student"
                },
                "course101": {
                    "id": "course101",
                    "name": "Course 101",
                    "role": "student"
                },
                "course101-with-zero-limit": {
                    "id": "course101-with-zero-limit",
                    "name": "Course 101 - With Zero Limit",
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

const expectedServerUserListTable = `email  name    server-role     courses
course-admin@test.edulinq.org   course-admin    user    {
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "admin"
    },
    "course-with-lms": {
        "id": "course-with-lms",
        "name": "Course With LMS",
        "role": "admin"
    },
    "course-without-source": {
        "id": "course-without-source",
        "name": "Course Without Source",
        "role": "admin"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "admin"
    },
    "course101-with-zero-limit": {
        "id": "course101-with-zero-limit",
        "name": "Course 101 - With Zero Limit",
        "role": "admin"
    }
}
course-grader@test.edulinq.org  course-grader   user    {
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "grader"
    },
    "course-with-lms": {
        "id": "course-with-lms",
        "name": "Course With LMS",
        "role": "grader"
    },
    "course-without-source": {
        "id": "course-without-source",
        "name": "Course Without Source",
        "role": "grader"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "grader"
    },
    "course101-with-zero-limit": {
        "id": "course101-with-zero-limit",
        "name": "Course 101 - With Zero Limit",
        "role": "grader"
    }
}
course-other@test.edulinq.org   course-other    user    {
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "other"
    },
    "course-with-lms": {
        "id": "course-with-lms",
        "name": "Course With LMS",
        "role": "other"
    },
    "course-without-source": {
        "id": "course-without-source",
        "name": "Course Without Source",
        "role": "other"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "other"
    },
    "course101-with-zero-limit": {
        "id": "course101-with-zero-limit",
        "name": "Course 101 - With Zero Limit",
        "role": "other"
    }
}
course-owner@test.edulinq.org   course-owner    user    {
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "owner"
    },
    "course-with-lms": {
        "id": "course-with-lms",
        "name": "Course With LMS",
        "role": "owner"
    },
    "course-without-source": {
        "id": "course-without-source",
        "name": "Course Without Source",
        "role": "owner"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "owner"
    },
    "course101-with-zero-limit": {
        "id": "course101-with-zero-limit",
        "name": "Course 101 - With Zero Limit",
        "role": "owner"
    }
}
course-student@test.edulinq.org course-student  user    {
    "course-languages": {
        "id": "course-languages",
        "name": "Course Using Different Languages.",
        "role": "student"
    },
    "course-with-lms": {
        "id": "course-with-lms",
        "name": "Course With LMS",
        "role": "student"
    },
    "course-without-source": {
        "id": "course-without-source",
        "name": "Course Without Source",
        "role": "student"
    },
    "course101": {
        "id": "course101",
        "name": "Course 101",
        "role": "student"
    },
    "course101-with-zero-limit": {
        "id": "course101-with-zero-limit",
        "name": "Course 101 - With Zero Limit",
        "role": "student"
    }
}
root    root    root    {}
server-admin@test.edulinq.org   server-admin    admin   {}
server-creator@test.edulinq.org server-creator  creator {}
server-owner@test.edulinq.org   server-owner    owner   {}
server-user@test.edulinq.org    server-user     user    {}
`
