package imap_server

import (
	"crypto/tls"
	"fmt"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"mime"
	"testing"
	"time"
)

func TestStarTLS(t *testing.T) {
	config.Init()
	db.Init("")
	go StarTLS()
	time.Sleep(2 * time.Second)

	options := &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client, err := imapclient.DialTLS("127.0.0.1:993", options)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("_________")

	res, err := client.Capability().Wait() // wait forever!
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Response:", res)

	time.Sleep(10 * time.Second)
}

/*
Here Is My Server Input&Output Log:
Output:	* OK [CAPABILITY IMAP4 IMAP4rev1 AUTH=PLAIN AUTH=LOGIN] PMail Server ready
Input:	T1 CAPABILITY
Output:	* CAPABILITY IMAP4rev1 UNSELECT IDLE AUTH=PLAIN AUTH=LOGIN
Output:	T1 OK success
*/
