package main

import (
	"flag"
	"fmt"
	"github.com/stockyard-dev/stockyard-cartograph/internal/server"
	"github.com/stockyard-dev/stockyard-cartograph/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
	portFlag := flag.String("port", "", "")
	dataFlag := flag.String("data", "", "")
	flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9230"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./cartograph-data"
	}

	if *portFlag != "" {
		port = *portFlag
	}
	if *dataFlag != "" {
		dataDir = *dataFlag
	}
	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("cartograph: open database: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits(), dataDir)

	fmt.Printf("\n  Cartograph — Self-hosted sitemap generator\n")
	fmt.Printf("  Questions? hello@stockyard.dev\n")
	fmt.Printf("  ─────────────────────────────────\n")
	fmt.Printf("  Dashboard:  http://localhost:%s/ui\n", port)
	fmt.Printf("  API:        http://localhost:%s/api\n", port)
	fmt.Printf("  Data:       %s\n", dataDir)
	fmt.Printf("  ─────────────────────────────────\n\n")

	log.Printf("cartograph: listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Fatalf("cartograph: %v", err)
	}
}
