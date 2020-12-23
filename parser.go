package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	TLDCode = "de"
	MaxPage = 1
)

// Trial is a trial result from the eu-ctr
type Trial struct {
	startDate             string
	name                  string
	sponsorName           string
	sponsorProtocolNumber string
	genders               string
	eudraCTNumber         string
	fullTitle             string
	populationAge         string
	trialprotocol         string
	trialResults          string
	disease               string
	medicalCondition      string
}

func getPage(page int) ([]Trial, error) {
	fmt.Printf("downloading page %d\n", page)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	res, err := http.Get(fmt.Sprintf("https://www.clinicaltrialsregister.eu/ctr-search/search?query=&country=%s&page=%d", TLDCode, page))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	trials := []Trial{}

	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		current := Trial{}
		s.Find("td").Each(func(ii int, sel *goquery.Selection) {
			found := strings.TrimSpace(sel.Contents().Text())
			foundComponents := strings.Split(found, ":")
			switch foundComponents[0] {
			case "EudraCT Number":
				current.eudraCTNumber = strings.TrimSpace(foundComponents[1])
			case "Disease":
				current.disease = strings.TrimSpace(foundComponents[1])
			case "Start Date*":
				current.startDate = strings.TrimSpace(foundComponents[1])
			case "Sponsor Name":
				current.sponsorName = strings.TrimSpace(foundComponents[1])
			case "Full Title":
				current.fullTitle = strings.TrimSpace(foundComponents[1])
			case "Medical condition":
				current.medicalCondition = strings.TrimSpace(foundComponents[1])
			case "Population Age":
				current.populationAge = strings.TrimSpace(foundComponents[1])
			case "Gender":
				current.genders = strings.TrimSpace(foundComponents[1])
			case "Sponsor Protocol Number":
				current.sponsorProtocolNumber = strings.TrimSpace(foundComponents[1])
			case "Trial results":
				current.trialResults = strings.TrimSpace(foundComponents[1])
			case "Trial protocol":
				current.trialprotocol = strings.TrimSpace(foundComponents[1])
			default:
				//fmt.Printf("encountered unparseable element: %s\n", found)
			}
		})
		trials = append(trials, current)
	})

	return trials, nil
}

func main() {
	trials := []Trial{}

	for counter := 0; counter < MaxPage; counter++ {
		partials, err := getPage(counter)
		if err != nil {
			log.Fatalf("failed to get page %d: %v", counter, err)
			break
		}
		trials = append(trials, partials...)
	}

	file, err := os.Create("eu-ctr.csv")
	if err != nil {
		log.Fatalf("failed to create output file")
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()
	headers := renderHeaders()
	if err := w.Write(headers); err != nil {
		log.Fatalf("failed to write headers")
	}

	for _, s := range trials {
		flat := flattenStruct(s)
		if err := w.Write(flat); err != nil {
			log.Fatalf("failed to write CSV row")
		}
	}
}

func renderHeaders() []string {
	return []string{
		"EudraCT Number",
		"Full Title",
		"Start Date",
		"Sponsor Name",
		"Sponsor Protocol Number",
		"Medical condition",
		"Population age",
		"Trial results",
		"Trial protocol",
		"Disease",
		"Gender",
	}
}

func flattenStruct(trial Trial) []string {
	out := []string{}
	out = append(out, trial.eudraCTNumber)
	out = append(out, trial.fullTitle)
	out = append(out, trial.startDate)
	out = append(out, trial.sponsorName)
	out = append(out, trial.sponsorProtocolNumber)
	out = append(out, trial.medicalCondition)
	out = append(out, trial.populationAge)
	out = append(out, trial.trialResults)
	out = append(out, trial.trialprotocol)
	out = append(out, trial.disease)
	out = append(out, trial.genders)
	return out
}
