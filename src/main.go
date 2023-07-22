package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/jgfrazie/scp_webscraper/src/scraper"
)

var numSeries int = len(*scraper.AcquireAllSCPSeries())

// TODO: IMPLEMENT A CLI INTERFACE FOR ACCESSING SCPS

// Checks if first arg is an SCP request and if so returns the SCP item number
func isSCPRequest(CLArg string) (int, error) {
	scpNum, isRequest := strconv.Atoi(CLArg)
	if isRequest != nil {
		return 0, errors.New("is not an SCP Entity request")
	}

	return scpNum, isRequest
}

func checkErrorRate(series int) {
	checkSeries := scraper.CollectSCPs(&[]string{(*scraper.AcquireAllSCPSeries())[series - 1]})
	errorRate := scraper.ScraperErrorRate(checkSeries, 0)

	fmt.Println()
	fmt.Printf("The error rate for Series %v is %v\n", series, errorRate)
}

func performSER() {
	if len(os.Args[1:]) <= 1 {
		panic("too few arguements passed for 'ser'. Make sure to pass an SCP series")
	}

	series, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	if series < 1 || series > numSeries {
		panic("invalid series request")
	}

	checkErrorRate(series)
}

func performRANGE(scpNum int, argsWithProg []string) {
	endSCP, err := strconv.Atoi(argsWithProg[1])
		if err != nil {
			panic(err)
		}
		fmt.Println(scpNum, endSCP)
		scps := scraper.GetRange(scpNum, endSCP)
		for _, scp := range *scps {
			fmt.Println(scp.String())
		}
}

// Controls the flow of the program
func main() {
	defer fmt.Println("REQUEST COMPLETE")
	argsWithProg := os.Args[1:]

	if argsWithProg[0] == "ser" {
		fmt.Println("PROCESSING REQUEST: Performing Series Error Rate Service. This may take a moment...")
		performSER()
		return
	}

	scpNum, isRequest := isSCPRequest(argsWithProg[0])
	if isRequest != nil {
		panic(isRequest.Error())
	} else if len(argsWithProg) > 1 {
		fmt.Println("PROCESSING REQUEST: Acquiring set of requested SCPs. This may take a moment...")
		performRANGE(scpNum, argsWithProg)
		return
	}

	fmt.Printf("PROCESSING REQUEST: Acquiring SCP-%v.\n\n", scpNum)
	scpEntity := scraper.GetSCP(scpNum)
	fmt.Println(scpEntity.String())
}
