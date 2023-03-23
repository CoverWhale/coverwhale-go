package main

import (
	"fmt"
	"log"

	"github.com/CoverWhale/coverwhale-go/grist"
)

func main() {
	client := grist.NewClient(
		grist.SetURL("http://localhost:8484"),
		grist.SetAPIKey("api-key"),
	)

	data, err := client.GetRecords("doc-id", "My_table")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(data))
}
