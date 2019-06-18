# pvga

pvga (Pretty Vulnerable Go Application) is a webapp with intentional vulnerabilities. Please don't use this for anything but educational purposes!

pvga demonstrates all ten of the OWASP Top Ten security risks. It does so via an incremental clicker game ("Click On This Cat") containing intentionally insecure code, and depends on a [package containing vulnerabilities](https://github.com/empayne/redundantserializer).

## Running pvga
* Clone this repository.
* Run `dep ensure && docker-compose up` in the project's root folder
* Access `localhost:8080` via your web browser. Login details can be found in `schema.sql`.

Please don't run pvga anywhere but on your local machine in a trusted network. This project contains many vulnerabilities. 

## Reading the pvga source code
Search for "OWASP" in this repository's source code (and in [redundantserializer's source code]((https://github.com/empayne/redundantserializer))) to find the intentionally vulnerable code. Explanatory comments accompany each of these vulnerabilities. A quick summary is as follows:
1. **Injection:** Our SQL query to update a user's bio is vulnerable to SQL injection.
2. **Broken Authentication:** There is no timeout on a failed login, so our login page can be bruteforced.
3. **Sensitive Data Exposure:** We store plaintext passwords in our database, which can be read via #1's SQL injection.
4. **XML External Entities (XXE):** redundantserializer v1.0.0 parses untrusted XML with external entity processing enabled.
5. **Broken Access Control:** Any user can reset any other user's score by modifying the user ID in a POST request.
6. **Security Misconfiguration:** Some of our error messages send out stack traces to the client, as the DEBUG environment variable is enabled.
7. **Cross-Site Scripting (XSS):** A query string parameter used to render error messages can be used to execute arbitrary Javascript on the login page.
8. **Insecure Deserialization:** The base64 data from redundantserializer can be altered, allowing a attackers to give themselves arbitrarily high scores. 
9. **Using Components with Known Vulnerabilities:** The XXE flaw in redundantserializer has been patched in v1.0.1, but pvga is pinned to use v1.0.0.
10. **Insufficient Logging & Monitoring:** There's no logging in pvga beyond the default logging provided by Gin.

## // TODO
Video demonstration of pvga.
