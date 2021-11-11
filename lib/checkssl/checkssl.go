package checkssl

import (
	"errors"
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const RETURNCODE_PASS = 0
const RETURNCODE_EXPIRED = 2
const RETURNCODE_THRESHOLDFAIL = 3

type CheckedServer struct {
	Target string
	Err string
	ExitCode int
	ServerInfo string
	Certs []CheckCert
	Passed bool
}
type CheckCert struct {
	CommonName string
	IsCertificateAuthority bool
	ValidNotBefore time.Time
	ValidNotAfter time.Time
	IsInvalid bool
}

func CheckServer(target string, dateThreshold time.Time) (output CheckedServer) {
	target = strings.Replace(target, "http://", "https://", 1)
	if !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	output.Target = target
	output.Passed = true

	response, err := http.Head(target)
	if err != nil {
		certError := errors.Unwrap(err)
		output.Err = certError.Error()
		output.Passed = false
		output.ExitCode = RETURNCODE_EXPIRED
		return
	}

	output.ServerInfo += response.Header.Get("Server")
	if output.ServerInfo != "" { output.ServerInfo += " - "}
	output.ServerInfo += response.Header.Get("x-powered-by")

	for _, val := range response.TLS.VerifiedChains[0] {
		certInfo := CheckCert{}
		certInfo.IsCertificateAuthority = val.IsCA
		certInfo.ValidNotAfter = val.NotAfter
		certInfo.ValidNotBefore = val.NotBefore

		commonName := val.Subject.CommonName
		if commonName == "" {
			commonName = "(missing common name)"
		}
		certInfo.CommonName = commonName

		newCode := checkIfExpirationIsWithinTolerance(dateThreshold, val.NotBefore, val.NotAfter)
		if newCode > RETURNCODE_PASS {
			certInfo.IsInvalid = true
			output.ExitCode = newCode
			output.Passed = false
		}
		output.Certs = append(output.Certs, certInfo)

	}

	return
}

func checkIfExpirationIsWithinTolerance(dateThreshold time.Time, notBefore time.Time, notAfter time.Time) int {
	if dateThreshold.After(notAfter) {
		return RETURNCODE_THRESHOLDFAIL
	}

	if notBefore.After(time.Now()) {
		return RETURNCODE_EXPIRED
	}

	return RETURNCODE_PASS
}

func displayDate(input time.Time) string {
	numberOfDays := fmt.Sprintf("%.1f",input.Sub(time.Now()).Hours() / 24)
	//	Mon Jan 2 15:04:05 -0700 MST 2006
	return input.Format("2006-01-02 3:04PM Mon") + " (" + numberOfDays + " days)"
}

func (a CheckedServer) AsString() (output string) {
	output += fmt.Sprintf("\n%s\n", a.Target)

	output += fmt.Sprintf("%s\n", a.ServerInfo)

	for i, cert := range a.Certs {

		if cert.IsCertificateAuthority {
			output += fmt.Sprintf(" CA-%d) ", i+1)
		}else{
			output += fmt.Sprintf(" %d) ", i+1)
		}

		output += fmt.Sprintf("%s expires on %s", cert.CommonName, displayDate(cert.ValidNotAfter))

		if cert.IsInvalid {
			output += fmt.Sprintf("\n     ↳ [FAIL] expires in %s", cert.ValidNotAfter.Sub(time.Now()))
		}
		output += fmt.Sprintf("\n")
	}

	if a.Passed {
		output += fmt.Sprintf("[PASS] %s\n", a.Target)
	}else{
		output += fmt.Sprintf(" %v\n", a.Err)
		output += fmt.Sprintf("[FAIL] %s\n", a.Target)
	}

	return
}

func (a CheckedServer) AsJson() string {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "{ \"error\": \"Unable to convert result to json\"}"
	}
	return string(jsonBytes)
}