package extractors

import "charex/internal/core"

// Extractor defines the interface for all character extractors.
// Each extractor is responsible for parsing data from a specific source
// and converting it into a standardized TavernCardV2 format.
type Extractor interface {
	// Extract processes the given input data (e.g., a URL or a JSON body)
	// and returns a populated TavernCardV2 object, the raw data used for
	// the extraction, a byte slice for a character image if found, and an error
	// if the process fails.
	Extract(input []byte) (card *core.TavernCardV2, rawData []byte, cardImage []byte, err error)
}