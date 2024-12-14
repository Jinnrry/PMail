package goimap

import (
	"reflect"
	"testing"
)

func Test_paramsErr(t *testing.T) {

}

func Test_getCommand(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			"STATUS命令测试",
			args{`15.64 STATUS "Deleted Messages" (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`},
			"15.64",
			"STATUS",
			`"Deleted Messages" (MESSAGES UIDNEXT UIDVALIDITY UNSEEN)`,
		},
		{
			"SELECT命令测试",
			args{`9.79 SELECT INBOX`},
			"9.79",
			"SELECT",
			`INBOX`,
		},
		{
			"CAPABILITY命令测试",
			args{`1.81 CAPABILITY`},
			"1.81",
			"CAPABILITY",
			``,
		},
		{
			"异常命令测试",
			args{`GET/HTTP/1.0`},
			"",
			"",
			``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := getCommand(tt.args.line)
			if got != tt.want {
				t.Errorf("getCommand() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCommand() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("getCommand() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
