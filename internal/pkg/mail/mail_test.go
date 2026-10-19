package mail

import (
	"bufio"
	"net"
	"strings"
	"testing"

	systemModel "octoops/internal/model/system"
)

func TestEnvelopeRecipients(t *testing.T) {
	tests := []struct {
		name    string
		opt     MailOptions
		want    []string
		wantErr bool
	}{
		{
			name: "single",
			opt:  MailOptions{To: "a@example.com"},
			want: []string{"a@example.com"},
		},
		{
			name: "list and cc",
			opt: MailOptions{
				To: "A <a@example.com>, b@example.com",
				Cc: []string{"c@example.com"},
			},
			want: []string{"a@example.com", "b@example.com", "c@example.com"},
		},
		{name: "empty", opt: MailOptions{}, wantErr: true},
		{name: "invalid", opt: MailOptions{To: "not-an-address"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := envelopeRecipients(tt.opt)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendNoneDoesNotStartTLS(t *testing.T) {
	server := newFakeSMTPServer(t, true)
	if err := sendWithConfig(server.config(systemModel.MailSecurityNone), MailOptions{
		To:      "to@example.com",
		Subject: "test",
		Body:    "<p>hi</p>",
	}); err != nil {
		t.Fatalf("send failed: %v", err)
	}
	for _, command := range server.commandsSnapshot() {
		if strings.HasPrefix(strings.ToUpper(command), "STARTTLS") {
			t.Fatalf("none mode must not send STARTTLS, commands: %v", server.commandsSnapshot())
		}
	}
	select {
	case message := <-server.messages:
		if !strings.Contains(message, "Subject: test") {
			t.Fatalf("unexpected message: %s", message)
		}
	default:
		t.Fatal("server did not receive message")
	}
}

func TestSendTLSRequiresSTARTTLS(t *testing.T) {
	server := newFakeSMTPServer(t, false)
	err := sendWithConfig(server.config(systemModel.MailSecurityTLS), MailOptions{
		To:      "to@example.com",
		Subject: "test",
		Body:    "<p>hi</p>",
	})
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("expected STARTTLS error, got %v", err)
	}
	for _, command := range server.commandsSnapshot() {
		if strings.HasPrefix(strings.ToUpper(command), "MAIL FROM") {
			t.Fatalf("must not send MAIL FROM without STARTTLS, commands: %v", server.commandsSnapshot())
		}
	}
}

type fakeSMTPServer struct {
	listener          net.Listener
	advertiseStartTLS bool
	commands          chan string
	messages          chan string
}

func newFakeSMTPServer(t *testing.T, advertiseStartTLS bool) *fakeSMTPServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &fakeSMTPServer{
		listener:          listener,
		advertiseStartTLS: advertiseStartTLS,
		commands:          make(chan string, 128),
		messages:          make(chan string, 1),
	}
	t.Cleanup(func() { _ = listener.Close() })
	go server.serve()
	return server
}

func (s *fakeSMTPServer) config(security string) systemModel.MailSetting {
	return systemModel.MailSetting{
		Enable:      true,
		SMTPAddress: "127.0.0.1",
		SMTPPort:    s.listener.Addr().(*net.TCPAddr).Port,
		SMTPUser:    "from@example.com",
		Security:    security,
	}
}

func (s *fakeSMTPServer) commandsSnapshot() []string {
	commands := make([]string, 0, len(s.commands))
	for {
		select {
		case command := <-s.commands:
			commands = append(commands, command)
		default:
			return commands
		}
	}
}

func (s *fakeSMTPServer) serve() {
	conn, err := s.listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	write := func(line string) {
		_, _ = writer.WriteString(line + "\r\n")
		_ = writer.Flush()
	}
	write("220 localhost ESMTP")

	inData := false
	var data strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		s.commands <- line

		if inData {
			if line == "." {
				inData = false
				s.messages <- data.String()
				write("250 OK")
				continue
			}
			data.WriteString(line + "\n")
			continue
		}

		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"):
			write("250-localhost")
			if s.advertiseStartTLS {
				write("250-STARTTLS")
			}
			write("250 OK")
		case strings.HasPrefix(upper, "HELO"):
			write("250 localhost")
		case strings.HasPrefix(upper, "MAIL FROM"):
			write("250 OK")
		case strings.HasPrefix(upper, "RCPT TO"):
			write("250 OK")
		case upper == "DATA":
			write("354 End data with <CR><LF>.<CR><LF>")
			inData = true
		case upper == "QUIT":
			write("221 Bye")
			return
		case upper == "STARTTLS":
			write("454 TLS not available")
		default:
			write("250 OK")
		}
	}
}
