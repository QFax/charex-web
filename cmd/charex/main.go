package main

import (
	"charex/internal/extractors"
	"charex/internal/saver"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Define command-line flags.
	extractorType := flag.String("type", "", "The type of extractor to use ('sakura' or 'janitor').")
	inputFile := flag.String("input", "", "Path to the input file (containing a URL for sakura, or JSON for janitor).")
	outputDir := flag.String("output", "output", "Directory to save the output files.")
	flag.Parse()

	// Validate flags.
	if *extractorType == "" || *inputFile == "" {
		fmt.Println("Usage: go run cmd/charex/main.go --type=<sakura|janitor> --input=<filepath>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read the input file.
	inputData, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

	// Select the extractor based on the type flag.
	var extractor extractors.Extractor
	switch *extractorType {
	case "sakura":
		extractor = extractors.NewSakuraFMExtractor()
	case "janitor":
		extractor = extractors.NewJanitorAIExtractor()
	default:
		log.Fatalf("Unknown extractor type: %s", *extractorType)
	}

	// Run the extraction process.
	log.Printf("Running %s extractor...", *extractorType)
	card, rawData, cardImage, err := extractor.Extract(inputData)
	if err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}
	log.Println("Extraction successful.")

	// Save the card.
	log.Printf("Saving card to directory: %s", *outputDir)
	if err := saver.SaveCard(card, rawData, cardImage, *outputDir); err != nil {
		log.Fatalf("Failed to save card: %v", err)
	}

	log.Println("Card saved successfully!")
}