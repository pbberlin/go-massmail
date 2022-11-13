package main

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

// Client is an example struct, you can use "-" to ignore a field
type Client struct {
	Id            string `csv:"client_id"`
	Name          string `csv:"client_name"`
	Age           string `csv:"client_age"`
	NotUsedString string `csv:"-"`
}

func ReadCSVExample() {

	inFile, err := os.OpenFile(
		"./csv/recipients.csv",
		os.O_RDWR|os.O_CREATE,
		os.ModePerm,
	)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(
		"./csv/recipients_out.csv",
		os.O_RDWR|os.O_CREATE,
		os.ModePerm,
	)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	clients := []*Client{}

	// load clients from file
	if err := gocsv.UnmarshalFile(inFile, &clients); err != nil {
		panic(err)
	}

	for _, client := range clients {
		fmt.Println("Hello", client.Name)
	}

	// go to start of file
	if _, err := inFile.Seek(0, 0); err != nil {
		panic(err)
	}

	clients = append(clients, &Client{Id: "12", Name: "John", Age: "21"}) // Add clients
	clients = append(clients, &Client{Id: "13", Name: "Fred"})
	clients = append(clients, &Client{Id: "14", Name: "James", Age: "32"})
	clients = append(clients, &Client{Id: "15", Name: "Danny"})

	// get all clients as CSV string
	csvContent, err := gocsv.MarshalString(&clients)
	if err != nil {
		panic(err)
	}
	fmt.Println(csvContent)

	// save the CSV back to the file
	err = gocsv.MarshalFile(&clients, outFile)
	if err != nil {
		panic(err)
	}

}
