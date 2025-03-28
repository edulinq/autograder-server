package config

var (
	// Base
	NAME = MustNewStringOption("instance.name", "autograder",
		"A name to identify this autograder instance."+
			" Should only contain alphanumerics and underscores.")

	// Directories
	BASE_DIR = MustNewStringOption("dirs.base", GetDefaultBaseDir(),
		"The base dir for autograder to store data."+
			" SHOULD NOT be set in config files (to prevent cycles), only on the command-line or ENV.")
	BACKUP_DIR = MustNewStringOption("dirs.backup", "", "Path to where backups are made. Defaults to inside BASE_DIR.")

	// Debugging / Testing
	KEEP_BUILD_DIRS   = MustNewBoolOption("build.keep", false, "Keep artifacts/dirs used when building (not building the server itself, but things like assignment images).")
	UNIT_TESTING_MODE = MustNewBoolOption("testing", false, "Assume tests are being run, which may alter some operations.")
	LOAD_TEST_DATA    = MustNewBoolOption("testdata.load", false, "Load test data when the database opens.")
	STORE_HTTP        = MustNewStringOption("http.store", "", "Store HTTP requests made by the server to the specified directory.")
	CI_DOCKER_BUILD   = MustNewBoolOption("ci.docker.build", false, "Inform the system that we are running in a Docker build step of a CI. There are some tests that are flaky in this environment.")

	// Logging
	LOG_TEXT_LEVEL    = MustNewStringOption("log.text.level", "INFO", "The default logging level for the text (stderr) logger.")
	LOG_BACKEND_LEVEL = MustNewStringOption("log.backend.level", "INFO", "The default logging level for the backend (database) logger.")

	// Stats
	STATS_SYSTEM_INTERVAL_MS = MustNewIntOption("stats.system.interval", 5*1000, "The number of milliseconds between system stats collection events.")

	// Email
	EMAIL_FROM                 = MustNewStringOption("email.from", "", "From address for emails sent from the autograder.")
	EMAIL_HOST                 = MustNewStringOption("email.host", "", "SMTP host for emails sent from the autograder.")
	EMAIL_PASS                 = MustNewStringOption("email.pass", "", "SMTP password for emails sent from the autograder.")
	EMAIL_PORT                 = MustNewStringOption("email.port", "", "SMTP port for emails sent from the autograder.")
	EMAIL_USER                 = MustNewStringOption("email.user", "", "SMTP username for emails sent from the autograder.")
	EMAIL_SMTP_IDLE_TIMEOUT_MS = MustNewIntOption("email.smtp.idle", 2*60*1000, "Consider an SMTP connection idle if no emails are sent for this number of milliseconds.")
	EMAIL_MIN_PERIOD           = MustNewIntOption("email.smtp.minperiod", 250, "The minimum time (in MS) between sending emails.")

	// Docker
	DOCKER_DISABLE            = MustNewBoolOption("docker.disable", false, "Disable the use of docker (usually for testing).")
	DOCKER_MAX_OUTPUT_SIZE_KB = MustNewIntOption("docker.output.maxsize", 4*1024, "The maximum allowed size (in KB) for stdout and stderr combined. The default is 4096 KB (4 MB).")

	// Grading
	GRADING_RUNTIME_MAX_SECS = MustNewIntOption("grading.runtime.max", 60*5, "The maximum number of seconds a Docker container can be running for.")

	// Tasks
	NO_TASKS             = MustNewBoolOption("tasks.disable", false, "Disable all scheduled tasks.")
	TASK_MAX_WAIT_SECS   = MustNewIntOption("tasks.maxwait", 2*60, "The maximum wait between checking for the next task to run.")
	TASK_MIN_PERIOD_SECS = MustNewIntOption("tasks.minperiod", 10*60, "The minimum period between the runs of the same task.")

	// Server
	WEB_HTTP_PORT        = MustNewIntOption("web.http.port", 8080, "The port to serve HTTP traffic on. Standard is 80 (but requires root to use).")
	WEB_HTTP_REDIRECT    = MustNewBoolOption("web.http.redirect", true, "Redirect HTTP traffic to HTTPS. Only used if HTTPS is enabled.")
	WEB_HTTPS_ENABLE     = MustNewBoolOption("web.https.enable", false, "Enable HTTPS. A certificate and key must be provided.")
	WEB_HTTPS_PORT       = MustNewIntOption("web.https.port", 8081, "The port to serve HTTPS traffic on. Standard is 443 (but requires root to use).")
	WEB_HTTPS_CERT       = MustNewStringOption("web.https.cert", "", "The path to a PEM encoded SSL certificate file. The certificate file may contain intermediate certificates following the leaf certificate to form a certificate chain.")
	WEB_HTTPS_KEY        = MustNewStringOption("web.https.key", "", "The path to a PEM encoded SSL key file.")
	WEB_MAX_FILE_SIZE_KB = MustNewIntOption("web.maxsize", 2*1024, "The maximum allowed file size (in KB) submitted via POST request. The default is 2048 KB (2 MB).")
	WEB_STATIC_ROOT      = MustNewStringOption("web.static.root", "", "The root directory to serve as part of the static portion of the API. Defaults to empty string, which indicates the embedded static directory.")
	WEB_STATIC_FALLBACK  = MustNewBoolOption("web.static.fallback", false, "For any unmatched route (potential 404) that does not have an API prefix, try to match it in the static root before giving the final 404.")

	// Database
	DB_TYPE   = MustNewStringOption("db.type", "disk", "The type of database to use.")
	DB_PG_URI = MustNewStringOption("db.pg.uri", "", "Connection string to connect to a Postgres Database. Empty if not using Postgres.")

	// Code Analysis
	ANALYSIS_INDIVIDUAL_COURSE_POOL_SIZE = MustNewIntOption("analysis.individual.poolsize", 1, "The number of parallel workers per course when computing individual analysis.")
	ANALYSIS_PAIRWISE_COURSE_POOL_SIZE   = MustNewIntOption("analysis.pairwise.poolsize", 1, "The number of parallel workers per course when computing pairwise analysis.")

	STALELOCK_DURATION_SECS = MustNewIntOption("lockmanager.staleduration", 2*60*60, "Number of seconds a lock can be unused before getting removed.")
)
