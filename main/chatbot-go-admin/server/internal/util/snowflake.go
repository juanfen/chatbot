package util

import (
	"fmt"
	"sync"
	"time"
)

const snowflakeNodeID = 1

var (
	snowflakeMu sync.Mutex
	snowflakeMS int64
	snowflakeSeq uint16
)

// NewSnowflakeID32 returns a 32-digit decimal snowflake-style unique ID.
func NewSnowflakeID32() string {
	snowflakeMu.Lock()
	defer snowflakeMu.Unlock()

	now := time.Now().UnixMilli()
	if now == snowflakeMS {
		snowflakeSeq++
		if snowflakeSeq > 9999 {
			for now <= snowflakeMS {
				time.Sleep(time.Millisecond)
				now = time.Now().UnixMilli()
			}
			snowflakeSeq = 0
			snowflakeMS = now
		}
	} else {
		snowflakeSeq = 0
		snowflakeMS = now
	}

	// 13(ms) + 4(node) + 6(seq) + 9(pad from ms low bits) = 32 digits
	partC := (now / 1000) % 1000000000
	return fmt.Sprintf("%013d%04d%06d%09d", now, snowflakeNodeID, snowflakeSeq, partC)
}
