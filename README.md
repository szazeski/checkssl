# checkssl
command line tool to check if a webserver has a valid https certificate.

https://www.checkssl.org/

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
zazeski.com
 certificate signed by unknown authority
[FAIL] zazeski.com
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

`-no-color` will not add terminal color syntax to output, helpful for CI systems that do not have color enabled.

### Return Codes
`0` All certificates passed

`2` Certificate(s) expired

`3` User specified threshold failed 

## Installation

On 64-bit linux systems, you can run
`wget https://github.com/szazeski/checkssl/releases/download/v0.2/checkssl-linux-amd64 && chmod +x checkssl-linux-amd64 && sudo mv checkssl-linux-amd64 /usr/bin/checkssl`
(will ask for a sudo password to move it into the system-wide bin folder, switch it to a local path if you don't want to do that)

On mac, open terminal
`curl -O -L https://github.com/szazeski/checkssl/releases/download/v0.2/checkssl-macos && chmod +x checkssl-macos`
Since the app is not signed, you should open it here by right clicking on it and clicking open to tell Gatekeeper that you approve running it.
`mv checkssl-macos /usr/local/bin/checkssl`

On windows (powershell)
`wget https://github.com/szazeski/checkssl/releases/download/v0.2/checkssl-windows-amd64.exe -outfile checkssl.exe`
then move to C:\Windows\checkssl.exe
