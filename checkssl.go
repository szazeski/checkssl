package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)
const VERSION = "0.1"

func main() {

	var arguments []string
	for _, value := range os.Args {
		if strings.HasPrefix(value, "-") {
			continue
			// this allows flags to be mixed into the arugments
		}
		arguments = append(arguments, value)
	}
	if len(arguments) >= 2 {
		target := arguments[1]
		fmt.Println(target)

		if !strings.HasPrefix(target, "https://") {
			target = "https://" + target
		}


		response, err := http.Head(target)
		if err !=nil {
			fmt.Printf(" %s\n", err)
			//fmt.Println(err.Cert)
			fmt.Println("[FAIL]")
			os.Exit(2) // bad cert
		}

		fmt.Printf("%s%s\n", response.Header.Get("Server"), response.Header.Get("x-powered-by"))

		for i, val := range response.TLS.VerifiedChains[0] {
			fmt.Printf(" %d) ", i+1)
			if val.IsCA {
				fmt.Printf("-CA- ")
			}
			commonName := val.Subject.CommonName
			if commonName == "" {
				commonName = "(missing common name)"
			}
			fmt.Printf("%s expires on %s", commonName, displayDate(val.NotAfter))

			fmt.Printf("\n")
		}

		fmt.Printf("[PASS] ★★★ \n")
		os.Exit(0)
	}




	//commandFlagIp := ""
	//flag.StringVar(&commandFlagIp, "ip", "", "set the ip address of the LED Controller")

	//flag.Parse()


	displayHelpText("")
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

	fmt.Println("aboutssl [url] ")
	fmt.Println(" easy to read/parse information about ssl certificates (version " + VERSION + ")")
}
