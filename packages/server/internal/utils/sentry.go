package utils

import (
	"os"

	"github.com/getsentry/sentry-go"
)

func SendExceptionToSentry(err error) {
	if os.Getenv("ENV") == "production" {
		sentry.CaptureException(err)
	}
}
