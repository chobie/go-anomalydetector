package anomalydetector

import (
	"strconv"
	"testing"
	"encoding/csv"
	"log"
	"os"
	"io"
)

func TestAnomalyDetector(t *testing.T) {
	expected := "3.209952"

	v := NewAnomalyDetector(5, 5.0)
	score := strconv.FormatFloat(v.Update(1.0), 'f', 6, 64)

	if score != expected {
		t.Errorf("Got %s, want %s", score, expected)
	}
}

func TestAnomalyDetector2(t *testing.T) {
	file, err := os.Open("stock.2432.csv")
	if err != nil {
		log.Fatal("Error: %s", err)
	}

	reader := csv.NewReader(file)
	_, _ = reader.Read()

	result := 0.0
	a := NewAnomalyDetector(28, 0.05)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else {
			if err != nil {
				log.Fatal("Error: %s", err)
			}
		}

		value , err := strconv.ParseFloat(record[1], 64);
		if err != nil {
			log.Fatal("decode failed: %s", err)
			break
		}
		result = a.Update(value)
		log.Printf("result: %f (%f)\n", result, value)
	}

	log.Printf("result: %f\n", result)
}
