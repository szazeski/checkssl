#!/bin/bash

TEST_PASS=0
TEST_FAIL=0
function failtest() {
    TEST_FAIL=$((TEST_FAIL+1))
    echo "[FAIL] $1"
}
function passtest() {
    TEST_PASS=$((TEST_PASS+1))
    echo "[PASS] $1"
}

echo "cli-test-suite.sh"
echo "  This suite tests functionality of the locally built ./checkssl"

if [ ! -f ./checkssl ]; then
    echo "  ./checkssl does not exist, attempting to build one"
    go build
    if [ ! -f ./checkssl ]; then
        echo "  ./checkssl still does not exist, [FAIL] test"
        exit 1
    fi
fi

OUTPUT=$(./checkssl)
EXPECTED=$(cat <<-END
checkssl [url] [url] [url] ...
 easy to read/parse information about ssl certificates
 version 0.6.0 built 2024-Aug-5
  -days=5 (will fail the check if the cert is within 5 days of renewal)
  -json (will output in JSON format)
  -csv (will output in comma seperated format for spreadsheets)
  -no-color (will disable color syntax from output)
  -no-output (will only produce exit code)
  -no-header (will disable the header row in CSV output)
  -short (will show only 1 line per result)
  -timeout=5 (will set the timeout to 5 seconds)  default = 15
END
)
diff <(echo "$OUTPUT") <(echo "$EXPECTED") && passtest "blank input matches" || failtest "blank input does not match"


# TODO version check


./checkssl checkssl.org -no-output
if [ $? -eq 0 ]; then
    passtest "checkssl.org is valid today"
else
    failtest "expected checkssl.org today to be valid"
fi


./checkssl checkssl.org -no-output -days=500
if [ $? -eq 0 ]; then
    failtest "checkssl.org in 500 days should be expired"
else
    passtest "checkssl.org shows expired in 500 days"
fi

# 192.168.200.200 is a private ip that shouldn't exist
./checkssl 192.168.200.200 -no-output -timeout=1
if [ $? -eq 0 ]; then
    failtest "timeout url passed but it should not have"
else
    passtest "timeout url does not pass"
fi

./checkssl expired.badssl.com -no-output
if [ $? -eq 0 ]; then
    failtest "expired url passed but it should not have"
else
    passtest "expired url does not pass"
fi




if [ $TEST_FAIL -eq 0 ]; then
    echo "ALL tests passed"
    exit 0
else
    echo "$TEST_FAIL tests failed"
    exit 1
fi
