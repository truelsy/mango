package mango

import "time"

const (
	sessionLifeTime = time.Minute * 60
	bufferSize      = 1024 * 4
)

const (
	defaultPendingWriteNum = 100000
	maxWriteRetryCount     = 3
)
