package smtp_server

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"net"
	"net/netip"
	"os"
	"pmail/config"
	"pmail/db"
	parsemail2 "pmail/dto/parsemail"
	"pmail/hooks"
	"pmail/session"
	"pmail/utils/context"
	"testing"
	"time"
)

func testInit() {
	// 设置日志格式为json格式
	//log.SetFormatter(&log.JSONFormatter{})

	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		//以下设置只是为了使输出更美观
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:03:04",
	})

	// 设置将日志输出到标准输出（默认的输出为stderr,标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	log.SetLevel(log.ErrorLevel)

	var cst, _ = time.LoadLocation("Asia/Shanghai")
	time.Local = cst

	config.Init()
	config.Instance.DkimPrivateKeyPath = "../config/dkim/dkim.priv"
	config.Instance.DbType = config.DBTypeSQLite
	config.Instance.DbDSN = "../config/pmail_temp.db"

	parsemail2.Init()
	db.Init("")
	session.Init()
	hooks.Init("dev")
}

func TestPmailEmail(t *testing.T) {
	testInit()
	emailData := `DKIM-Signature: a=rsa-sha256; bh=x7Rh+N2y2K9exccEAyKCTAGDgYKfnLZpMWc25ug5Ny4=;
 c=simple/simple; d=domain.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693831868;
 v=1;
 b=1PZEupYvSMtGyYx42b4G65YbdnRj4y2QFo9kS7GXiTVhUM5EYzJhZzknwRMN5RL5aFY26W4E
 DmzJ85XvPPvrDtnU/B4jkc5xthE+KEsb1Go8HcL8WQqwvsE9brepeA0t0RiPnA/x7dbTo3u72SG
 WqtviWbJH5lPFc9PkSbEPFtc=
Content-Type: multipart/mixed;
 boundary=3c13260efb7bd8bad8315c21215489fe283f36cdf82813674f6e11215f6c
Mime-Version: 1.0
Subject: =?utf-8?q?=E6=8F=92=E4=BB=B6=E6=B5=8B=E8=AF=95?=
To: =?utf-8?q?=E5=90=8D?= <ok@jinnrry.com>
From: =?utf-8?q?=E5=8F=91=E9=80=81=E4=BA=BA?= <j@jinnrry.com>
Date: Mon, 04 Sep 2023 20:51:08 +0800

--3c13260efb7bd8bad8315c21215489fe283f36cdf82813674f6e11215f6c
Content-Type: multipart/alternative;
 boundary=9ebf2f3c4f97c51dd9a285ae28a54d2d0d84aa6d0ad28b76547e2096bb66

--9ebf2f3c4f97c51dd9a285ae28a54d2d0d84aa6d0ad28b76547e2096bb66
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

=E8=BF=99=E6=98=AFText
--9ebf2f3c4f97c51dd9a285ae28a54d2d0d84aa6d0ad28b76547e2096bb66
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<div>=E8=BF=99=E6=98=AFHtml</div>
--9ebf2f3c4f97c51dd9a285ae28a54d2d0d84aa6d0ad28b76547e2096bb66--

--3c13260efb7bd8bad8315c21215489fe283f36cdf82813674f6e11215f6c--
`
	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx: &context.Context{
			UserID:      0,
			UserName:    "",
			UserAccount: "",
		},
		To: []string{"ok@jinnrry.com"},
	}

	s.Data(bytes.NewReader([]byte(emailData)))

}

