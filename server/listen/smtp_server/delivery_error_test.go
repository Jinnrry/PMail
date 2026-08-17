package smtp_server

import (
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"testing"

	smtp "github.com/emersion/go-smtp"
)

func TestDownstreamDeliverySMTPError(t *testing.T) {
	tests := []struct {
		name         string
		deliveryErr  error
		domainErrors map[string]error
		wantCode     int
		wantEnhanced smtp.EnhancedCode
		wantMessages []string
	}{
		{name: "success"},
		{
			name:         "network failure is temporary",
			deliveryErr:  errors.New("delivery failed"),
			domainErrors: map[string]error{"example.com": &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}},
			wantCode:     451,
			wantEnhanced: smtp.EnhancedCode{4, 4, 0},
			wantMessages: []string{"connection refused"},
		},
		{
			name:         "missing domain details is temporary",
			deliveryErr:  errors.New("delivery failed"),
			wantCode:     451,
			wantEnhanced: smtp.EnhancedCode{4, 4, 0},
			wantMessages: []string{"delivery failed"},
		},
		{
			name:        "downstream 4xx is temporary",
			deliveryErr: errors.New("delivery failed"),
			domainErrors: map[string]error{
				"example.com": &textproto.Error{Code: 421, Msg: "try again later"},
			},
			wantCode:     451,
			wantEnhanced: smtp.EnhancedCode{4, 4, 0},
			wantMessages: []string{"421", "try again later"},
		},
		{
			name:        "downstream 5xx is permanent",
			deliveryErr: errors.New("delivery failed"),
			domainErrors: map[string]error{
				"example.com": fmt.Errorf("remote rejected recipient: %w", &textproto.Error{Code: 550, Msg: "mailbox unavailable"}),
			},
			wantCode:     550,
			wantEnhanced: smtp.EnhancedCode{5, 0, 0},
			wantMessages: []string{"550", "mailbox unavailable"},
		},
		{
			name:        "nxdomain is permanent",
			deliveryErr: errors.New("delivery failed"),
			domainErrors: map[string]error{
				"missing.example": &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "no such host", Name: "missing.example", IsNotFound: true}},
			},
			wantCode:     550,
			wantEnhanced: smtp.EnhancedCode{5, 0, 0},
			wantMessages: []string{"no such host"},
		},
		{
			name:        "mixed permanent and temporary failures are temporary",
			deliveryErr: errors.New("delivery failed"),
			domainErrors: map[string]error{
				"rejected.example": &textproto.Error{Code: 550, Msg: "mailbox unavailable"},
				"slow.example":     &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "timeout", Name: "slow.example", IsTimeout: true}},
			},
			wantCode:     451,
			wantEnhanced: smtp.EnhancedCode{4, 4, 0},
			wantMessages: []string{"550", "mailbox unavailable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downstreamDeliverySMTPError(tt.deliveryErr, tt.domainErrors)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("downstreamDeliverySMTPError() = %v, want nil", err)
				}
				return
			}

			var smtpErr *smtp.SMTPError
			if !errors.As(err, &smtpErr) {
				t.Fatalf("downstreamDeliverySMTPError() type = %T, want *smtp.SMTPError", err)
			}
			if smtpErr.Code != tt.wantCode {
				t.Errorf("SMTP code = %d, want %d", smtpErr.Code, tt.wantCode)
			}
			if smtpErr.EnhancedCode != tt.wantEnhanced {
				t.Errorf("enhanced code = %v, want %v", smtpErr.EnhancedCode, tt.wantEnhanced)
			}
			for _, part := range tt.wantMessages {
				if !strings.Contains(smtpErr.Message, part) {
					t.Errorf("SMTP message = %q, want it to contain %q", smtpErr.Message, part)
				}
			}
			if strings.ContainsAny(smtpErr.Message, "\r\n") {
				t.Errorf("SMTP message must be one line, got %q", smtpErr.Message)
			}
			if len(smtpErr.Message) > maxDeliverySMTPMessageBytes {
				t.Errorf("SMTP message is too long: %d bytes", len(smtpErr.Message))
			}
		})
	}
}

func TestDeliverySMTPMessageSanitizesAndBoundsDetails(t *testing.T) {
	message := deliverySMTPMessage(
		"Downstream delivery temporarily failed",
		errors.New("fallback"),
		map[string]error{
			"b.example": errors.New("line one\r\nline two"),
			"a.example": errors.New(strings.Repeat("x", 200)),
		},
	)

	if strings.ContainsAny(message, "\r\n") {
		t.Fatalf("SMTP message must be one line, got %q", message)
	}
	if len(message) > maxDeliverySMTPMessageBytes {
		t.Fatalf("SMTP message is too long: %d bytes", len(message))
	}
	if !strings.Contains(message, "a.example") {
		t.Fatalf("SMTP message is not deterministically sorted: %q", message)
	}
}
