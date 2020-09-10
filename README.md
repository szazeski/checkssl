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

If the certificate is not valid, a non zero return code will be sent. 
```
zazeski.com
 Head "https://zazeski.com": x509: certificate signed by unknown authority
[FAIL]
```