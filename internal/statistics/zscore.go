package statistics

import (
	"bactic/internal/data"
	"database/sql"
	"log"
	"time"
)

type GaussianDist struct {
	mean     float32
	variance float32
}

// A type that dynamically comptues zscores based on the distribution of stats that have been called
type ZScoreCache struct {
	distns           map[data.EventType]GaussianDist
	dbUpdateInterval int // How often to update the relational database table
	currentUpdate    int
	db               *sql.DB
}

// Here is my general plan:
// 1. Each week, we compute the summary statistics across events, across divisions, and any other category one would want to compute
// 2. Apply exponential filtering to these computations with some chosen value
// 3. Create additional rules that prevent seasonality from requiring a really high smoothing factor
//   - we know when the seasons start and end, so we can create explicit time rules on this
func InitializeAndUpdateDistributions(tx *sql.Tx, now time.Time, delta time.Duration, smoothing float64) {
	// update any values where there are no existing distributions haven't
	_, err := tx.Exec(`
		WITH stats AS (
			SELECT AVG(result.quant) AS mean,
           		VARIANCE(result.quant) AS var,
             	heat.event_type AS event
            FROM result
            LEFT JOIN heat ON result.heat_id = heat.id
            LEFT JOIN meet ON meet.id = heat.meet_id
            WHERE $1 - meet.date <= $2
            GROUP BY event
        )
        INSERT INTO distns (mean, var, event)
        SELECT mean, var, event
        FROM stats
        ON CONFLICT (event) DO UPDATE
        	SET mean = $3 * EXCLUDED.mean + (1 - $3) * distns.mean,
         	var = $3 * EXCLUDED.var + (1 - $3) * distns.var`,
		now, delta, smoothing)
	if err != nil {
		log.Fatal("query error when trying to select results", err)
	}
}

// Compute the z-score of a result within its distribution.
func ComputeZScore(tx *sql.Tx, res data.Result) {

}

func NewZScoreCache(db *sql.DB) *ZScoreCache {
	rows, err := db.Query("SELECT event, mean, variance FROM distns")
	if err != nil {
		panic("Colud not query database to setup zscore cache")
	}

	var (
		event    data.EventType
		mean     float32
		variance float32
	)
	d := make(map[data.EventType]GaussianDist)
	for rows.Next() {
		rows.Scan(&event, &mean, &variance)
		d[event] = GaussianDist{
			mean,
			variance,
		}
	}

	return &ZScoreCache{
		db:               db,
		dbUpdateInterval: 100,
		currentUpdate:    0,
		distns:           d,
	}

}

// Compute a global zscore across all the distributions of an athlete's performance'
// TODO: create and document an algorithm here. Check out welford's online algorithm at https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance '
//func (z *ZScoreCache) GetZScoreAndUpdate(map[internal.EventType]internal.Result) float32 {
//log.Fatalf("Not Implemented!")

//	// update database distribution
//	if z.currentUpdate > z.dbUpdateInterval {
//		z.currentUpdate = 0
//		tx, err := z.db.Begin()
//		if err != nil {
//			panic(err)
//		}
//
//		for k, v := range z.distns {
//			row := tx.QueryRow("SELECT size FROM distns WHERE event = $1", k)
//			var size int
//			if err := row.Scan(&size); err != nil {
//				panic(err)
//			}
//			_, err := tx.Exec("UPDATE distns SET mean = $1, variance = $2 WHERE event = $3", v.mean, v.variance, k)
//			if err != nil {
//				panic(err)
//			}
//		}
//
//	}
//}
