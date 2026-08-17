package smtp_server

import (
	"errors"
	"net"
	"net/textproto"
	"sort"
	"strings"
	"unicode/utf8"

	smtp "github.com/emersion/go-smtp"
)

const maxDeliverySMTPMessageBytes = 128

// downstreamDeliverySMTPError turns the synchronous downstream delivery result
// into the SMTP response returned to the submitting client. Unknown failures are
// temporary so a transient network problem does not permanently drop a message.
func downstreamDeliverySMTPError(deliveryErr error, domainErrors map[string]error) error {
	if deliveryErr == nil {
		return nil
	}

	permanent := len(domainErrors) > 0
	for _, domainErr := range domainErrors {
		if domainErr == nil || !isPermanentDeliveryError(domainErr) {
			permanent = false
			break
		}
	}

	if permanent {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 0, 0},
			Message:      deliverySMTPMessage("Downstream delivery permanently failed", deliveryErr, domainErrors),
		}
	}

	return &smtp.SMTPError{
		Code:         451,
		EnhancedCode: smtp.EnhancedCode{4, 4, 0},
		Message:      deliverySMTPMessage("Downstream delivery temporarily failed", deliveryErr, domainErrors),
	}
}

func deliverySMTPMessage(prefix string, deliveryErr error, domainErrors map[string]error) string {
	details := make([]string, 0, len(domainErrors))
	domains := make([]string, 0, len(domainErrors))
	for domain := range domainErrors {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	for _, domain := range domains {
		if domainErr := domainErrors[domain]; domainErr != nil {
			details = append(details, domain+": "+domainErr.Error())
		}
	}
	if len(details) == 0 && deliveryErr != nil {
		details = append(details, deliveryErr.Error())
	}

	message := prefix
	if len(details) > 0 {
		// SMTP replies must stay on one bounded line, even when a remote server
		// returns malformed text. Retain useful detail for the submitting client.
		message += ": " + strings.Join(strings.Fields(strings.Join(details, "; ")), " ")
	}
	if len(message) <= maxDeliverySMTPMessageBytes {
		return message
	}

	limit := maxDeliverySMTPMessageBytes - len("...")
	for len(message) > limit {
		_, size := utf8.DecodeLastRuneInString(message)
		message = message[:len(message)-size]
	}
	return message + "..."
}

func isPermanentDeliveryError(err error) bool {
	var protocolErr *textproto.Error
	if errors.As(err, &protocolErr) {
		return protocolErr.Code >= 500 && protocolErr.Code <= 599
	}

	var smtpErr *smtp.SMTPError
	if errors.As(err, &smtpErr) {
		return smtpErr.Code >= 500 && smtpErr.Code <= 599
	}

	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.IsNotFound
}
