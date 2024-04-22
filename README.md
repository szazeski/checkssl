# checkssl
command line tool to check if a webserver has a valid https certificate.

[![Go](https://github.com/szazeski/checkssl/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/szazeski/checkssl/actions/workflows/go.yml)

https://www.checkssl.org/

<a rel="me" href="https://fosstodon.org/@checkssl">Follow us on Mastodon</a>

## Example

`checkssl steve.zazeski.com`
```
steve.zazeski.com
 -> openresty - 
 -> HTTP/2 with TLS v1.3 (released 2018) - latest version
 -> TLS_AES_128_GCM_SHA256 = TLS, message encrypted with AES128 GCM, hashes are SHA256 
 1) steve.zazeski.com expires on 2022-02-13 12:32AM Sun (44.8 days)
 CA-2) R3 expires on 2025-09-15 4:00PM Mon (1355.5 days)
 CA-3) ISRG Root X1 expires on 2035-06-04 11:04AM Mon (4904.3 days)
[PASS] https://steve.zazeski.com

```

If the certificate is not valid, a non-zero exit code will be returned to stop a ci build. 
```
expired.badssl.com
 -> nginx/1.10.3 (Ubuntu) - 
 -> HTTP/1.1 (OLD) with TLS v1.2 (released 2008) - Consider upgrading to TLS v1.3
 -> TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 
 1) *.badssl.com expired on 2015-04-12 11:59PM Sun (-2621.0 days)
 CA-2) COMODO RSA Domain Validation Secure Server CA expires on 2029-02-11 11:59PM Sun (2433.0 days)
 CA-3) COMODO RSA Certification Authority expired on 2020-05-30 10:48AM Sat (-746.6 days)
https://expired.badssl.com x509: certificate has expired or is not yet valid: current time 2022-06-15T19:48:19-05:00 is after 2015-04-12T23:59:59Z
[FAIL] https://expired.badssl.com
```

```
ebay.com
 stopped after 10 redirects
[FAIL] https://ebay.com
```

### Parameters
(You can use - or -- for all parameters)

`-days=60` allows you to specify a threshold of when checkssl should error to allow CI jobs to fail if the certs are about to expire in a few days.

`-json` will switch the text output from human-readable to JSON format for easier parsing with other systems.

`-csv` will switch the text output to comma seperated values that are easier to copy and paste into a spreadsheet.

`-no-color` will not add terminal color syntax to output, helpful for CI systems that do not have color enabled.

`-no-output` will not show any text but still return a status code. Helpful for CI that just want to know if the certs are valid.

`-short` will reduce each target's output to just the pass/fail line with the url/dns.


### Return Codes

`0` All certificates passed

`2` Certificate(s) are expired

`3` Certificate(s) are valid now but user specified threshold failed (from -days flag)

`4` Certificate(s) are not yet valid

`5` General error, normally due to network failure

## Installation

### Linux/Mac
```
wget https://github.com/szazeski/checkssl/releases/download/v0.5.1/checkssl_0.5.1_$(uname -s)_$(uname -m).tar.gz -O checkssl.tar.gz && tar -xf checkssl.tar.gz && chmod +x checkssl && sudo mv checkssl /usr/bin/
```

### Docker
[Dockerhub](https://hub.docker.com/r/szazeski/checkssl) `docker run --rm szazeski/checkssl hub.docker.com`

### Mac

[macports](https://ports.macports.org/port/checkssl/) `sudo port install checkssl`

[homebrew](https://brew.sh/) `brew install szazeski/tap/checkssl`


### Windows (Powershell)

```
Invoke-WebRequest https://github.com/szazeski/checkssl/releases/download/v0.5.1/checkssl_0.5.1_Windows_x86_64.tar.gz -outfile checkssl.tar.gz; tar -xzf checkssl.tar.gz; echo "if you want, move the file to a PATH directory like WINDOWS folder"
```

then move to `C:\Windows\` or other PATH directory
