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

func Test_Display(t *testing.T) {
	date := time.Now().Add(-48 * time.Hour)
	actual := displayDate(date)
	if strings.Contains(actual, " (-2.0 days)") == false {
		t.Log(actual)
		t.Error("Expecting result to contain -2.0 days")
	}
}

func assert(t *testing.T, actual string, expected string, failureHint string) {
	if actual != expected {
		t.Log("ACTUAL = ", actual)
		t.Log("EXPECT = ", expected)
		t.Fatal("[FAILED]", failureHint)
	}
}
