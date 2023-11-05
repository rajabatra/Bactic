package athnet

import (
	"log"
	"os"
	"scraper/internal/database"
	"sync"
)

func NewAthnetCollector(db *database.BacticDB, sig chan os.Signal, wg *sync.WaitGroup) {
	log.Fatal("Athnet collector not yet implemented")
}
