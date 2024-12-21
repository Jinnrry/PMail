package imap_server

import (
	"reflect"
	"testing"
)

func Test_splitCommand(t *testing.T) {
	type args struct {
		commands string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal",
			args: args{
				commands: `(UID ENVELOPE FLAGS INTERNALDATE RFC822.SIZE BODY.PEEK[HEADER.FIELDS ("date" "subject" "from" "to" "cc" "message-id" "in-reply-to" "references" "content-type" "x-priority" "x-uniform-type-identifier" "x-universally-unique-identifier" "list-id" "list-unsubscribe" "bimi-indicator" "bimi-location" "x-bimi-indicator-hash" "authentication-results" "dkim-signature")])`,
			},
			want: []string{
				"UID", "ENVELOPE", "FLAGS", "INTERNALDATE", "RFC822.SIZE",
				"BODY.PEEK[HEADER.FIELDS (\"date\" \"subject\" \"from\" \"to\" \"cc\" \"message-id\" \"in-reply-to\" \"references\" \"content-type\" \"x-priority\" \"x-uniform-type-identifier\" \"x-universally-unique-identifier\" \"list-id\" \"list-unsubscribe\" \"bimi-indicator\" \"bimi-location\" \"x-bimi-indicator-hash\" \"authentication-results\" \"dkim-signature\")]",
			},
		},
		{
			name: "fetch",
			args: args{
				commands: "(FLAGS UID)",
			},
			want: []string{"FLAGS", "UID"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitCommand(tt.args.commands, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
