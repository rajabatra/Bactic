package athnet

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"time"
)

func NewAthnetCollector(db *sql.DB, sig chan os.Signal, wg *sync.WaitGroup, scrapeLoop time.Duration) {
	log.Fatal("Athnet collector not yet implemented")
}
