package list

import (
	"strings"
	"time"

	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/models"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	log "github.com/sirupsen/logrus"
)

// IMAPSearchCriteria represents the search criteria for IMAP SEARCH command
type IMAPSearchCriteria struct {
	// UID and sequence number sets
	UIDs    []imap.UIDSet
	SeqNums []imap.SeqSet

	// Date filters (only date is used, time and timezone are ignored)
	Since      time.Time // Internal date since
	Before     time.Time // Internal date before
	SentSince  time.Time // Sent date since
	SentBefore time.Time // Sent date before

	// Header field searches
	Header []HeaderField // Key-Value pairs for header search

	// Content search
	Body []string // Search in body
	Text []string // Search in headers + body

	// Flag filters
	Flag    []imap.Flag // Messages with these flags
	NotFlag []imap.Flag // Messages without these flags

	// Size filters
	Larger  int64 // Messages larger than this (bytes)
	Smaller int64 // Messages smaller than this (bytes)

	// Logical combinations
	Not []IMAPSearchCriteria
	Or  [][2]IMAPSearchCriteria
}

// HeaderField represents a header field search criterion
type HeaderField struct {
	Key   string
	Value string
}

// IMAPSearchResult represents a search result item
type IMAPSearchResult struct {
	UID          int  // user_email.id as UID
	SeqNum       int  // sequence number
	EmailID      int  // email.id
	IsRead       int8 // read status
	SerialNumber int  // serial number
}

// SearchEmails performs IMAP search based on criteria
func SearchEmails(ctx *context.Context, groupName string, criteria *imap.SearchCriteria) ([]*response.UserEmailUIDData, error) {
	// First get the base list for the mailbox
	baseList := GetUEListByUID(ctx, groupName, 0, 0, nil)
	if len(baseList) == 0 {
		return baseList, nil
	}

	// If no criteria specified, return all
	if criteria == nil || isEmptyCriteria(criteria) {
		return baseList, nil
	}

	// Build UID to sequence number mapping
	uidToSeq := make(map[int]int)
	seqToUID := make(map[int]int)
	for _, item := range baseList {
		uidToSeq[item.ID] = item.SerialNumber
		seqToUID[item.SerialNumber] = item.ID
	}

	// Filter by UID sets
	if len(criteria.UID) > 0 {
		baseList = filterByUIDSets(baseList, criteria.UID)
	}

	// Filter by sequence number sets
	if len(criteria.SeqNum) > 0 {
		baseList = filterBySeqNumSets(baseList, criteria.SeqNum)
	}

	// For more complex filters, we need to fetch email data
	if needsEmailData(criteria) {
		baseList = filterWithEmailData(ctx, baseList, criteria)
	}

	// Filter by flags (is_read status in this implementation)
	if len(criteria.Flag) > 0 || len(criteria.NotFlag) > 0 {
		baseList = filterByFlags(baseList, criteria.Flag, criteria.NotFlag)
	}

	// Handle NOT criteria
	if len(criteria.Not) > 0 {
		for _, notCriteria := range criteria.Not {
			baseList = applyNotCriteria(ctx, groupName, baseList, &notCriteria)
		}
	}

	// Handle OR criteria
	if len(criteria.Or) > 0 {
		baseList = applyOrCriteria(ctx, groupName, baseList, criteria.Or)
	}

	return baseList, nil
}

// isEmptyCriteria checks if the search criteria is empty
func isEmptyCriteria(criteria *imap.SearchCriteria) bool {
	return len(criteria.UID) == 0 &&
		len(criteria.SeqNum) == 0 &&
		criteria.Since.IsZero() &&
		criteria.Before.IsZero() &&
		criteria.SentSince.IsZero() &&
		criteria.SentBefore.IsZero() &&
		len(criteria.Header) == 0 &&
		len(criteria.Body) == 0 &&
		len(criteria.Text) == 0 &&
		len(criteria.Flag) == 0 &&
		len(criteria.NotFlag) == 0 &&
		criteria.Larger == 0 &&
		criteria.Smaller == 0 &&
		len(criteria.Not) == 0 &&
		len(criteria.Or) == 0
}

