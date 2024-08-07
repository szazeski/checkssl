package checkssl

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"
)

const (
	RETURNCODE_PASS          = 0
	RETURNCODE_EXPIRED       = 2
	RETURNCODE_THRESHOLDFAIL = 3
	RETURNCODE_NOTVALIDYET   = 4
	RETURNCODE_ERROR         = 5

	dateLayout = "2006-01-02 3:04PM Mon"

	DEFAULT_TIMEOUT_SEC = 15
)

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
	IpAddress    string
}
type CheckCert struct {
	CommonName             string
	IsCertificateAuthority bool
	ValidNotBefore         time.Time
	ValidNotAfter          time.Time
	IsInvalid              bool
}

type CheckSSL struct {
	timeoutSeconds     int
	dateNeededValidFor time.Time
}

func NewCheckSSL() CheckSSL {
	return CheckSSL{
		timeoutSeconds:     DEFAULT_TIMEOUT_SEC,
		dateNeededValidFor: time.Now(),
	}
}
func (a *CheckSSL) SetTimeout(seconds int) {
	a.timeoutSeconds = seconds
}
func (a *CheckSSL) SetThreshold(threshold time.Time) {
	a.dateNeededValidFor = threshold
}

func (a *CheckSSL) CheckServer(target string, insecure bool) (output CheckedServer) {
	target = strings.Replace(target, "http://", "https://", 1)
	if !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	output.Target = target
	output.Passed = true

	dialerContext := &net.Dialer{
		Timeout: time.Duration(a.timeoutSeconds) * time.Second,
	}

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: insecure},
		ForceAttemptHTTP2: true,
		DialContext:       dialerContext.DialContext,
	}

	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			// TODO(sz) get a debug logger so people can see these state changes
			//fmt.Printf("  gotConn %+v", connInfo.Conn)
			ip, _, _ := net.SplitHostPort(connInfo.Conn.RemoteAddr().String())
			output.IpAddress = ip
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			//fmt.Printf("  dnsDone %+v", dnsInfo)
			if output.IpAddress == "" && len(dnsInfo.Addrs) > 0 {
				output.IpAddress = dnsInfo.Addrs[0].IP.String()
			}
			if dnsInfo.Err != nil {
				output.Err = dnsInfo.Err.Error()
				output.Passed = false
				output.ExitCode = RETURNCODE_ERROR
			}
		},
		//TLSHandshakeDone: func(state tls.ConnectionState, err error) {
		//	fmt.Println("  tlsHandshakeDone", state, err)
		//},
	}

	req, err := http.NewRequest("HEAD", target, nil)
	if err != nil {
		output.Err = err.Error()
		return
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	client := &http.Client{Transport: tr}
	response, err := client.Do(req)
	if err != nil {
		if !insecure {
			if !isTimeout(err) {
				//fmt.Println("  secure request failed, attempting insecure request")
				output = a.CheckServer(target, true)
			}
		}
		certError := errors.Unwrap(err)
		if certError != nil {
			output.Err = certError.Error()
		}
		output.Passed = false
		output.ExitCode = RETURNCODE_ERROR
		return
	}

	output.ServerInfo += response.Header.Get("Server")
	if output.ServerInfo != "" {
		output.ServerInfo += " - "
	}
	output.ServerInfo += response.Header.Get("x-powered-by")
	if output.HttpVersion == "" {
		output.HttpVersion = response.Proto
	}

	if response.TLS != nil {
		output.ServerName = response.TLS.ServerName
		output.TlsVersion = response.TLS.Version
		output.TlsAlgorithm = response.TLS.CipherSuite
		output.HttpVersion = response.TLS.NegotiatedProtocol

		for _, val := range response.TLS.PeerCertificates {
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

			newCode := checkIfExpirationIsWithinTolerance(a.dateNeededValidFor, val.NotBefore, val.NotAfter)
			if newCode > RETURNCODE_PASS {
				certInfo.IsInvalid = true
				output.ExitCode = newCode
				output.Passed = false
			}
			output.Certs = append(output.Certs, certInfo)
		}
	} else {
		output.Passed = false
		output.Err = "Missing TLS Connection"
	}

	return
}

func checkIfExpirationIsWithinTolerance(dateThreshold time.Time, notBefore time.Time, notAfter time.Time) int {
	if dateThreshold.After(notBefore) && dateThreshold.Before(notAfter) {
		return RETURNCODE_PASS
	}

	if time.Now().After(notBefore) && time.Now().Before(notAfter) {
		return RETURNCODE_THRESHOLDFAIL
	}

	if dateThreshold.Before(notBefore) && dateThreshold.Before(notAfter) {
		return RETURNCODE_NOTVALIDYET
	}

	return RETURNCODE_EXPIRED
}

func isTimeout(err error) bool {
	if strings.Contains(err.Error(), "i/o timeout") {
		return true
	}
	return false
}

