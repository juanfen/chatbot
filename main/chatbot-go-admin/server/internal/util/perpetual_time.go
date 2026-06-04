package util

import "time"

// PerpetualTime is used for expire_time/delete_time when a device audio is bound to a speaker.
var PerpetualTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.Local)

// IsPerpetual reports whether t represents an indefinite retention time.
func IsPerpetual(t time.Time) bool {
	return !t.Before(PerpetualTime)
}