// needsEmailData checks if we need to load email data for filtering
func needsEmailData(criteria *imap.SearchCriteria) bool {
	return !criteria.Since.IsZero() ||
		!criteria.Before.IsZero() ||
		!criteria.SentSince.IsZero() ||
		!criteria.SentBefore.IsZero() ||
		len(criteria.Header) > 0 ||
		len(criteria.Body) > 0 ||
		len(criteria.Text) > 0 ||
		criteria.Larger > 0 ||
		criteria.Smaller > 0
}

// filterByUIDSets filters the list by UID sets
func filterByUIDSets(list []*response.UserEmailUIDData, uidSets []imap.UIDSet) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		for _, uidSet := range uidSets {
			if uidSet.Contains(imap.UID(item.ID)) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// filterBySeqNumSets filters the list by sequence number sets
func filterBySeqNumSets(list []*response.UserEmailUIDData, seqSets []imap.SeqSet) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		for _, seqSet := range seqSets {
			if seqSet.Contains(uint32(item.SerialNumber)) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// filterByFlags filters by message flags
func filterByFlags(list []*response.UserEmailUIDData, flags []imap.Flag, notFlags []imap.Flag) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		match := true

		// Check required flags
		for _, flag := range flags {
			if !hasFlag(item, flag) {
				match = false
				break
			}
		}

		// Check flags that should NOT be present
		if match {
			for _, flag := range notFlags {
				if hasFlag(item, flag) {
					match = false
					break
				}
			}
		}

		if match {
			result = append(result, item)
		}
	}
	return result
}

// hasFlag checks if a message has a specific flag
func hasFlag(item *response.UserEmailUIDData, flag imap.Flag) bool {
	switch flag {
	case imap.FlagSeen:
		return item.IsRead == 1
	case imap.FlagDeleted:
		return item.Status == 3
	case imap.FlagDraft:
		return item.Status == 4
	case imap.FlagJunk:
		return item.Status == 5
	// For flags we don't track, return false
	case imap.FlagAnswered, imap.FlagFlagged:
		return false
	default:
		return false
	}
}

// filterWithEmailData loads email data and applies filters that need it
func filterWithEmailData(ctx *context.Context, list []*response.UserEmailUIDData, criteria *imap.SearchCriteria) []*response.UserEmailUIDData {
	if len(list) == 0 {
		return list
	}

	// Get email IDs
	var emailIDs []int
	ueMap := make(map[int]*response.UserEmailUIDData) // emailID -> UserEmailUIDData
	for _, item := range list {
		emailIDs = append(emailIDs, item.EmailID)
		ueMap[item.EmailID] = item
	}

	// Fetch emails from database
	var emails []models.Email
	err := db.Instance.Table("email").In("id", emailIDs).Find(&emails)
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to fetch emails for search: %v", err)
		return list
	}

	// Build email map
	emailMap := make(map[int]*models.Email)
	for i := range emails {
		emailMap[emails[i].Id] = &emails[i]
	}

	// Filter
	var result []*response.UserEmailUIDData
	for _, item := range list {
		email, ok := emailMap[item.EmailID]
		if !ok {
			continue
		}

		if matchesEmailCriteria(email, criteria) {
			result = append(result, item)
		}
	}

	return result
}

// matchesEmailCriteria checks if an email matches the search criteria
func matchesEmailCriteria(email *models.Email, criteria *imap.SearchCriteria) bool {
	// Date filters (internal date = CreateTime)
	if !criteria.Since.IsZero() {
		if email.CreateTime.Before(truncateToDate(criteria.Since)) {
			return false
		}
	}
	if !criteria.Before.IsZero() {
		if !email.CreateTime.Before(truncateToDate(criteria.Before)) {
			return false
		}
	}

	// Sent date filters
	if !criteria.SentSince.IsZero() {
		if email.SendDate.Before(truncateToDate(criteria.SentSince)) {
			return false
		}
	}
	if !criteria.SentBefore.IsZero() {
		if !email.SendDate.Before(truncateToDate(criteria.SentBefore)) {
			return false
		}
	}

	// Size filters
	if criteria.Larger > 0 {
		if int64(email.Size) <= criteria.Larger {
			return false
		}
	}
	if criteria.Smaller > 0 {
		if int64(email.Size) >= criteria.Smaller {
			return false
		}
	}

	// Header field search
	for _, hf := range criteria.Header {
		if !matchesHeader(email, hf.Key, hf.Value) {
			return false
		}
	}

	// Body search
	for _, pattern := range criteria.Body {
		if !matchesBody(email, pattern) {
			return false
		}
	}

	// Text search (headers + body)
	for _, pattern := range criteria.Text {
		if !matchesText(email, pattern) {
			return false
		}
	}

	return true
}

