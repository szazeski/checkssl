package checkssl

import (
	"strings"
	"testing"
	"time"
)

func Test_CheckServer_Blank(t *testing.T) {
	actual := CheckServer("", time.Now(), false)

	if actual.Passed != false {
		t.Fatal("no target should of produced a failing response")
	}
	assert(t, actual.Err, "http: no Host in request URL", "")
}

func Test_CheckServer_checksslorg(t *testing.T) {
	actual := CheckServer("checkssl.org", time.Now(), false)

	if !actual.Passed {
		t.Fatal("expecting to get a passing reply")
	}
	assert(t, actual.Err, "", "")
}

func Test_CheckServer_expired(t *testing.T) {
	testFailure(t, "https://expired.badssl.com/")
}
func Test_CheckServer_wrongHost(t *testing.T) {
	testFailure(t, "https://wrong.host.badssl.com/")
}
func Test_CheckServer_selfSigned(t *testing.T) {
	testFailure(t, "https://self-signed.badssl.com/")
}
func Test_CheckServer_untrustedRoot(t *testing.T) {
	testFailure(t, "https://untrusted-root.badssl.com/")
}
func Test_CheckServer_revoked(t *testing.T) {
	t.Skipf("This should fail eventually")
	testFailure(t, "https://revoked.badssl.com/")
}
func Test_CheckServer_pinningTest(t *testing.T) {
	t.Skipf("This should fail eventually")
	testFailure(t, "https://pinning-test.badssl.com/")
}
func testFailure(t *testing.T, target string) {
	actual := CheckServer(target, time.Now(), false)

	if actual.Passed {
		t.Fatal("expecting to get a failure reply")
	}
}

func Test_checkIfExpirationIsWithinTolerance_Before(t *testing.T) {
	today := time.Now()
	lastWeek := today.Add(-7 * 24 * time.Hour)
	nextWeek := today.Add(7 * 24 * time.Hour)

	//      t--X
	//   |---------|    PASS - 0
	//  |--|            FAIL - Expired - 2
	//   |----|         FAIL - Threshold Fail - 3
	//           |----| FAIL - Not Valid Yet - 4

	// 0
	actual := checkIfExpirationIsWithinTolerance(today, lastWeek, nextWeek)
	if actual != RETURNCODE_PASS {
		t.Log("checkIfExpirationIsWithinTolerance", actual, "but expected", RETURNCODE_PASS)
		t.Error("Expected to get a passing")
	}

	// 2
	actual = checkIfExpirationIsWithinTolerance(today.Add(8*24*time.Hour), lastWeek, nextWeek)
	if actual != RETURNCODE_THRESHOLDFAIL {
		t.Log("checkIfExpirationIsWithinTolerance", actual, "but expected", RETURNCODE_THRESHOLDFAIL)
		t.Error("Expected to get a THRESHOLDFAIL")
	}

	// 3
	actual = checkIfExpirationIsWithinTolerance(today, nextWeek, nextWeek)
	if actual != RETURNCODE_NOTVALIDYET {
		t.Log("checkIfExpirationIsWithinTolerance", actual, "but expected", RETURNCODE_NOTVALIDYET)
		t.Error("Expected to get a NOTVALIDYET")
	}

	// 4
	actual = checkIfExpirationIsWithinTolerance(today, lastWeek, lastWeek)
	if actual != RETURNCODE_EXPIRED {
		t.Log("checkIfExpirationIsWithinTolerance", actual, "but expected", RETURNCODE_EXPIRED)
		t.Error("Expected to get a RETURNCODE_EXPIRED")
	}
}

func Test_DisplayDate(t *testing.T) {
	date := time.Now().Add(-48 * time.Hour)
	actual := displayDate(date)
	assert(t, actual, date.Format(dateLayout)+" (-2.0 days)", "Expecting result to contain -2.0 days")
}

func Test_DisplayDate_Zero(t *testing.T) {
	date := time.Time{}
	actual := displayDate(date)
	assert(t, actual, date.Format(dateLayout)+" ( days)", "Expecting zero time to not calculate offset")
}