func TestRuleForward(t *testing.T) {
	testInit()

	forwardEmail := `DKIM-Signature: a=rsa-sha256; bh=bpOshF+iimuqAQijVxqkH6gPpWf8A+Ih30/tMjgEgS0=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992640;
 v=1;
 b=XiOgYL9iGrkuYzXBAf7DSO0sRbFr6aPOE4VikmselNKEF1UTjMPdiqpeHyx/i6BOQlJWWZEC
 PzceHTDFIStcZE6a5Sc1nh8Fis+gRkrheBO/zK/P5P/euK+0Fj5+0T82keNTSCgo1ZtEIubaNR0
 JvkwJ2ZC9g8xV6Yiq+ZhRriT8lZ6zeI55PPEFJIzFgZ7xDshDgx5E7J1xRXQqcEMV1rgVq04d3c
 6wjU+LLtghmgtUToRp3ASn6DhVO+Bbc4QkmcQ/StQH3681+1GVMHvQSBhSSymSRA71SikE2u3a1
 JnvbOP9fThP7h+6oFEIRuF7MwDb3JWY5BXiFFKCkecdFg==
Content-Type: multipart/mixed;
 boundary=8e9d5abb6bdac11b8d7d6e13280af1a87d12b904a59368d6e852b0a4ce3e
Mime-Version: 1.0
Subject: forward
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:30:40 +0800

--8e9d5abb6bdac11b8d7d6e13280af1a87d12b904a59368d6e852b0a4ce3e
Content-Type: multipart/alternative;
 boundary=a62ae91c159ea22e8196d57d344626eb00d1ddfa9c5064a39b01588aa992

--a62ae91c159ea22e8196d57d344626eb00d1ddfa9c5064a39b01588aa992
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

hello pls Forward the email.
--a62ae91c159ea22e8196d57d344626eb00d1ddfa9c5064a39b01588aa992
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>hello pls Forward the email.</p>
--a62ae91c159ea22e8196d57d344626eb00d1ddfa9c5064a39b01588aa992--

--8e9d5abb6bdac11b8d7d6e13280af1a87d12b904a59368d6e852b0a4ce3e--`

	readEmail := `DKIM-Signature: a=rsa-sha256; bh=JcCDj6edb1bAwRbcFZ63plFZOeB5AdGWLE/PQ2FQ1Tc=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992600;
 v=1;
 b=rwlqSkDFKYH42pA1jsajemaw+4YdeLHPeqV4mLQrRdihgma1VSvXl5CEOur/KuwQuUarr2cu
 SntWrHE6+RnDaQcPEHbkgoMjEJw5+VPwkIvE6VSlMIB7jg93mGzvN2yjheWTePZ+cVPjOaIrgir
 wiT24hkrTHp+ONT8XoS0sDuY+ieyBZp/GCv/YvgE4t0JEkNozMAVWotrXxaICDzZoWP3NNmKLqg
 6He6zwWAl51r3W5R5weGBi6A/FqlHgHZGroXnNi+wolDuN6pQiVAJ7MZ6hboPCbCCRrBQDTdor5
 wEI2+MwlJ/d2f17wxoGmluCewbeYttuVcpUOVwACJKw3g==
Content-Type: multipart/mixed;
 boundary=9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3
Mime-Version: 1.0
Subject: read
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:30:00 +0800

--9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3
Content-Type: multipart/alternative;
 boundary=54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884

--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

12 aRead 1sadf
--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>12 aRead 1sadf</p>
--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884--

--9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3--`

	moveEmail := `DKIM-Signature: a=rsa-sha256; bh=YQfG/wlHGhky6FNmpIwgDYDOc/uyivdBv+9S02Z04xY=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992542;
 v=1;
 b=IhxswOCq8I7CmCas1EMp+n8loR7illqlF0IJC6eN1+OLjI/E5BPzpP4HWkyqaAkd0Vn9i+Bn
 MVb5kNHZ2S7qt0rqAAc6Atc0i9WpLEI3Cng+VDn+difcMZlJSAkhLLn2sUsS4Fzqqo3Cbw62qSO
 TgnWRmlj9aM+5xfGcl/76WOvQQpahJbGg6Go51kFMeHVom/VeGKIgFBCeMe37T/LS03c3pAV8gA
 i6Zy3GYE57W/qU3oCzaGeS3n5zom/i74H4VipiVIMX/OBNYhdHWrP8vyjvzLFpJlXp6RvzcRl0P
 ytyiCZfE8G7fAFntp20LW70Y5Xgqqczk1jR578UDczVoA==
Content-Type: multipart/mixed;
 boundary=c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d
Mime-Version: 1.0
Subject: Move
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:29:02 +0800

--c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d
Content-Type: multipart/alternative;
 boundary=a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966

--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

MOVE move Move
--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>MOVE move Move</p>
--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966--

--c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d--`

	deleteEmail := `DKIM-Signature: a=rsa-sha256; bh=dNtHGqd1NbRj0WSwrJmPsqAcAy3h/4kZK2HFQ0Asld8=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992495;
 v=1;
 b=QllU8lqGdoOMaGYp8d13oWytb7+RebqKjq4y8Rs/kOeQxoE8dSEVliK3eBiXidsNTdDtkTqf
 eiwjyRBK92NVCYprdJqLbu9qZ39BC2lk3NXttTSJ1+1ZZ/bGtIW5JIYn2pToED0MqVVkxGFUtl+
 qFmc4mWo5a4Mbij7xaAB3uJtHpBDt7q4Ovr2hiMetQv7YrhZvCt/xrH8Q9YzZ6xzFUL5ekW40eH
 oWElU1GyVBHWCKh31aweyhA+1XLPYojjREQYd4svRqTbSFSsBqFwFIUGdnyJh2WgmF8eucmttAw
 oRhgzyZkHL1jAskKFBpO10SDReyk50Cvc+0kSLj+QcUpg==
Content-Type: multipart/mixed;
 boundary=bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3
Mime-Version: 1.0
Subject: test
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:28:15 +0800

--bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3
Content-Type: multipart/alternative;
 boundary=7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04

--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

Delete
--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>Delete</p>
--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04--

--bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3--`

	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(deleteEmail)))
	s.Data(bytes.NewReader([]byte(readEmail)))
	s.Data(bytes.NewReader([]byte(forwardEmail)))
	s.Data(bytes.NewReader([]byte(moveEmail)))
}

