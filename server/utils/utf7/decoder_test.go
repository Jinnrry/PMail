package utf7_test

import (
	"github.com/Jinnrry/pmail/utils/utf7"
	"strings"
	"testing"
)

var decode = []struct {
	in  string
	out string
	ok  bool
}{
	// Basics (the inverse test on encode checks other valid inputs)
	{"", "", true},
	{"abc", "abc", true},
	{"&-abc", "&abc", true},
	{"abc&-", "abc&", true},
	{"a&-b&-c", "a&b&c", true},
	{"&ABk-", "\x19", true},
	{"&AB8-", "\x1F", true},
	{"ABk-", "ABk-", true},
	{"&-,&-&AP8-&-", "&,&\u00FF&", true},
	{"&-&-,&AP8-&-", "&&,\u00FF&", true},
	{"abc &- &AP8A,wD,- &- xyz", "abc & \u00FF\u00FF\u00FF & xyz", true},

	// Illegal code point in ASCII
	{"\x00", "", false},
	{"\x1F", "", false},
	{"abc\n", "", false},
	{"abc\x7Fxyz", "", false},

	// Invalid UTF-8
	{"\xc3\x28", "", false},
	{"\xe2\x82\x28", "", false},

	// Invalid Base64 alphabet
	{"&/+8-", "", false},
	{"&*-", "", false},
	{"&ZeVnLIqe -", "", false},

	// CR and LF in Base64
	{"&ZeVnLIqe\r\n-", "", false},
	{"&ZeVnLIqe\r\n\r\n-", "", false},
	{"&ZeVn\r\n\r\nLIqe-", "", false},

	// Padding not stripped
	{"&AAAAHw=-", "", false},
	{"&AAAAHw==-", "", false},
	{"&AAAAHwB,AIA=-", "", false},
	{"&AAAAHwB,AIA==-", "", false},

	// One byte short
	{"&2A-", "", false},
	{"&2ADc-", "", false},
	{"&AAAAHwB,A-", "", false},
	{"&AAAAHwB,A=-", "", false},
	{"&AAAAHwB,A==-", "", false},
	{"&AAAAHwB,A===-", "", false},
	{"&AAAAHwB,AI-", "", false},
	{"&AAAAHwB,AI=-", "", false},
	{"&AAAAHwB,AI==-", "", false},

	// Implicit shift
	{"&", "", false},
	{"&Jjo", "", false},
	{"Jjo&", "", false},
	{"&Jjo&", "", false},
	{"&Jjo!", "", false},
	{"&Jjo+", "", false},
	{"abc&Jjo", "", false},

	// Null shift
	{"&AGE-&Jjo-", "", false},
	{"&U,BTFw-&ZeVnLIqe-", "", false},

	// Long input with Base64 at the end
	{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa &2D3eCg- &2D3eCw- &2D3eDg-",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \U0001f60a \U0001f60b \U0001f60e", true},

	// Long input in Base64 between short ASCII
	{"00000000000000000000 &MEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEI- 00000000000000000000",
		"00000000000000000000 " + strings.Repeat("\U00003042", 37) + " 00000000000000000000", true},

	// ASCII in Base64
	{"&AGE-", "", false},            // "a"
	{"&ACY-", "", false},            // "&"
	{"&AGgAZQBsAGwAbw-", "", false}, // "hello"
	{"&JjoAIQ-", "", false},         // "\u263a!"

	// Bad surrogate
	{"&2AA-", "", false},    // U+D800
	{"&2AD-", "", false},    // U+D800
	{"&3AA-", "", false},    // U+DC00
	{"&2AAAQQ-", "", false}, // U+D800 'A'
	{"&2AD,,w-", "", false}, // U+D800 U+FFFF
	{"&3ADYAA-", "", false}, // U+DC00 U+D800

	// Chinese
	{"&V4NXPpCuTvY-", "垃圾邮件", true},
	{"&UXZO1mWHTvZZOQ-", "其他文件夹", true},
}

func TestDecoder(t *testing.T) {
	for _, test := range decode {
		out, err := utf7.Decode(test.in)
		if out != test.out {
			t.Errorf("UTF7Decode(%+q) expected %+q; got %+q", test.in, test.out, out)
		}
		if test.ok {
			if err != nil {
				t.Errorf("UTF7Decode(%+q) unexpected error; %v", test.in, err)
			}
		} else if err == nil {
			t.Errorf("UTF7Decode(%+q) expected error", test.in)
		}
	}
}
