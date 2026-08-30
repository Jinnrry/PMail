package main

import (
	"testing"

	"github.com/Jinnrry/pmail/dto/parsemail"
)

func TestBuildNotificationContent(t *testing.T) {
	tests := []struct {
		name  string
		email *parsemail.Email
		want  string
	}{
		{
			name:  "Chinese verification code in text",
			email: &parsemail.Email{Subject: "登录验证", Text: []byte("您的验证码为：123456，请勿泄露")},
			want:  "验证码:123456",
		},
		{
			name:  "verification code before keyword",
			email: &parsemail.Email{Subject: "847291 是您的验证码"},
			want:  "验证码:847291",
		},
		{
			name:  "alphanumeric English verification code",
			email: &parsemail.Email{Subject: "Sign in", Text: []byte("Your verification code is A7B9C2")},
			want:  "验证码:A7B9C2",
		},
		{
			name:  "verification code in HTML only",
			email: &parsemail.Email{Subject: "安全验证", HTML: []byte("<p>验证码：</p><strong>654321</strong>")},
			want:  "验证码:654321",
		},
		{
			name:  "regular email keeps original content",
			email: &parsemail.Email{Subject: "项目进度 2026", Text: []byte("今天已完成开发")},
			want:  "<<项目进度 2026>>  今天已完成开发",
		},
		{
			name:  "verification keyword without a code keeps original content",
			email: &parsemail.Email{Subject: "2026 年验证码安全规则更新", Text: []byte("请阅读新的安全规则")},
			want:  "<<2026 年验证码安全规则更新>>  请阅读新的安全规则",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildNotificationContent(tt.email); got != tt.want {
				t.Fatalf("buildNotificationContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