func TestRuleRead(t *testing.T) {
	testInit()

	readEmail := `DKIM-Signature: a=rsa-sha256; bh=JcCDj6edb1bAwRbcFZ63plFZOeB5AdGWLE/PQ2FQ1Tc=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992600;
 v=1;
 b=rwlqSkDFKYH42pA1jsajemaw+4YdeLHPeqV4mLQrRdihgma1VSvXl5CEOur/KuwQuUarr2cu
 SntWrHE6+RnDaQcPEHbkgoMjEJw5+VPwkIvE6VSlMIB7jg93mGzvN2yjheWTePZ+cVPjOaIrgir
 wiT24hkrTHp+ONT8XoS0sDuY+ieyBZp/GCv/YvgE4t0JEkNozMAVWotrXxaICDzZoWP3NNmKLqg
 6He6zwWAl51r3W5R5weGBi6A/FqlHgHZGroXnNi+wolDuN6pQiVAJ7MZ6hboPCbCCRrBQDTdor5
 wEI2+MwlJ/d2f17wxoGmluCewbeYttuVcpUOVwACJKw3g==
Content-Type: multipart/mixed;
 boundary=9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3
Mime-Version: 1.0
Subject: read
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:30:00 +0800

--9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3
Content-Type: multipart/alternative;
 boundary=54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884

--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

12 aRead 1sadf
--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>12 aRead 1sadf</p>
--54a95f3429f3cdb342383db10293780bed341f8dc20d2f876eb0853e3884--

--9e33a130a8a976102a93e296d6408d228e151f7841ca9ee0d777234fd6f3--`

	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(readEmail)))

}

func TestRuleDelete(t *testing.T) {
	testInit()

	deleteEmail := `DKIM-Signature: a=rsa-sha256; bh=dNtHGqd1NbRj0WSwrJmPsqAcAy3h/4kZK2HFQ0Asld8=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992495;
 v=1;
 b=QllU8lqGdoOMaGYp8d13oWytb7+RebqKjq4y8Rs/kOeQxoE8dSEVliK3eBiXidsNTdDtkTqf
 eiwjyRBK92NVCYprdJqLbu9qZ39BC2lk3NXttTSJ1+1ZZ/bGtIW5JIYn2pToED0MqVVkxGFUtl+
 qFmc4mWo5a4Mbij7xaAB3uJtHpBDt7q4Ovr2hiMetQv7YrhZvCt/xrH8Q9YzZ6xzFUL5ekW40eH
 oWElU1GyVBHWCKh31aweyhA+1XLPYojjREQYd4svRqTbSFSsBqFwFIUGdnyJh2WgmF8eucmttAw
 oRhgzyZkHL1jAskKFBpO10SDReyk50Cvc+0kSLj+QcUpg==
Content-Type: multipart/mixed;
 boundary=bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3
Mime-Version: 1.0
Subject: test
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:28:15 +0800

--bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3
Content-Type: multipart/alternative;
 boundary=7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04

--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

Delete
--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>Delete</p>
--7352524eaae801790245f6bf095460fd1f4e01f5748b4dba48635bf59b04--

--bdfa9bf94e22e218105281e06bd59bd6df3ce70e71367bf49fbe73301af3--`

	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(deleteEmail)))

}

