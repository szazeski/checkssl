package checkssl

import (
	"strings"
	"testing"
	"time"
)

func Test_CheckServer_Blank(t *testing.T) {
	actual := CheckServer("", time.Now())

	if actual.Passed != false {
		t.Fatal("no target should of produced a failing response")
	}
	assert(t, actual.Err, "http: no Host in request URL", "")
}

func Test_CheckServer_checksslorg(t *testing.T) {
	actual := CheckServer("checkssl.org", time.Now())

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
	actual := CheckServer(target, time.Now())

	if actual.Passed {
		t.Fatal("expecting to get a failure reply")
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
