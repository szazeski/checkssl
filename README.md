# checkssl

command line tool to check if a webserver has its https certificates correctly setup.

## Example

`checkssl steve.zazeski.com`
```
steve.zazeski.com
openresty - PHP/7.4.4
 1) steve.zazeski.com expires on 2020-11-07 2:58PM Sat (58.6 days)
 2) -CA- Let's Encrypt Authority X3 expires on 2021-03-17 4:40PM Wed (188.6 days)
 3) -CA- DST Root CA X3 expires on 2021-09-30 2:01PM Thu (385.5 days)
[PASS] ★★★
```

If the certificate is not valid, a non-zero exit code will be returned to stop a ci build. 
```
zazeski.com
 Head "https://zazeski.com": x509: certificate signed by unknown authority
[FAIL]
```

### Parameters
`-days=60` allows you to specify a threshold of when checkssl should error to allow CI jobs to fail if the certs are about to expire in a few days.

### Return Codes
`0` All certificates passed

`2` Certificate(s) expired

`3` User specified threshold failed 

## Installation

On 64-bit linux systems, you can run
`wget https://github.com/szazeski/checkssl/releases/download/v0.1/checkssl-linux-amd64 && chmod +x checkssl-linux-amd64 && sudo mv checkssl-linux-amd64 /usr/bin/checkssl`
(will ask for a sudo password to move it into the system-wide bin folder, switch it to a local path if you don't want to do that)

On mac, open terminal
`curl -O -L https://github.com/szazeski/checkssl/releases/download/v0.1/checkssl-macos && chmod +x checkssl-macos`
Since the app is not signed, you should open it here by right clicking on it and clicking open to tell Gatekeeper that you approve running it.
`mv checkssl-macos /usr/local/bin/checkssl`

On windows
