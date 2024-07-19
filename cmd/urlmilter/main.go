package main

import (
	"bytes"
	"flag"
	"log"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strings"

	"github.com/phalaaxx/milter"
)

/* UrlMilter object */
type UrlMilter struct {
	milter.Milter
	multipart bool

	headers    textproto.MIMEHeader
	subjCount  int
	topicCount int

	message *bytes.Buffer
}

/* handle headers one by one */
func (e *UrlMilter) Header(name, value string, m *milter.Modifier) (milter.Response, error) {
	// if message has multiple parts set processing flag to true
	if name == "Content-Type" && strings.HasPrefix(strings.ToLower(value), "multipart/") {
		e.multipart = true
	}
	if name == "Subject" {
		e.subjCount++
	}
	if name == "Thread-Topic" {
		e.topicCount++
	}
	return milter.RespContinue, nil
}

/* at end of headers initialize message buffer and add headers to it */
func (e *UrlMilter) Headers(headers textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	// return accept if not a multipart message
	//if !e.multipart {
	//	return milter.RespAccept, nil
	//}
	// TODO issue with stalwart interop - accept after eoh is not parsed?

	// prepare message buffer
	e.message = new(bytes.Buffer)

	// TODO stalwart bug? body contains headers again..
	/*// print headers to message buffer
	for k, vl := range headers {
		for _, v := range vl {
			if _, err := fmt.Fprintf(e.message, "%s: %s\n", k, v); err != nil {
				return nil, err
			}
		}
	}
	if _, err := fmt.Fprintf(e.message, "\n"); err != nil {
		return nil, err
	}*/

	e.headers = headers

	// continue with milter processing
	return milter.RespContinue, nil
}

// accept body chunk
func (e *UrlMilter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {
	//log.Printf("Body chunk: \n%s\n", chunk)

	// save chunk to buffer
	if _, err := e.message.Write(chunk); err != nil {
		return nil, err
	}
	return milter.RespContinue, nil
}

/* Body is called when email message body has been sent */
func (e *UrlMilter) Body(m *milter.Modifier) (milter.Response, error) {

	if e.subjCount < 1 {
		err := m.AddHeader("Subject", "[EXTERNAL] - Empty Subject")
		if err != nil {
			return nil, err
		}
	}

	originalSubject := e.headers.Get("Subject")
	originalThreadTopic := e.headers.Get("Thread-Topic")

	// Unfortunately, RFC is often not followed in favor of localization
	// Source: https://en.wikipedia.org/wiki/List_of_email_subject_abbreviations#Abbreviations_in_other_languages
	// Avoid chains of Re: [EXTERNAL] and replace with a single "Re: "
	re := regexp.MustCompile(`(?i)((رد|回复|回覆|SV|Antw|VS|REF|RE|AW|ΑΠ|ΣΧΕΤ|השב|תשובה|Vá|R|RIF|BLS|Atb\.|RES|Odp|பதில்|YNT|ATB):\s+\[EXTERNAL\]\s+((رد|回复|回覆|SV|Antw|VS|REF|RE|AW|ΑΠ|ΣΧΕΤ|השב|תשובה|Vá|R|RIF|BLS|Atb\.|RES|Odp|பதில்|YNT|ATB):\s+)?)+`)
	cleanSubject := string(re.ReplaceAll([]byte(originalSubject), []byte("Re: ")))
	cleanTopic := string(re.ReplaceAll([]byte(originalThreadTopic), []byte("Re: ")))

	for i := 1; i <= e.subjCount; i++ {
		err := m.ChangeHeader(i, "Subject", "[EXTERNAL] "+cleanSubject)
		if err != nil {
			return nil, err
		}
	}
	for i := 1; i <= e.topicCount; i++ {
		err := m.ChangeHeader(i, "Thread-Topic", "[EXTERNAL] "+cleanTopic)
		if err != nil {
			return nil, err
		}
	}

	// prepare buffer

	mailCopy, err := os.CreateTemp("/tmp", "urlmilter-mail")
	if err != nil {
		return nil, err
	}
	_, err = mailCopy.Write(e.message.Bytes())
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(e.message.Bytes())

	// parse email message and get accept flag
	if newBody, err := RewriteEmailMessage(buffer); err != nil {
		log.Println("Error rewriting message:", err)
		log.Println("Message copy is in ", mailCopy.Name())
		return milter.RespAccept, nil
	} else {
		//log.Printf("%s", newBody)
		err = m.ReplaceBody(newBody)
		if err != nil {
			log.Println("Error replacing body: ", err)
			return milter.RespAccept, nil
		}
	}

	os.RemoveAll(mailCopy.Name())

	// accept message by default
	return milter.RespAccept, nil
}

/* NewObject creates new ExtMilter instance */
func RunServer(socket net.Listener) {
	// declare milter init function
	init := func() (milter.Milter, milter.OptAction, milter.OptProtocol) {
		return &UrlMilter{},
			milter.OptAddHeader | milter.OptChangeHeader | milter.OptChangeBody,
			milter.OptNoConnect | milter.OptNoHelo | milter.OptNoMailFrom | milter.OptNoRcptTo
	}
	// start server
	if err := milter.RunServer(socket, init); err != nil {
		log.Fatal(err)
	}
}

/* main program */
func main() {
	// parse commandline arguments
	var protocol, address string
	flag.StringVar(&protocol,
		"proto",
		"unix",
		"Protocol family (unix or tcp)")
	flag.StringVar(&address,
		"addr",
		"/var/spool/postfix/milters/ext.sock",
		"Bind to address or unix domain socket")
	flag.Parse()

	// make sure the specified protocol is either unix or tcp
	if protocol != "unix" && protocol != "tcp" {
		log.Fatal("invalid protocol name")
	}

	// make sure socket does not exist
	if protocol == "unix" {
		// ignore os.Remove errors
		os.Remove(address)
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// bind to listening address
	socket, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	if protocol == "unix" {
		// set mode 0660 for unix domain sockets
		if err := os.Chmod(address, 0660); err != nil {
			log.Fatal(err)
		}
		// remove socket on exit
		defer os.Remove(address)
	}
	log.Println("Starting server")
	// run server
	RunServer(socket)
}
