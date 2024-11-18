package config

var (
	// Base
	NAME = MustNewStringOption("instance.name", "autograder",
		"A name to identify this autograder instance."+
			" Should only contain alphanumerics and underscores.")

	// Directories
	BASE_DIR = MustNewStringOption("dirs.base", GetDefaultBaseDir(),
		"The base dir for autograder to store data."+
			" SHOULD NOT be set in config files (to prevent cycles), only on the command-line.")

	// Debugging / Testing
	KEEP_BUILD_DIRS   = MustNewBoolOption("build.keep", false, "Keep artifacts/dirs used when building (not building the server itself, but things like assignment images).")
	UNIT_TESTING_MODE = MustNewBoolOption("testing", false, "Assume tests are being run, which may alter some operations.")
	LOAD_TEST_DATA    = MustNewBoolOption("testdata.load", false, "Load test data when the database opens.")
	STORE_HTTP        = MustNewStringOption("http.store", "", "Store HTTP requests made by the server to the specified directory.")

	// Logging
	LOG_TEXT_LEVEL    = MustNewStringOption("log.text.level", "INFO", "The default logging level for the text (stderr) logger.")
	LOG_BACKEND_LEVEL = MustNewStringOption("log.backend.level", "INFO", "The default logging level for the backend (database) logger.")

	// Email
	EMAIL_FROM                 = MustNewStringOption("email.from", "", "From address for emails sent from the autograder.")
	EMAIL_HOST                 = MustNewStringOption("email.host", "", "SMTP host for emails sent from the autograder.")
	EMAIL_PASS                 = MustNewStringOption("email.pass", "", "SMTP password for emails sent from the autograder.")
	EMAIL_PORT                 = MustNewStringOption("email.port", "", "SMTP port for emails sent from the autograder.")
	EMAIL_USER                 = MustNewStringOption("email.user", "", "SMTP username for emails sent from the autograder.")
	EMAIL_SMTP_IDLE_TIMEOUT_MS = MustNewIntOption("email.smtp.idle", 2000, "Consider an SMTP connection idle if no emails are sent for this number of milliseconds.")

	// Docker
	DOCKER_DISABLE            = MustNewBoolOption("docker.disable", true, "Disable the use of docker (usually for testing).")
	DOCKER_RUNTIME_MAX_SECS   = MustNewIntOption("docker.runtime.max", 60*5, "The maximum number of seconds a Docker container can be running for.")
	DOCKER_MAX_OUTPUT_SIZE_KB = MustNewIntOption("docker.output.maxsize", 4*1024, "The maximum allowed size (in KB) for stdout and stderr combined. The default is 4096 KB (4 MB).")

	// Tasks
	NO_TASKS           = MustNewBoolOption("tasks.disable", false, "Disable all scheduled tasks.")
	TASK_MIN_REST_SECS = MustNewIntOption("tasks.minrest", 5*60,
		"The minimum time (in seconds) between invocations of the same task."+
			" A task instance that tries to run too quickly will be skipped.")
	TASK_BACKUP_DIR = MustNewStringOption("tasks.backup.dir", "", "Path to where backups are made. Defaults to inside BASE_DIR.")

	// Server
	WEB_PORT             = MustNewIntOption("web.port", 8080, "The port for the web interface to serve on.")
	WEB_MAX_FILE_SIZE_KB = MustNewIntOption("web.maxsize", 2*1024, "The maximum allowed file size (in KB) submitted via POST request. The default is 2048 KB (2 MB).")
	WEB_STATIC_ROOT      = MustNewStringOption("web.static.root", "", "The root directory to serve as part of the static portion of the API. Defaults to empty string, which indicates the embedded static directory.")
	WEB_STATIC_FALLBACK  = MustNewBoolOption("web.static.fallback", false, "For any unmatched route (potential 404) that does not have an API prefix, try to match it in the static root before giving the final 404.")

	// Database
	DB_TYPE   = MustNewStringOption("db.type", "disk", "The type of database to use.")
	DB_PG_URI = MustNewStringOption("db.pg.uri", "", "Connection string to connect to a Postgres Database. Empty if not using Postgres.")

	STALELOCK_DURATION_SECS = MustNewIntOption("lockmanager.staleduration", 2*60*60, "Number of seconds a lock can be unused before getting removed.")
)