func displayDate(input time.Time) string {
	//	Mon Jan 2 15:04:05 -0700 MST 2006
	return input.Format(dateLayout) + " (" + numberOfDays(input) + " days)"
}
func numberOfDays(input time.Time) string {
	if input.IsZero() {
		return ""
	}
	return fmt.Sprintf("%.1f", input.Sub(time.Now()).Hours()/24)
}
func durationDays(before time.Time, after time.Time) string {
	return fmt.Sprintf("%.1f", after.Sub(before).Hours()/24)
}

func (a CheckedServer) AsString(enableColors bool) (output string) {
	setTerminalColor(enableColors)

	if a.ServerName != "" && a.IpAddress != "" {
		output += fmt.Sprintf("\n%s => %s\n", a.ServerName, a.IpAddress)
	}

	if a.HttpVersion != "" && a.TlsAlgorithm > 0 {
		if a.ServerInfo == "" {
			output += " -> (no server name or versions found)\n"
		} else {
			output += fmt.Sprintf(" -> %s\n", expandServerNames(a.ServerInfo))
		}
		output += fmt.Sprintf(" -> %s with %s\n", getHttpVersion(a.HttpVersion), getTlsVersion(a.TlsVersion))
		output += fmt.Sprintf(" -> %s %s\n", getTlsAlgo(a.TlsAlgorithm), getMozillaRecommendedCipher(a.TlsAlgorithm))
	}

	for i, cert := range a.Certs {

		if cert.IsCertificateAuthority {
			output += fmt.Sprintf(" CA-%d) ", i+1)
		} else {
			output += fmt.Sprintf(" %d) ", i+1)
		}

		if cert.IsInvalid {
			output += fmt.Sprintf("%s%s expired on %s%s", terminalRed, cert.CommonName, displayDate(cert.ValidNotAfter), terminalNoColor)
		} else {
			output += fmt.Sprintf("%s expires on %s", cert.CommonName, displayDate(cert.ValidNotAfter))
		}
		output += "\n"
	}

	output += a.summaryLine()
	return
}

func (a CheckedServer) AsShortString(enableColors bool) (output string) {
	setTerminalColor(enableColors)
	output = a.summaryLine()
	return
}

func (a CheckedServer) summaryLine() string {
	if a.Passed {
		return fmt.Sprintf("%s[PASS]%s %s\n", terminalGreen, terminalNoColor, a.Target)
	}
	return fmt.Sprintf("%s[FAIL]%s %s\n", terminalRed, terminalNoColor, a.Target)
}

type OutputFormat int64

const (
	JSON OutputFormat = iota
	CSV
	TEXT
	NONE
	SHORT
)

func (a OutputFormat) String() string {
	switch a {
	case JSON:
		return "JSON"
	case CSV:
		return "CSV"
	case TEXT:
		return "TEXT"
	case NONE:
		return "NONE"
	case SHORT:
		return "SHORT"
	}
	return ""
}

func (a CheckedServer) AsJson() string {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "{ \"error\": \"Unable to convert result to json\"}"
	}
	return string(jsonBytes)
}

func CsvHeaderRow() string {
	return "Target,Result,Days to Expire,Duration,Common Name,CA Name,Error"
}
func (a CheckedServer) AsCsv() string {
	var leastDays time.Time
	duration := ""
	commonName := ""
	caName := ""
	for i, cert := range a.Certs {
		if leastDays.After(cert.ValidNotAfter) || leastDays.IsZero() {
			leastDays = cert.ValidNotAfter
		}
		if i == 0 {
			commonName = cert.CommonName
			duration = durationDays(cert.ValidNotBefore, cert.ValidNotAfter)
		}
		if cert.IsCertificateAuthority {
			caName = cert.CommonName
		}
	}
	return strings.Join([]string{a.Target, csvConvertResult(a.ExitCode), numberOfDays(leastDays), duration, commonName, caName, a.Err}, ",")
}
func csvConvertResult(input int) string {
	if input == 0 {
		return "PASS"
	} else if input == RETURNCODE_EXPIRED {
		return "EXPIRED"
	}
	return "FAIL"
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

	switch strings.ToLower(input) {
	case "h3":
		return "HTTP/3" // currently, go cannot do http3 natively
	case "h2":
		return terminalGreen + "HTTP/2" + terminalNoColor
	case "http/1.1":
		return terminalRed + "HTTP/1.1 (OLD)" + terminalNoColor
	case "http/1":
		return terminalRed + "HTTP/1 (OLD)" + terminalNoColor
	}
	return fmt.Sprintf("unknown HTTPS version %s", input)

}

func expandServerNames(input string) (output string) {

	output = strings.Replace(input, "ATS", "Apache Traffic Server (ATS)", 1)
	output = strings.Replace(output, "gws", "Google Web Server (gws)", 1)

	return
}

func getAllDnsRecordsFor(input string) []string {
	output := []string{}
	addresses, err := net.LookupIP(input)
	if err != nil {
		return output
	}

	for i, ip := range addresses {
		ipv4 := ip.To4()
		if ipv4 != nil {
			fmt.Println(i, ip)
			output = append(output, ipv4.String())
		}
	}
	return output
}
