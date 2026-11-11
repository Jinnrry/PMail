package smtp_server

import "testing"

func TestSessionResetClearsState(t *testing.T) {
	session := &Session{
		From: "sender@example.com",
		To:   []string{"rcpt1@example.com", "rcpt2@example.com"},
	}

	session.Reset()

	if session.From != "" {
		t.Errorf("Reset() From = %q, want empty string", session.From)
	}
	if session.To != nil {
		t.Errorf("Reset() To = %v, want nil", session.To)
	}
}

func TestSessionResetIsIdempotent(t *testing.T) {
	session := &Session{From: "sender@example.com", To: []string{"rcpt@example.com"}}

	session.Reset()
	session.Reset()

	if session.From != "" || session.To != nil {
		t.Fatalf("Reset() left stale state: From=%q To=%v", session.From, session.To)
	}
}
