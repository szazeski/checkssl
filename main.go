package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/szazeski/checkssl/lib/checkssl"
)

const VERSION = "0.4.3"
const BUILD_DATE = "2022-June-19"
const FLAG_DAYS = "-days="
const FLAG_JSON = "-json"
const FLAG_NO_COLOR = "-no-color"

var returnCode = 0
var dateThreshold time.Time
var outputAsJson = false
var enableTerminalColor = true

func main() {
	arguments := separateCommandLineArgumentsFromFlags()
	if noTargetsWereGiven(arguments) {
		displayHelpText("")
	}

	for i := range arguments {
		result := checkssl.CheckServer(arguments[i], dateThreshold, false)
		returnCode += result.ExitCode
		if outputAsJson {
			fmt.Println(result.AsJson())
		} else {
			fmt.Println(result.AsString(enableTerminalColor))
		}
	}
	os.Exit(returnCode)
}

func noTargetsWereGiven(arguments []string) bool {
	return len(arguments) == 0
}

func separateCommandLineArgumentsFromFlags() []string {
	var arguments []string
	dateThreshold = time.Now()
	for i, value := range os.Args {
		if i == 0 {
			continue
		}
		if strings.HasPrefix(value, "-") {
			value = strings.Replace(value, "--", "-", 1)
			if strings.HasPrefix(value, FLAG_DAYS) {
				parsableDays := strings.Replace(value, FLAG_DAYS, "", 1)
				parsedDays, _ := strconv.ParseInt(parsableDays, 10, 32)
				daysThreshold := int(parsedDays)

				offset, _ := time.ParseDuration(strconv.Itoa(daysThreshold*24) + "h")
				dateThreshold = time.Now().Add(offset)
			}
			if strings.HasPrefix(value, FLAG_JSON) {
				outputAsJson = true
			}
			if strings.HasPrefix(value, FLAG_NO_COLOR) {
				enableTerminalColor = false
			}
			continue
			// this allows flags to be mixed into the arguments
		}
		arguments = append(arguments, value)
	}

	return arguments
}

func displayHelpText(errorText string) {
	if errorText != "" {
		fmt.Println(errorText)
	}

	fmt.Println("checkssl [url] ")
	fmt.Println(" easy to read/parse information about ssl certificates")
	fmt.Println(" version " + VERSION + " built " + BUILD_DATE)
	fmt.Println("  -days=5 (will fail the check if the cert is within 5 days of renewal)")
	fmt.Println("  -json (will output in JSON format)")
	fmt.Println("  -no-color (will disable color syntax from output)")
}
