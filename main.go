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
	VERSION        = "0.6.0"
	BUILD_DATE     = "2024-Aug-5"
	FLAG_DAYS      = "-days="
	FLAG_JSON      = "-json"
	FLAG_CSV       = "-csv"
	FLAG_NO_COLOR  = "-no-color"
	FLAG_NO_OUTPUT = "-no-output"
	FLAG_SHORT     = "-short"
	FLAG_NO_HEADER = "-no-header"
	FLAG_TIMEOUT   = "-timeout="
)

var (
	returnCode          = 0
	dateThreshold       time.Time
	enableTerminalColor = true
	enableHeader        = true
	timeoutSeconds      = checkssl.DEFAULT_TIMEOUT_SEC
	outputFormat        = checkssl.TEXT
)

func main() {
	arguments := separateCommandLineArgumentsFromFlags()
	if noTargetsWereGiven(arguments) {
		displayHelpText("")
	}

	if outputFormat == checkssl.CSV && enableHeader {
		fmt.Println(checkssl.CsvHeaderRow())
	}

	a := checkssl.NewCheckSSL()
	a.SetThreshold(dateThreshold)
	a.SetTimeout(timeoutSeconds)

	for i := range arguments {
		result := a.CheckServer(arguments[i], false)
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
			if strings.HasPrefix(value, FLAG_NO_HEADER) {
				enableHeader = false
			}
			if strings.HasPrefix(value, FLAG_TIMEOUT) {
				parsableTimeout := strings.Replace(value, FLAG_TIMEOUT, "", 1)
				seconds, _ := strconv.ParseInt(parsableTimeout, 10, 32)
				timeoutSeconds = int(seconds)
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
	fmt.Println("  -no-header (will disable the header row in CSV output)")
	fmt.Println("  -short (will show only 1 line per result)")
	fmt.Println("  -timeout=5 (will set the timeout to 5 seconds)", " default =", checkssl.DEFAULT_TIMEOUT_SEC)
}
