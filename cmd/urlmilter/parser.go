package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"

	"mvdan.cc/xurls/v2"
)

const prependText = `** WARNING - EXTERNAL MESSAGE **
This e-mail originated outside of Konrad Technologies.
Be careful when opening links or attachments, unless you recognize the sender and know the content is safe!
-----

`

var rtfHeader = regexp.MustCompile(`({\\rtf[0-9])\s*(\\ansicpg[0-9]+|\\ansi|\\mac|\\pc|\\pca|\\fbidis|\\fromtext|\\fromhtml[0-9]?|\\uc[0-9]|\\deff[0-9]+|\\adeff[0-9]+|\\stshfdbch[0-9]+|\\stshfloch[0-9]+|\\stshfhich[0-9]+|\\stshfbi[0-9]+|\\deflang[0-9]+|\\deflangfe[0-9]+|\\adeflang[0-9]+|({\\fonttbl\s*({?(\s*\\f.+?;\s*)+}?\s*)*})|({\\filetbl\s*({\\.+?;})+\s*})|({\\colortbl\s*.+?\s*;})|({\\stylesheet\s*({\s*\\.+?;\s*})+\s*})|({(\\\*)?(\\latentstyles|\\lsdstimax[0-9]+|\\lsdlockeddef[0-9]+|\\lsdsemihiddendef[0-9]+|\\lsdunhideuseddef[0-9]+|\\lsdqformatdef[0-9]+|\\lsdprioritydef[0-9]+)+.*?})|({(\\\*)?\\(listtable|listoverridetable)\s*.+?\s*({\s*\\.+?;?\s*})+\s*})|({(\\\*)?\\revtbl\s*({\s*\\.+?;?\s*})+\s*})|({(\\\*)?\\pgptbl\s*({\s*\\pgp.+?;?\s*})+\s*})|({(\\\*)?\\rsidtbl(\s*\\rsid[0-9]+\s*)+})|({(\\\*)?\\mmathPr(\s*\\m.+?\s*)+})|({(\\\*)?\\generator\s*.+?\s*;?})|({\\info\s*.+?\s*({\s*\\.+?;?\s*})+\s*})|({(\\\*)?\\userprops\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\xmlnstbl\s*.+?\s*({\s*\\xmlns.+?;?\s*})+\s*;?})|({(\\\*)?\\defchp\s*.+?\s*})|({(\\\*)?\\defpap\s*.+?\s*;?})|({(\\\*)?\\pgdscno[0-9]?})|\\noqfpromote|\\aenddoc|\\aendnotes|\\afelev|\\aftnbjaftncn|\\aftnnalc|\\aftnnar|\\aftnnauc|\\aftnnchi|\\aftnnchosung|\\aftnncnum|\\aftnndbar|\\aftnndbnum|\\aftnndbnumd|\\aftnndbnumk|\\aftnndbnumt|\\aftnnganada|\\aftnngbnum|\\aftnngbnumd|\\aftnngbnumk|\\aftnngbnuml|\\aftnnrlc|\\aftnnruc|\\aftnnzodiac|\\aftnnzodiacd|\\aftnnzodiacl|\\aftnrestart|\\aftnrstcont|\\aftnstart[0-9]+|\\aftntj|\\allowfieldendsel|\\allprot|\\alntblind|\\annotprot|\\ApplyBrkRules|\\asianbrkrule|\\autofmtoverride|\\bdbfhdr|\\bdrrlswsix|\\bookfold|\\bookfoldrev|\\bookfoldsheets[0-9]+|\\brdrart[0-9]+|\\brkfrm|\\cachedcolbal|\\cts[0-9]+|\\cvmme|\\defformat|\\deftab[0-9]+|\\deleted|\\dghorigin[0-9]+|\\dghshow[0-9]+|\\dghspace[0-9]+|\\dgmargin|\\dgsnap|\\dgvorigin[0-9]+|\\dgvshow[0-9]+|\\dgvspace[0-9]+|\\dntblnsbdb|\\dntblnsbdbwid|\\dntultrlspc|\\doctemp|\\doctype[0-9]+|\\donotembedlingdata[0-9]+|\\donotembedsysfont[0-9]+|\\donotshowcomments|\\donotshowinsdel|\\donotshowmarkup|\\donotshowprops|\\dontadjustlineheightintable|\\enddoc|\\endnotes|\\enforceprot[0-9]+|\\expshrtn|\\facingp|\\felnbrelev|\\fet[0-9]+|\\forceupgrade|\\formdisp|\\formprot|\\formshade|\\fracwidth|\\ftnbj|\\ftnlytwnine|\\ftnnalc|\\ftnnar|\\ftnnauc|\\ftnnchi|\\ftnnchosung|\\ftnncnum|\\ftnndbar|\\ftnndbnum|\\ftnndbnumd|\\ftnndbnumk|\\ftnndbnumt|\\ftnnganada|\\ftnngbnum|\\ftnngbnumd|\\ftnngbnumk|\\ftnngbnuml|\\ftnnrlc|\\ftnnruc|\\ftnnzodiac|\\ftnnzodiacd|\\ftnnzodiacl|\\ftnrestart|\\ftnrstcont|\\ftnrstpg|\\ftnstart[0-9]+|\\ftntj|\\grfdocevents[0-9]+|\\gutter[0-9]+|\\gutterprl|\\horzdoc|\\htmautsp|\\hwelev|\\hyphauto[0-1]?|\\hyphcaps[0-1]?|\\hyphconsec[0-9]+|\\hyphhotz[0-9]+|\\ignoremixedcontent[0-9]+|\\ilfomacatclnup[0-9]+|\\indrlsweleven|\\jcompress|\\jexpand|\\jsksu|\\krnprsnet|\\ksulang[0-9]+|\\landscape|\\linestart[0-9]+|\\linkstyles|\\lnbrkrule|\\lnongrid|\\ltrdoc|\\ltrsect|\\lytcalctblwd|\\lytexcttp|\\lytprtmet|\\lyttblrtgr|\\makebackup|\\margb[0-9]+|\\margl[0-9]+|\\margmirror|\\margr[0-9]+|\\margt[0-9]+|\\msmcap|\\muser|\\newtblstyruls|\\noafcnsttbl|\\nobrkwrptbl|\\nocolbal|\\nocompatoptions|\\nocxsptable|\\noextrasprl|\\nofeaturethrottle[0-9]+|\\nogrowautofit|\\noindnmbrts|\\nojkernpunct|\\nolead|\\nolnhtadjtbl|\\nospaceforul|\\notabind|\\notbrkcnstfrctbl|\\notcvasp|\\notvatxbx|\\nouicompat|\\noultrlspc|\\noxlattoyen|\\ogutter[0-9]+|\\oldas|\\oldlinewrap|\\otblrul|\\paperh[0-9]+|\\paperw[0-9]+|\\pgbrdrb|\\pgbrdrfoot|\\pgbrdrhead|\\pgbrdrl|\\pgbrdropt[0-9]+|\\pgbrdrr|\\pgbrdrsnap|\\pgbrdrt|\\pgnstart[0-9]+|\\prcolbl|\\printdata|\\protend|\\protlevel[0-9]+|\\protstart|\\psover|\\psz[0-9]+|\\readonlyrecommended|\\readprot|\\relyonvml[0-9]+|\\remdttm|\\rempersonalinfo|\\revbar[0-9]+|\\revised|\\revisions|\\revprop[0-9]+|\\revprot|\\rsidroot[0-9]+|\\rtldoc|\\rtlgutter|\\saveinvalidxml[0-9]+|\\saveprevpict|\\shidden|\\showplaceholdtext[0-9]+|\\showxmlerrors[0-9]+|\\shp|\\snaptogridincell|\\spltpgpar|\\splytwnine|\\spriority[0-9]+|\\sprsbsp|\\sprslnsp|\\sprsspbf|\\sprstsm|\\sprstsp|\\ssemihidden[0-9]+|\\stylelock|\\stylelockbackcomp|\\stylelockenforced|\\stylelockqfset|\\stylelocktheme|\\stylesortmethod[0-9]+|\\subfontbysize|\\swpbdr|\\themelangcs[0-9]+|\\themelangfe[0-9]+|\\themelang[0-9]+|\\toplinepunct|\\trackformatting[0-9]+|\\trackmoves[0-9]+|\\transmf|\\truncatefontheight|\\truncex|\\tsd[0-9]+|\\twoonone|\\useltbaln|\\usenormstyforlist|\\usexform|\\utinl|\\validatexml[0-9]+|\\vertdoc|\\viewbksp[0-9]+|\\viewkind[0-9]+|\\viewnobound|\\viewscale[0-9]+|\\viewzk[0-9]+|\\widowctrl|\\wpjst|\\wpsp|\\wptab|\\wraptrsp|\\wrppunct|({(\\\*)?\\aftncn\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\aftnsep\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\aftnsepc\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\ftncn\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\ftnsep\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\ftnsepc\s*.+?\s*({\s*\\.+?;?\s*})*\s*})|({(\\\*)?\\background.+?;?})|({(\\\*)?\\fchars.+?;?})|({(\\\*)?\\lchars.+?;?})|({(\\\*)?\\nextfile.+?;?})|({(\\\*)?\\private.+?;?})|({(\\\*)?\\template\s*(\S|(\\{)|(\\}))*})|({(\\\*)?\\wgrffmtfilter\s*[0-9a-fA-F]{4}})|({(\\\*)?\\windowcaption.+?;?})|({(\\\*)?\\writereservation.+?;?})|({(\\\*)?\\writereservhash.+?;?})|({(\\\*)?\\xform.+?;?})|\s+)*`)

