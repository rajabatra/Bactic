package database

import (
	"bactic/internal"
	"database/sql"
)

// Return a bucketing of data from table into nBuckets
func Histogram(db *sql.DB, eventType internal.EventType, nBuckets int) []int {
	if nBuckets <= 0 {
		panic("nBuckets must be greater than zero")
	}
	hist := make([]int, nBuckets)
	var (
		low  float32
		high float32
	)
	row := db.QueryRow("SELECT MAX(r.quant) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type=?", eventType)
	err := row.Scan(&high)
	if err != nil {
		panic(err)
	}

	row = db.QueryRow("SELECT MIN(r.quant) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type=?", eventType)
	err = row.Scan(&low)
	if err != nil {
		panic(err)
	}

	inc := (high - low) / float32(nBuckets)
	for i := 0; i < nBuckets-1; i++ {
		row = db.QueryRow("SELECT COUNT(r.id) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.EVENT_TYPE=? AND r.quant >= ? AND r.quant < ?", eventType, inc*float32(i), inc*float32(i+1))
		if err = row.Scan(&hist[i]); err != nil {
			panic(err)
		}
	}
	row = db.QueryRow("SELECT COUNT(r.id) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.EVENT_TYPE=? AND r.quant >= ? AND r.quant <= ?", eventType, inc*float32(nBuckets-1), inc*float32(nBuckets))
	if err = row.Scan(&hist[nBuckets-1]); err != nil {
		panic(err)
	}

	return hist
}

func PersonalRecord(db *sql.DB, eventType internal.EventType, athID uint32) float32 {
	panic("Not implemented!")
}

func PersonalHistory(db *sql.DB, eventType internal.EventType, athID uint32) []float32 {
	panic("Not implemented!")
}