func TestNullCC(t *testing.T) {
	testInit()

	emailData := `Date: Mon, 29 Jan 2024 16:54:30 +0800
Return-Path: 1231@111.com
From: =?utf-8?B?b2VhdHY=?= 1231@111.com
To: =?utf-8?B?ODQ2ODAzOTY=?= 123213@qq.com
Cc:
Bcc:
Reply-To: <>
Subject: =?utf-8?B?6L+Z5piv5LiA5bCB5p2l6IeqUmVsYXhEcmFtYeeahOmCruS7tg==?=
Message-ID: <cf43cc780b72dad392d4f90dfced88a8@1231@111.com>
X-Priority: 3
X-Mailer: Mailer (https://github.com/txthinking/Mailer)
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="6edc2ef285d93010a080caccc858c67b"

--6edc2ef285d93010a080caccc858c67b
Content-Type: text/plain; charset="UTF-8"
Content-Transfer-Encoding: base64

PGRpdiBzdHlsZT0ibWluLWhlaWdodDo1NTBweDsgcGFkZGluZzogMTAwcHggNTVweCAyMDBweDsi
Pui/meaYr+S4gOWwgeadpeiHqlJlbGF4RHJhbWHnmoTmoKHpqozpgq7ku7Ys55So5LqO5qCh6aqM
6YKu5Lu26YWN572u5piv5ZCm5q2j5bi4ITwvZGl2Pg==

--6edc2ef285d93010a080caccc858c67b
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: base64

PGRpdiBzdHlsZT0ibWluLWhlaWdodDo1NTBweDsgcGFkZGluZzogMTAwcHggNTVweCAyMDBweDsi
Pui/meaYr+S4gOWwgeadpeiHqlJlbGF4RHJhbWHnmoTmoKHpqozpgq7ku7Ys55So5LqO5qCh6aqM
6YKu5Lu26YWN572u5piv5ZCm5q2j5bi4ITwvZGl2Pg==

--6edc2ef285d93010a080caccc858c67b--`
	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(emailData)))
}

func TestRuleMove(t *testing.T) {
	testInit()

	moveEmail := `DKIM-Signature: a=rsa-sha256; bh=YQfG/wlHGhky6FNmpIwgDYDOc/uyivdBv+9S02Z04xY=;
 c=simple/simple; d=jinnrry.com;
 h=Content-Type:Mime-Version:Subject:To:From:Date; s=default; t=1693992542;
 v=1;
 b=IhxswOCq8I7CmCas1EMp+n8loR7illqlF0IJC6eN1+OLjI/E5BPzpP4HWkyqaAkd0Vn9i+Bn
 MVb5kNHZ2S7qt0rqAAc6Atc0i9WpLEI3Cng+VDn+difcMZlJSAkhLLn2sUsS4Fzqqo3Cbw62qSO
 TgnWRmlj9aM+5xfGcl/76WOvQQpahJbGg6Go51kFMeHVom/VeGKIgFBCeMe37T/LS03c3pAV8gA
 i6Zy3GYE57W/qU3oCzaGeS3n5zom/i74H4VipiVIMX/OBNYhdHWrP8vyjvzLFpJlXp6RvzcRl0P
 ytyiCZfE8G7fAFntp20LW70Y5Xgqqczk1jR578UDczVoA==
Content-Type: multipart/mixed;
 boundary=c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d
Mime-Version: 1.0
Subject: Move
To: <t@jiangwei.one>
From: "i" <i@jinnrry.com>
Date: Wed, 06 Sep 2023 17:29:02 +0800

--c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d
Content-Type: multipart/alternative;
 boundary=a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966

--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/plain

MOVE move Move
--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966
Content-Transfer-Encoding: quoted-printable
Content-Disposition: inline
Content-Type: text/html

<p>MOVE move Move</p>
--a69985ebcf3c1c44d6e69e5a29c1044743cd9e44d4bc9bb6886f83a73966--

--c84d60b253aa6caee345c73e717ad59b1975448bbdfad7a23ac4d76e022d--`

	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(moveEmail)))
}

func TestQAEmailForward(t *testing.T) {
	testInit()
	data := `Mime-Version: 1.0
X-QQ-MIME: TCMime 1.0 by Tencent
X-Mailer: QQMail 2.x
X-QQ-Mailer: QQMail 2.x
Message-ID: tencent_D82739970C66D2BFBA23F4A3@qq.com
Subject: =?UTF-8?B?5rWL6K+V5Y+R6YCB?=
Date: Wed, 10 Apr 2024 11:11:12 +0800 (GMT+08:00)
From: =?UTF-8?B?YWRtaW5AamlubnJyeS5jb20=?=<admin@jinnrry.com>
To: =?UTF-8?B??=<test@jinnrry.com>
Content-Type: multipart/alternative; 
        boundary="----=_Part_174_107154538.1712718674768"

------=_Part_174_107154538.1712718674768
Content-Type: text/plain; charset=us-ascii
Content-Transfer-Encoding: base64


------=_Part_174_107154538.1712718674768
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

PGRpdj7ov5nph4zmmK/lhoXlrrk8L2Rpdj48ZGl2PjwhLS1lbXB0eXNpZ24tLT48L2Rpdj4=
------=_Part_174_107154538.1712718674768--`

	s := Session{
		RemoteAddress: net.TCPAddrFromAddrPort(netip.AddrPortFrom(netip.AddrFrom4([4]byte{}), 25)),
		Ctx:           &context.Context{},
	}

	s.Data(bytes.NewReader([]byte(data)))
}
