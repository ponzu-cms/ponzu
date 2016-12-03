## Email

I needed a way to send email from a [Ponzu](https://ponzu-cms.org) installation 
running on all kinds of systems without shelling out. `sendmail` or `postfix` et 
al are not standard on all systems, and I didn't want to force users to add API 
keys from a third-party just to send something like an account recovery email. 

### Usage:
`$ go get github.com/nilslice/email`

```go
package main

import (
    "fmt"
    "github.com/nilslice/email"
)

func main() {
    msg := email.Message{
        To: "you@server.name", // do not add < > or name in quotes
        From: "me@server.name", // do not add < > or name in quotes
        Subject: "A simple email",
        Body: "Plain text email body. HTML not yet supported, but send a PR!",
    }

    err := msg.Send()
    if err != nil {
        fmt.Println(err)
    }
}

```

### Under the hood
`email` looks at a `Message`'s `To` field, splits the string on the @ symbol and 
issues an MX lookup to find the mail exchange server(s). Then it iterates over 
all the possibilities in combination with commonly used SMTP ports for non-SSL
clients: `25, 2525, & 587`

It stops once it has an active client connected to a mail server and sends the
initial information, the message, and then closes the connection.

Currently, this doesn't support any additional headers or `To` field formatting
(the recipient's email must be the only string `To` takes). Although these would
be fairly strightforward to implement, I don't need them yet.. so feel free to 
contribute anything you find useful.

#### Warning
Be cautious of how often you run this locally or in testing, as it's quite 
likely your IP will be blocked/blacklisted if it is not already. 