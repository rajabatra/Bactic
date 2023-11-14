package athnet

import (
	"bactic/internal/database"
	"log"
	"os"
	"sync"
	"time"
)

func NewAthnetCollector(db *database.BacticDB, sig chan os.Signal, wg *sync.WaitGroup, scrapeLoop time.Duration) {
	log.Fatal("Athnet collector not yet implemented")
}