const insertRtfPar = `
{\pard \ql \b \fs32 WARNING - External Message\par}
{\pard \ql This e-mail originated outside of Konrad Technologies.\par}
{\pard \ql Be careful when opening links or attachments, unless you recognize the sender and know the content is safe!\par}
{\pard \qc \emdash\emdash\emdash\emdash\emdash\par}
`

const prependHtml = `<!DOCTYPE html>
<html><head></head>
<body>
<h3>WARNING - External Message</h3>
<p>This e-mail originated outside of Konrad Technologies.<br>Be careful when opening links or attachments, unless you recognize the sender and know the content is safe!</p>
<hr/>
</body>
</html>
`

var urlMatcher = xurls.Strict()

// ParseMessage processes an email message parts
func RewriteEmailMessage(r io.Reader) ([]byte, error) {

	// get message from input stream
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}
	// get media type from email message
	//media, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))

	contentType, _, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Error handling content-type - assuming text/plain for:", msg.Header.Get("Content-Type"))
		//return nil, err
	}

	// accept messages without attachments

	//if !strings.HasPrefix(media, "multipart/") {
	if contentType == "text/plain" || contentType == "text/html" || contentType == "" || contentType == "application/rtf" || contentType == "text/rtf" {
		log.Println("No multipart, processing body")
		return processPartBody(msg.Body, textproto.MIMEHeader(msg.Header))
	}

	if strings.HasPrefix(contentType, "multipart/") {
		log.Println("Starting to process type ", contentType)
		return processMultipart(msg.Body, textproto.MIMEHeader(msg.Header))
	}

	// accept message by default
	return nil, fmt.Errorf("Unimplemented")
}