// truncateToDate removes time component from a time.Time
func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// matchesHeader checks if an email matches a header field search
func matchesHeader(email *models.Email, key, value string) bool {
	key = strings.ToLower(key)
	value = strings.ToLower(value)

	switch key {
	case "subject":
		return strings.Contains(strings.ToLower(email.Subject), value)
	case "from":
		return strings.Contains(strings.ToLower(email.FromAddress), value) ||
			strings.Contains(strings.ToLower(email.FromName), value)
	case "to":
		return strings.Contains(strings.ToLower(email.To), value)
	case "cc":
		return strings.Contains(strings.ToLower(email.Cc), value)
	case "bcc":
		return strings.Contains(strings.ToLower(email.Bcc), value)
	case "reply-to":
		return strings.Contains(strings.ToLower(email.ReplyTo), value)
	case "sender":
		return strings.Contains(strings.ToLower(email.Sender), value)
	default:
		// For unknown headers, we can't match
		return false
	}
}

// matchesBody checks if an email body matches the pattern
func matchesBody(email *models.Email, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// Check text body
	if email.Text.Valid && strings.Contains(strings.ToLower(email.Text.String), pattern) {
		return true
	}

	// Check HTML body
	if email.Html.Valid && strings.Contains(strings.ToLower(email.Html.String), pattern) {
		return true
	}

	return false
}

// matchesText checks if an email (headers + body) matches the pattern
func matchesText(email *models.Email, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// Check headers
	if strings.Contains(strings.ToLower(email.Subject), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.FromAddress), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.FromName), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.To), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.Cc), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.Bcc), pattern) {
		return true
	}

	// Check body
	return matchesBody(email, pattern)
}

// applyNotCriteria applies NOT criteria
func applyNotCriteria(ctx *context.Context, groupName string, list []*response.UserEmailUIDData, notCriteria *imap.SearchCriteria) []*response.UserEmailUIDData {
	// Get the list of items that match the NOT criteria
	matchedList, _ := SearchEmails(ctx, groupName, notCriteria)

	// Build a set of matched UIDs
	matchedUIDs := make(map[int]bool)
	for _, item := range matchedList {
		matchedUIDs[item.ID] = true
	}

	// Return items that are NOT in the matched set
	var result []*response.UserEmailUIDData
	for _, item := range list {
		if !matchedUIDs[item.ID] {
			result = append(result, item)
		}
	}

	return result
}

// applyOrCriteria applies OR criteria
func applyOrCriteria(ctx *context.Context, groupName string, list []*response.UserEmailUIDData, orCriteria [][2]imap.SearchCriteria) []*response.UserEmailUIDData {
	if len(orCriteria) == 0 {
		return list
	}

	// Build a set of current UIDs for intersection
	currentUIDs := make(map[int]bool)
	for _, item := range list {
		currentUIDs[item.ID] = true
	}

	// For each OR pair, find items matching either condition
	resultUIDs := make(map[int]bool)

	for _, pair := range orCriteria {
		// Get matches for first condition
		matches1, _ := SearchEmails(ctx, groupName, &pair[0])
		for _, item := range matches1 {
			if currentUIDs[item.ID] {
				resultUIDs[item.ID] = true
			}
		}

		// Get matches for second condition
		matches2, _ := SearchEmails(ctx, groupName, &pair[1])
		for _, item := range matches2 {
			if currentUIDs[item.ID] {
				resultUIDs[item.ID] = true
			}
		}
	}

	// Build result from original list preserving order
	var result []*response.UserEmailUIDData
	for _, item := range list {
		if resultUIDs[item.ID] {
			result = append(result, item)
		}
	}

	return result
}
