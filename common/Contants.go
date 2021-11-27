package common

const (
	JOB_SAVE_DIR = "cron/jobs/"
	KILL_JOB_DIR = "cron/kill/"

	JOB_EVENT_SAVE   = 1
	JOB_EVENT_DELETE = 2
	JOB_EVENT_KILL   = 3

	JOB_LOCK_DIR = "cron/lock/"

	JOB_REGISTER_DIR = "cron/workers/"
)