func Test_AsString_NoColors(t *testing.T) {
	results := generateRealisticResult()
	actual := results.AsString(false)
	expected := "\n" +
		"www.checkssl.org => 2600:9000:24d0:fa00:1e:e294:3240:93a1\n" +
		" -> AmazonS3 - \n" +
		" -> HTTP/2 with TLS v1.3 (released 2018) - latest version\n" +
		" -> TLS_AES_128_GCM_SHA256 = TLS, message encrypted with AES128 GCM, hashes are SHA256 \n" +
		" 1) *.checkssl.org expires on " + displayDate(results.Certs[0].ValidNotAfter) + "\n" +
		" CA-2) Amazon RSA 2048 M01 expires on " + displayDate(results.Certs[1].ValidNotAfter) + "\n" +
		" CA-3) Amazon Root CA 1 expires on " + displayDate(results.Certs[2].ValidNotAfter) + "\n" +
		" CA-4) Starfield Services Root Certificate Authority - G2 expires on " + displayDate(results.Certs[3].ValidNotAfter) + "\n" +
		"[PASS] https://checkssl.org\n"
	assert(t, actual, expected, "")
}

func Test_AsCsv_Pass(t *testing.T) {
	results := generateRealisticResult()
	actual := results.AsCsv()
	expected := "https://checkssl.org,PASS,5.0,10.0,*.checkssl.org,Starfield Services Root Certificate Authority - G2,"
	assert(t, actual, expected, "")
}
func Test_AsCsv_Fail(t *testing.T) {
	results := CheckedServer{Target: "example.com", ExitCode: 2, Err: "dial tcp: lookup example.com: no such host"}
	actual := results.AsCsv()
	expected := "example.com,EXPIRED,,,,,dial tcp: lookup example.com: no such host"
	assert(t, actual, expected, "")
}

func Test_CsvHeaderRowsMatchOutputRows(t *testing.T) {
	headerRow := CsvHeaderRow()
	splitHeader := strings.Split(headerRow, ",")
	sizeOfHeader := len(splitHeader)
	resultRow := CheckedServer{}.AsCsv()
	splitResult := strings.Split(resultRow, ",")
	sizeOfResult := len(splitResult)

	if sizeOfResult != sizeOfHeader {
		t.Log("Header Row -", sizeOfHeader, "-", headerRow)
		t.Log("Result Row -", sizeOfResult, "-", resultRow)
		t.Error("CSV header rows does not match size of result row")
	}
}

func generateRealisticResult() CheckedServer {
	local := time.Local
	return CheckedServer{
		Target:     "https://checkssl.org",
		Err:        "",
		ExitCode:   0,
		ServerInfo: "AmazonS3 - ",
		Certs: []CheckCert{
			{
				CommonName:             "*.checkssl.org",
				IsCertificateAuthority: false,
				ValidNotBefore:         time.Now().Add(-5 * 24 * time.Hour),
				ValidNotAfter:          time.Now().Add(5 * 24 * time.Hour),
				IsInvalid:              false,
			}, {
				CommonName:             "Amazon RSA 2048 M01",
				IsCertificateAuthority: true,
				ValidNotBefore:         time.Date(2022, 8, 23, 22, 21, 28, 0, local),
				ValidNotAfter:          time.Now().Add(10 * 24 * time.Hour),
				IsInvalid:              false,
			}, {
				CommonName:             "Amazon Root CA 1",
				IsCertificateAuthority: true,
				ValidNotBefore:         time.Date(2015, 5, 25, 12, 0, 0, 0, local),
				ValidNotAfter:          time.Now().Add(50 * 24 * time.Hour),
				IsInvalid:              false,
			}, {
				CommonName:             "Starfield Services Root Certificate Authority - G2",
				IsCertificateAuthority: true,
				ValidNotBefore:         time.Date(2009, 9, 2, 0, 0, 0, 0, local),
				ValidNotAfter:          time.Now().Add(100 * 24 * time.Hour),
				IsInvalid:              false,
			}},
		Passed:       true,
		HttpVersion:  "h2",
		TlsVersion:   772,
		TlsAlgorithm: 4865,
		ServerName:   "www.checkssl.org",
		IpAddress:    "2600:9000:24d0:fa00:1e:e294:3240:93a1",
	}
}

// = = = = = = =

func assert(t *testing.T, actual string, expected string, failureHint string) {
	if actual != expected {
		t.Log("ACTUAL = ", actual, len(actual), "chars")
		t.Log("EXPECT = ", expected, len(expected), "chars")
		t.Fatal("[FAILED]", failureHint)
	}
}
