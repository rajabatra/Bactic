package athnet

import (
	"bactic/internal/database"
	"log"
	"os"
	"sync"
)

func NewAthnetCollector(db *database.BacticDB, sig chan os.Signal, wg *sync.WaitGroup) {
	log.Fatal("Athnet collector not yet implemented")
}
