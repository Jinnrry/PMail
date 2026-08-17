package send

import (
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"testing"
)

func TestIsPermanentSMTPResponse(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil"},
		{name: "network error", err: errors.New("connection refused")},
		{name: "temporary SMTP response", err: &textproto.Error{Code: 451, Msg: "try later"}},
		{name: "wrapped temporary SMTP response", err: fmt.Errorf("delivery: %w", &textproto.Error{Code: 421, Msg: "service unavailable"})},
		{name: "permanent SMTP response", err: &textproto.Error{Code: 550, Msg: "mailbox unavailable"}, want: true},
		{name: "wrapped permanent SMTP response", err: fmt.Errorf("delivery: %w", &textproto.Error{Code: 554, Msg: "transaction failed"}), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPermanentSMTPResponse(tt.err); got != tt.want {
				t.Fatalf("isPermanentSMTPResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeliveryFailureCausePreservesTemporaryMXError(t *testing.T) {
	tests := []struct {
		name      string
		lookupErr *net.DNSError
	}{
		{
			name:      "SERVFAIL before fallback NXDOMAIN",
			lookupErr: &net.DNSError{Err: "server misbehaving", Name: "example.com", IsTemporary: true},
		},
		{
			name:      "timeout before fallback NXDOMAIN",
			lookupErr: &net.DNSError{Err: "i/o timeout", Name: "example.com", IsTimeout: true},
		},
	}

	fallbackErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Err: "no such host", Name: "smtp.example.com", IsNotFound: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := deliveryFailureCause(tt.lookupErr, fallbackErr)

			var dnsErr *net.DNSError
			if !errors.As(err, &dnsErr) {
				t.Fatalf("deliveryFailureCause() = %T, want wrapped DNS error", err)
			}
			if dnsErr.IsNotFound {
				t.Fatalf("deliveryFailureCause() selected fallback NXDOMAIN: %v", dnsErr)
			}
			if !dnsErr.IsTemporary && !dnsErr.IsTimeout {
				t.Fatalf("deliveryFailureCause() selected non-temporary DNS error: %v", dnsErr)
			}
		})
	}
}

func TestDeliveryFailureCauseKeepsExplicitSMTPRejection(t *testing.T) {
	lookupErr := &net.DNSError{Err: "server misbehaving", Name: "example.com", IsTemporary: true}
	rejection := &textproto.Error{Code: 550, Msg: "mailbox unavailable"}

	if got := deliveryFailureCause(lookupErr, rejection); got != rejection {
		t.Fatalf("deliveryFailureCause() = %v, want explicit SMTP rejection %v", got, rejection)
	}
}
