package main

const EXPECTED_SERVER_USER_GET = `{
    "found": true,
    "user": {
        "type": "server",
        "email": "server-user@test.edulinq.org",
        "name": "server-user",
        "role": "user",
        "courses": {}
    }
}
`

const EXPECTED_UNKNOWN_SERVER_USER_GET = `{
    "found": false,
    "user": null
}
`
