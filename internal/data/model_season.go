package data

import "log"

type Season uint32

const (
	XC Season = iota << 1
	INDOOR
	OUTDOOR
)

func (s Season) IsValid() bool {
	return s >= XC && s <= OUTDOOR
}

var seasonToString = map[Season]string{
	XC:      "XC",
	INDOOR:  "Indoor",
	OUTDOOR: "Outdoor",
}

func (s Season) String() string {
	if !s.IsValid() {
		log.Fatalf("Season value %d was invalid", uint32(s))
	}
	return seasonToString[s]
}
