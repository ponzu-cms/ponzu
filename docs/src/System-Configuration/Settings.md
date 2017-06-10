title: Configuring Your Ponzu System Settings

Ponzu has several options which can be configured at run-time. To view these
configuration settings, visit the `/admin/configure` page of your Ponzu CMS.

---

#### Site Name
The Site Name setting changes the displayed name on your admin dashboard. This is
visible publicly on the `/admin/login` page.

---

#### Domain Name
Internally, Ponzu needs to know where its canonical HTTP access origin is, and
requires you to add the qualified domain name you are using. In development, use 
`localhost` or some other name mapped to the loopback address (`127.0.0.1`).

Once you have deployed your Ponzu server to a remote host and pointed a public 
domain at it, you need to change the Domain Name setting to match. This is 
especially important when fetching TLS (SSL) certificates from [Let's Encrypt](https://letsencrypt.org)
- since the process requires an active, verifiable domain. To set up your server
with TLS over HTTPS connections, follow these steps:

1. Set your Domain Name in the system configuration
2. Set the Administrator Email to register with Let's Encrypt
2. Stop your Ponzu server
3. Run your Ponzu server with the `--https` flag e.g. `$ ponzu run --https`
4. Visit your CMS admin with `https://` prepended to your URL

!!! success "Verifying HTTPS / TLS Connections"
    If successful, your APIs and CMS will be accessible via HTTPS, and you will
    see a green indicator near the URL bar of most browsers. This also enables 
    your server to use the HTTP/2 protocol.

##### Development Environment

You can test HTTPS & HTTP/2 connections in your development environment on `localhost`,
by running Ponzu with the `--devhttps` flag e.g. `$ ponzu --devhttps run` 

If you're greeted with a warning from the browser saying the connection is not
secure, follow the steps outlined in the CLI message, or here:
```
If your browser rejects HTTPS requests, try allowing insecure connections on localhost.
on Chrome, visit chrome://flags/#allow-insecure-localhost
```

---

#### Administrator Email
The Administrator Email is the contact email for the person who is the main admin
of your Ponzu CMS. This can be changed at any point, but once a Let's Encrypt
certificate has been fetched using an Administrator Email, it will remain the 
contact until a new certificate is requested. 

---

#### Client Secret
The Client Secret is a secure value used by the server to sign tokens and authenticate requests.
**Do not share this** value with any untrusted party.

!!! danger "Security and the Client Secret"
    HTTP requests with a valid token, signed with the Client Secret, can take any
    action an Admin can within the CMS. Be cautious of this when sharing account
    logins or details with anyone.

---

#### Etag Header
The Etag Header value is automatically created when content is changed and serves
as a caching validation mechanism.

---

#### CORS
CORS, or "Cross-Origin Resource Sharing" is a security setting which defines how
resources (or URLs) can be accessed from outside clients / domains. By default, 
Ponzu HTTP APIs can be accessed from any origin, meaning a script from an unknown
website could fetch data. 

By disabling CORS, you limit API requests to only the Domain Name you set.

---

#### GZIP
GZIP is a popular codec which when applied to most HTTP responses, decreases data
transmission size and response times. The GZIP setting on Ponzu has a minor 
side-effect of using more CPU, so you can disable it if you notice your system 
is CPU-constrained. However, traffic levels would need to be extremely demanding
for this to be noticeable.

---

#### HTTP Cache
The HTTP Cache configuration allows a system to disable the default HTTP cache,
which saves the server from repeating API queries and sending responses -- it's
generally advised to keep this enabled unless you have _frequently_ changing data.

The `Max-Age` value setting overrides the default 2592000-second (30 day) cache
`max-age` duration set in API response headers. The `0` value is an alias to 
`2592000`, so check the `Disable HTTP Cache` box if you don't want any caching.


---

#### Invalidate Cache
If this box is checked and then the configuration is saved, the server will 
re-generate an Etag to send in responses. By doing so, the cache becomes invalidated
and reset so new content or assets will be included in previously cached responses.

The cache is invalidated when content changes, so this is typically not a widely 
used setting.

---

#### Database Backup Credentials
In order to enable HTTP backups of the components that make up your system, you
will need to add an HTTP Basic Auth user and password pair. When used to 
[run backups](/Running-Backups/Backups), the `user:password` pair tells your server
that the backup request is made from a trusted party. 

!!! danger "Backup Access with Credentials"
    This `user:password` pair should not be shared outside of your organization as 
    it allows full database downloads and archives of your system's uploads.
