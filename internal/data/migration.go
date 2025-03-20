package data

import "time"

type Migration struct {
	Version   uint64
	Name      string
	Checksum  string
	AppliedAt time.Time
}
