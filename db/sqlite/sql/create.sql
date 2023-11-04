DROP TABLE IF EXISTS Courses;
DROP TABLE IF EXISTS Assignments;
DROP TABLE IF EXISTS Users;
DROP TABLE IF EXISTS CourseUsers;
DROP TABLE IF EXISTS BackupTasks;
DROP TABLE IF EXISTS ReportTasks;
DROP TABLE IF EXISTS ScoreUploadTasks;

CREATE TABLE Courses (
    id INTEGER PRIMARY KEY,
    text_id TEXT UNIQUE,
    name TEXT,

    lms_type TEXT,
    lms_course_id TEXT,
    lms_token TEXT,
    lms_base_uri TEXT,
    lms_sync_user_attributes INTEGER,
    lms_sync_add_users INTEGER,
    lms_sync_remove_users INTEGER,

    -- TEST(eriq): Do we need this?
    source_path TEXT
);

CREATE TABLE Assignments (
    id INTEGER PRIMARY KEY,
    text_id TEXT UNIQUE,
    name TEXT,
    sort_id TEXT,

    course INTEGER REFERENCES Courses(id),

    lms_id TEXT,

    -- TODO(eriq): Docker info.

    late_policy_type TEXT,
    late_policy_penalty FLOAT,
    late_policy_reject_after_days INTEGER,
    late_policy_max_late_days INTEGER,
    late_policy_late_days_lms_id TEXT,

    -- TEST(eriq): Do we need this?
    source_path TEXT
);

CREATE TABLE Users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL,
    lms_id TEXT,
    name TEXT,
    role TEXT NOT NULL,
    pass TEXT NOT NULL,
    salt TEXT NOT NULL
);

CREATE TABLE CourseUsers (
    id INTEGER PRIMARY KEY,
    course INTEGER REFERENCES Courses(id),
    user INTEGER REFERENCES Users(id)
);

CREATE TABLE BackupTasks (
    id INTEGER PRIMARY KEY,
    course INTEGER REFERENCES Courses(id),
    disable INTEGER,
    day_of_week TEXT,
    time_of_day TEXT
);

CREATE TABLE ReportTasks (
    id INTEGER PRIMARY KEY,
    course INTEGER REFERENCES Courses(id),
    disable INTEGER,
    day_of_week TEXT,
    time_of_day TEXT,
    to_csv TEXT
);

CREATE TABLE ScoreUploadTasks (
    id INTEGER PRIMARY KEY,
    course INTEGER REFERENCES Courses(id),
    disable INTEGER,
    dry_run INTEGER,
    day_of_week TEXT,
    time_of_day TEXT
);