// Handle Content-Transfer-Encoding according to RFC 2045
// If a MIME part is 7bit, the Content-Transfer-Encoding header is optional.
// MIME parts with any other transfer encoding must contain a Content-Transfer-Encoding header.
// If the MIME part is a multipart content type, the part should not have an encoding of base64 or quoted-printable.

func decodeTransfer(encoded []byte, contentTransferEncoding string) ([]byte, error) {
	var decoded []byte
	var err error

	switch contentTransferEncoding {
	case "base64":
		decoded = make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
		decodedLength, err := base64.StdEncoding.Decode(decoded, encoded)
		if err != nil {
			return nil, fmt.Errorf("error decoding base64: %w", err)
		}
		decoded = decoded[:decodedLength]
	case "quoted-printable":
		decoded, err = io.ReadAll(quotedprintable.NewReader(bytes.NewReader(encoded)))
		if err != nil {
			return nil, fmt.Errorf("error decoding quoted-printable: %w", err)
		}
	case "", "7bit", "8bit", "binary":
		// No decoding?
		decoded = encoded
	default:
		log.Println("Unhandled transfer encoding: " + contentTransferEncoding)
		return nil, errors.New("unhandled transfer encoding " + contentTransferEncoding)
	}

	return decoded, nil
}

func encodeTransfer(plain []byte, contentTransferEncoding string) ([]byte, error) {
	var encoded []byte

	switch contentTransferEncoding {
	case "base64":
		var buf = new(bytes.Buffer)
		writer := NewLineSplitter(76, []byte("\r\n"), buf)

		encoded = make([]byte, base64.StdEncoding.EncodedLen(len(plain)))
		base64.StdEncoding.Encode(encoded, plain)

		_, err := writer.Write(encoded)
		if err != nil {
			return nil, err
		}

		encoded = buf.Bytes()

	case "quoted-printable":
		var buf = new(bytes.Buffer)
		writer := quotedprintable.NewWriter(buf)
		written, err := writer.Write(plain)
		if written != len(plain) || err != nil {
			return nil, fmt.Errorf("Short write or error encoding quoted-printable (%w)", err)
		}
		err = writer.Close()
		if err != nil {
			return nil, err
		}
		encoded = buf.Bytes()
	case "", "7bit", "8bit", "binary":
		// No decoding?
		encoded = plain
	default:
		log.Println("Unhandled transfer encoding: " + contentTransferEncoding)
		return nil, errors.New("unhandled transfer encoding " + contentTransferEncoding)
	}

	return encoded, nil
}

