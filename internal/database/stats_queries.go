package database

import "bactic/internal"

// Return a bucketing of data from table into nBuckets
func (db *BacticDB) Histogram(eventType internal.EventType, nBuckets int) []int {
	hist := make([]int, nBuckets)
	var (
		low  float32
		high float32
	)
	row := db.DBConn.QueryRow("SELECT MAX(r.quant) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type=?", eventType)
	err := row.Scan(&high)
	if err != nil {
		panic(err)
	}

	row = db.DBConn.QueryRow("SELECT MIN(r.quant) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.event_type=?", eventType)
	err = row.Scan(&low)
	if err != nil {
		panic(err)
	}

	inc := (high - low) / float32(nBuckets)
	for i := 0; i < nBuckets; i++ {
		row = db.DBConn.QueryRow("SELECT COUNT(r.id) FROM result r LEFT JOIN heat h ON r.heat_id = h.id WHERE h.EVENT_TYPE=? AND r.quant >= ? AND r.quant < ?", eventType, inc*float32(i), inc*float32(i+1))
		err = row.Scan(&hist[i])
		if err != nil {
			panic(err)
		}
	}
	return hist
}

func (db *BacticDB) PersonalRecord(eventType internal.EventType, athID uint32) float32 {
	// row, err := db.DBConn.Query("SELECT ")
	return 0.0
}

func (db *BacticDB) PersonalHistory(eventType internal.EventType, athID uint32) []float32 {
	panic("Not implemented!")
	return []float32{0.0}
}
