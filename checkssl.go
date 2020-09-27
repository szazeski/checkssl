package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)
const VERSION = "0.2" // 9/26/2020
const RETURNCODE_PASS = 0
const RETURNCODE_EXPIRED = 2
const RETURNCODE_THRESHOLDFAIL = 3
const FLAG_DAYS= "-days="

var returnCode = 0
var daysThreshold = 0
var dateThreshold time.Time


func main() {
	arguments := separateCommandLineArgumentsFromFlags()
	if noTargetsWereGiven(arguments) {
		displayHelpText("")
	}

	for i := range arguments {
		checkSsl(arguments[i])
	}
	os.Exit(returnCode)
}

func noTargetsWereGiven(arguments []string) bool {
	return len(arguments) == 0
}

func checkSsl(target string) {
	fmt.Printf("\n%s\n", target)

	if !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	target = strings.Replace(target, "http://", "https://", 1)

	response, err := http.Head(target)
	if err != nil {
		//fmt.Printf(" %s\n", err)
		certError := errors.Unwrap(err)
		fmt.Printf(" %v\n", certError)
		fmt.Printf("[FAIL] %s\n", target)
		os.Exit(RETURNCODE_EXPIRED)
	}

	outputServerInformation(response)

	for i, val := range response.TLS.VerifiedChains[0] {
		if val.IsCA {
			fmt.Printf(" CA-%d) ", i+1)
		}else{
			fmt.Printf(" %d) ", i+1)
		}
		commonName := val.Subject.CommonName
		if commonName == "" {
			commonName = "(missing common name)"
		}
		fmt.Printf("%s expires on %s", commonName, displayDate(val.NotAfter))

		newCode := checkIfExpirationIsWithinTolerance(val.NotBefore, val.NotAfter)
		if newCode > RETURNCODE_PASS {
			fmt.Printf("\n     ↳ [FAIL] expires before %d days", daysThreshold)
			returnCode = newCode
		}

		fmt.Printf("\n")
	}
	if returnCode == RETURNCODE_PASS {
		fmt.Printf("[PASS] %s\n", target)
	}else{
		fmt.Printf("[FAIL] %s\n", target)
	}
}

func separateCommandLineArgumentsFromFlags() []string {
	var arguments []string
	for i, value := range os.Args {
		if i == 0 {
			continue
		}
		if strings.HasPrefix(value, "-") {

			if strings.HasPrefix(value, FLAG_DAYS) {
				parsableDays := strings.Replace(value, FLAG_DAYS, "", 1)
				parsedDays,_ := strconv.ParseInt(parsableDays,10, 32)
				daysThreshold = int(parsedDays)

				offset, _ := time.ParseDuration(strconv.Itoa(daysThreshold * 24) + "h")
				dateThreshold = time.Now().Add(offset)
			}

			continue
			// this allows flags to be mixed into the arguments
		}
		arguments = append(arguments, value)
	}

	return arguments
}

func outputServerInformation(response *http.Response) {
	fmt.Printf("%s", response.Header.Get("Server"))
	fmt.Printf("%s", response.Header.Get("x-powered-by"))
	fmt.Printf("\n")
}

func displayDate(input time.Time) string {
	numberOfDays := fmt.Sprintf("%.1f",input.Sub(time.Now()).Hours() / 24)
	//	Mon Jan 2 15:04:05 -0700 MST 2006
	return input.Format("2006-01-02 3:04PM Mon") + " (" + numberOfDays + " days)"
}

func displayHelpText(errorText string) {
	if errorText != "" {
		fmt.Println(errorText)
	}

	fmt.Println("checkssl [url] ")
	fmt.Println(" easy to read/parse information about ssl certificates")
	fmt.Println(" version " + VERSION)
	fmt.Println("  -days=5 (will fail the check if the cert is within 5 days of renewal)")
}

func checkIfExpirationIsWithinTolerance(notBefore time.Time, notAfter time.Time) int {
	if dateThreshold.After(notAfter) {
		return RETURNCODE_THRESHOLDFAIL
	}

	if notBefore.After(time.Now()) {
		return RETURNCODE_EXPIRED
	}

	return RETURNCODE_PASS
}
