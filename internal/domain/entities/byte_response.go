package entities

import "io"

type ResponseStream struct {
	ContentType        string
	ContentDisposition string
	Payload            io.Reader
}