func processPartBody(reader io.Reader, header textproto.MIMEHeader) ([]byte, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	encoding := header.Get("Content-Transfer-Encoding")

	// Message is currently single part, apply transformations to body and return
	decodedBody, err := decodeTransfer(body, encoding)
	if err != nil {
		return nil, fmt.Errorf("error decoding transfer: %w", err)
	}

	newBody := xurls.Strict().ReplaceAllFunc(decodedBody, func(match []byte) []byte {

		str := string(match)

		// TODO allow mailto?

		// ignore inline attachment references
		if strings.HasPrefix(str, "cid:") {
			return match
		}

		if str == "https://demvreply.datevnet.de/web.app?op=init" {
			// TODO convert allowlist to array
			return match
		}

		if url, err := url.Parse(str); err == nil {
			if strings.HasSuffix(url.Host, "konrad-technologies.de") || strings.HasSuffix(url.Host, "konrad-technologies.com") {
				return match
			}
		}

		return []byte("https://urlcheck.konrad-technologies.de/check/" + base64.RawURLEncoding.EncodeToString(match))
	})

	// get media type from part
	// TODO redundant parsing?
	contentType, _, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		log.Println("Error handling content-type - assuming text/plain:", header.Get("Content-Type"))
		//return nil, err
	}

	switch contentType {
	case "text/plain", "":
		// Assume empty or missing content types to be plain text
		newBody = append([]byte(prependText), newBody...)
	case "text/html":
		newBody = append([]byte(prependHtml), newBody...)
	case "application/rtf", "text/rtf":
		newBody = rtfHeader.ReplaceAllFunc(newBody, func(match []byte) []byte {
			insertRtfParBytes := []byte(insertRtfPar)
			cp := make([]byte, len(match)+len(insertRtfParBytes))
			copy(cp, match)
			return append(cp, insertRtfParBytes...)
		})
	}

	return encodeTransfer(newBody, encoding)
}

func processMultipart(input io.Reader, header textproto.MIMEHeader) ([]byte, error) {

	_, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		log.Println("Error handling content-type:", header.Get("Content-Type"))
		return nil, err
	}

	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)

	err = mw.SetBoundary(params["boundary"])
	if err != nil {
		return nil, err
	}

	// deep inspect multipart messages
	mr := multipart.NewReader(input, params["boundary"])
	for {
		// examine next message part
		part, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		newPart, err := mw.CreatePart(part.Header)
		if err != nil {
			return nil, err
		}

		// get media type from part
		contentType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			log.Println("Error handling content-type:", part.Header.Get("Content-Type"))
			return nil, err
		}

		if strings.HasPrefix(contentType, "multipart/") {
			// TODO limit recursion depth!
			newPartBody, err := processMultipart(part, part.Header)
			if err != nil {
				return nil, err
			}
			_, err = newPart.Write(newPartBody)
			if err != nil {
				return nil, fmt.Errorf("short write or error writing to mime part: %w", err)
			}
		}

		// only process text/plain and text/html message parts
		if contentType == "text/plain" || contentType == "text/html" || contentType == "" || contentType == "application/rtf" || contentType == "text/rtf" {
			newPartBody, err := processPartBody(part, part.Header)
			if err != nil {
				return nil, err
			}
			_, err = newPart.Write(newPartBody)
			if err != nil {
				return nil, fmt.Errorf("short write or error writing to mime part: %w", err)
			}

			//log.Printf("Rewritten part of type %v to %s", contentType, newPartBody)

		}

		_, err = io.Copy(newPart, part)
		if err != nil {
			return nil, err
		}
	}

	err = mw.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}
