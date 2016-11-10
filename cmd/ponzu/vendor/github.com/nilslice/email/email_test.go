package email

import (
	"testing"
)

func TestSend(t *testing.T) {
	m := Message{
		To:      "",
		From:    "",
		Subject: "",
		Body:    "",
	}

	err := m.Send()
	if err != nil {
		t.Fatal("Send returned error:", err)
	}
}
