package email

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// Message creates a email to be sent
type Message struct {
	To      string
	From    string
	Subject string
	Body    string
}

var (
	ports = []int{25, 2525, 587}
)

// Send sends a message to recipient(s) listed in the 'To' field of a Message
func (m Message) Send() error {
	if !strings.Contains(m.To, "@") {
		return fmt.Errorf("Invalid recipient address: <%s>", m.To)
	}

	host := strings.Split(m.To, "@")[1]
	addrs, err := net.LookupMX(host)
	if err != nil {
		return err
	}

	c, err := newClient(addrs, ports)
	if err != nil {
		return err
	}

	err = send(m, c)
	if err != nil {
		return err
	}

	return nil
}

func newClient(mx []*net.MX, ports []int) (*smtp.Client, error) {
	for i := range mx {
		for j := range ports {
			server := strings.TrimSuffix(mx[i].Host, ".")
			hostPort := fmt.Sprintf("%s:%d", server, ports[j])
			client, err := smtp.Dial(hostPort)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			return client, nil
		}
	}

	return nil, fmt.Errorf("Couldn't connect to servers %v on any common port.", mx)
}

func send(m Message, c *smtp.Client) error {
	if err := c.Mail(m.From); err != nil {
		return err
	}

	if err := c.Rcpt(m.To); err != nil {
		return err
	}

	msg, err := c.Data()
	if err != nil {
		return err
	}

	if m.Subject != "" {
		_, err = msg.Write([]byte("Subject: " + m.Subject + "\r\n"))
		if err != nil {
			return err
		}
	}

	if m.From != "" {
		_, err = msg.Write([]byte("From: <" + m.From + ">\r\n"))
		if err != nil {
			return err
		}
	}

	if m.To != "" {
		_, err = msg.Write([]byte("To: <" + m.To + ">\r\n"))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprint(msg, m.Body)
	if err != nil {
		return err
	}

	err = msg.Close()
	if err != nil {
		return err
	}

	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}
