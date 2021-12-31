package checkssl

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const RETURNCODE_PASS = 0
const RETURNCODE_EXPIRED = 2
const RETURNCODE_THRESHOLDFAIL = 3

type CheckedServer struct {
	Target       string
	Err          string
	ExitCode     int
	ServerInfo   string
	Certs        []CheckCert
	Passed       bool
	HttpVersion  string
	TlsVersion   uint16
	TlsAlgorithm uint16
	ServerName   string
}
type CheckCert struct {
	CommonName             string
	IsCertificateAuthority bool
	ValidNotBefore         time.Time
	ValidNotAfter          time.Time
	IsInvalid              bool
}

func CheckServer(target string, dateNeededValidFor time.Time) (output CheckedServer) {
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
	if output.ServerInfo != "" {
		output.ServerInfo += " - "
	}
	output.ServerInfo += response.Header.Get("x-powered-by")
	output.ServerName = response.TLS.ServerName
	output.TlsVersion = response.TLS.Version
	output.TlsAlgorithm = response.TLS.CipherSuite
	output.HttpVersion = response.TLS.NegotiatedProtocol

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
		if output.ServerName == "" {
			output.ServerName = commonName
		}

		newCode := checkIfExpirationIsWithinTolerance(dateNeededValidFor, val.NotBefore, val.NotAfter)
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
	numberOfDays := fmt.Sprintf("%.1f", input.Sub(time.Now()).Hours()/24)
	//	Mon Jan 2 15:04:05 -0700 MST 2006
	return input.Format("2006-01-02 3:04PM Mon") + " (" + numberOfDays + " days)"
}

func (a CheckedServer) AsString(enableColors bool) (output string) {
	setTerminalColor(enableColors)

	if a.Err == "" {
		output += fmt.Sprintf("\n%s\n", a.ServerName)

		output += fmt.Sprintf(" -> %s\n", expandServerNames(a.ServerInfo))
		output += fmt.Sprintf(" -> %s with %s\n", getHttpVersion(a.HttpVersion), getTlsVersion(a.TlsVersion))
		output += fmt.Sprintf(" -> %s %s\n", getTlsAlgo(a.TlsAlgorithm), getMozillaRecommendedCipher(a.TlsAlgorithm))
	}
	for i, cert := range a.Certs {

		if cert.IsCertificateAuthority {
			output += fmt.Sprintf(" CA-%d) ", i+1)
		} else {
			output += fmt.Sprintf(" %d) ", i+1)
		}

		output += fmt.Sprintf("%s expires on %s", cert.CommonName, displayDate(cert.ValidNotAfter))

		if cert.IsInvalid {
			output += fmt.Sprintf("\n     ↳ [FAIL] expires in %s", cert.ValidNotAfter.Sub(time.Now()))
		}
		output += fmt.Sprintf("\n")
	}

	if a.Passed {
		output += fmt.Sprintf("%s[PASS]%s %s\n", terminalGreen, terminalNoColor, a.Target)
	} else {
		output += fmt.Sprintf("%s %v\n", a.Target, a.Err)
		output += fmt.Sprintf("%s[FAIL]%s %s\n", terminalRed, terminalNoColor, a.Target)
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

var terminalNoColor = ""
var terminalRed = ""
var terminalYellow = ""
var terminalGreen = ""

func setTerminalColor(enable bool) {
	if enable {
		terminalNoColor = "\033[0m"
		terminalRed = "\033[31m"
		terminalYellow = "\033[0;33m"
		terminalGreen = "\033[0;32m"
	} else {
		terminalNoColor = ""
		terminalRed = ""
		terminalYellow = ""
		terminalGreen = ""
	}
}

func getTlsVersion(input uint16) string {
	switch input {
	case tls.VersionSSL30:
		return terminalRed + "SSL v3.0 () - PLEASE UPGRADE to TLS v1.2" + terminalNoColor
	case tls.VersionTLS10:
		return terminalRed + "TLS v1.0 (released 1999) - PLEASE UPGRADE to TLS v1.2 or v1.3" + terminalNoColor
	case tls.VersionTLS11:
		return terminalRed + "TLS v1.1 (release 2006) - PLEASE UPGRADE to TLS v1.2 or v1.3" + terminalNoColor
	case tls.VersionTLS12:
		return terminalYellow + "TLS v1.2 (released 2008) - Consider upgrading to TLS v1.3" + terminalNoColor
	case tls.VersionTLS13:
		return terminalGreen + "TLS v1.3 (released 2018) - latest version" + terminalNoColor
	}

	return fmt.Sprintf("unknown TLS version: %d", input)
}

func getMozillaRecommendedCipher(input uint16) string {
	// Mozilla Recommended Ciphers - https://ssl-config.mozilla.org/
	// ECDHE-ECDSA-AES128-GCM-SHA256
	// ECDHE-RSA-AES128-GCM-SHA256
	// ECDHE-ECDSA-AES256-GCM-SHA384
	// ECDHE-RSA-AES256-GCM-SHA384
	// ECDHE-ECDSA-CHACHA20-POLY1305
	// ECDHE-RSA-CHACHA20-POLY1305
	// DHE-RSA-AES128-GCM-SHA256 ?
	// DHE-RSA-AES256-GCM-SHA384 ?
	if input == tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 ||
		input == tls.TLS_RSA_WITH_AES_128_GCM_SHA256 ||
		input == tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 ||
		input == tls.TLS_RSA_WITH_AES_256_GCM_SHA384 ||
		input == tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305 ||
		input == tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305 {
		return terminalGreen + " (Mozilla Recommended Cipher)" + terminalNoColor
	}
	return "" // not insecure, but consider upgrading
}

func getTlsAlgo(input uint16) string {
	switch input {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256 = TLS, message encrypted with AES128 GCM, hashes are SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 = TLS, key exchange is ECDHE, auth exchange is ECDSA, messages encrypted with AES128GCM, hashes are SHA256"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		return "TLS_RSA_WITH_AES_256_CBC_SHA - !!! WEAK Cipher, please upgrade !!!"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	}
	return fmt.Sprintf("unknown TLS algo: %x", input)

}

func getHttpVersion(input string) string {

	switch input {
	case "h3":
		return "HTTP/3" // currently, go cannot do http3 natively
	case "h2":
		return terminalGreen + "HTTP/2" + terminalNoColor
	case "http/1.1":
		return terminalRed + "HTTP/1.1 (OLD)" + terminalNoColor
	case "http/1":
		return terminalRed + "HTTP/1 (OLD)" + terminalNoColor
	}
	return fmt.Sprintf("unknown HTTPS version: %s", input)

}

func expandServerNames(input string) (output string) {

	output = strings.Replace(input, "ATS", "Apache Traffic Server (ATS)", 1)
	output = strings.Replace(output, "gws", "Google Web Server (gws)", 1)

	return
}
