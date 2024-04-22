package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/szazeski/checkssl/lib/checkssl"
)

const (
	VERSION        = "0.5.1"
	BUILD_DATE     = "2024-Apr-21"
	FLAG_DAYS      = "-days="
	FLAG_JSON      = "-json"
	FLAG_CSV       = "-csv"
	FLAG_NO_COLOR  = "-no-color"
	FLAG_NO_OUTPUT = "-no-output"
	FLAG_SHORT     = "-short"
)

var (
	returnCode          = 0
	dateThreshold       time.Time
	enableTerminalColor = true
	outputFormat        = checkssl.TEXT
)

func main() {
	arguments := separateCommandLineArgumentsFromFlags()
	if noTargetsWereGiven(arguments) {
		displayHelpText("")
	}

	if outputFormat == checkssl.CSV {
		fmt.Println(checkssl.CsvHeaderRow())
	}

	for i := range arguments {
		result := checkssl.CheckServer(arguments[i], dateThreshold, false)
		returnCode += result.ExitCode
		if outputFormat == checkssl.JSON {
			fmt.Println(result.AsJson())
		} else if outputFormat == checkssl.CSV {
			fmt.Println(result.AsCsv())
		} else if outputFormat == checkssl.TEXT {
			fmt.Println(result.AsString(enableTerminalColor))
		} else if outputFormat == checkssl.SHORT {
			fmt.Print(result.AsShortString(enableTerminalColor))
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
				outputFormat = checkssl.JSON
			}
			if strings.HasPrefix(value, FLAG_CSV) {
				outputFormat = checkssl.CSV
			}
			if strings.HasPrefix(value, FLAG_SHORT) {
				outputFormat = checkssl.SHORT
			}
			if strings.HasPrefix(value, FLAG_NO_OUTPUT) {
				outputFormat = checkssl.NONE
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

	fmt.Println("checkssl [url] [url] [url] ...")
	fmt.Println(" easy to read/parse information about ssl certificates")
	fmt.Println(" version " + VERSION + " built " + BUILD_DATE)
	fmt.Println("  -days=5 (will fail the check if the cert is within 5 days of renewal)")
	fmt.Println("  -json (will output in JSON format)")
	fmt.Println("  -csv (will output in comma seperated format for spreadsheets)")
	fmt.Println("  -no-color (will disable color syntax from output)")
	fmt.Println("  -no-output (will only produce exit code)")
	fmt.Println("  -short (will show only 1 line per result)")
}
