package httpfile

import (
	"bufio"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

const (
	PREMETHOD        = iota
	METHOD           = iota
	HEADER           = iota
	BODY             = iota
	RESPONSEFUNCTION = iota
)

type Parser struct {
	reqs    []http.Request
	req     HTTPFile
	scanner *bufio.Scanner
}

func newParser(r io.Reader) (p *Parser) {
	_p := new(Parser)
	_p.scanner = bufio.NewScanner(r)
	_p.req = NewHTTPFile()
	return _p
}

func removeComment(line string) string {
	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(line), "//") {
		return ""
	}
	s := strings.Split(line, " #")[0]
	s = strings.Split(line, " //")[0]
	return s
}

func trimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

// Everything before GET and POST Statements
func (p *Parser) parsePre(line string) int {
	//fmt.Println("Pre:" + line)

	if strings.HasPrefix(line, "// @Name ") {
		p.req.Name = strings.TrimSpace(line[8:])
		return PREMETHOD
	}

	if strings.HasPrefix(line, "// @Tags ") {
		p.req.Tags = strings.Split(strings.TrimSpace(line[8:]), ",")
		for idx := range p.req.Tags {
			p.req.Tags[idx] = strings.TrimSpace(p.req.Tags[idx])
		}
		return PREMETHOD
	}

	// this might from pevious request
	if strings.HasPrefix(strings.TrimSpace(line), "###") {
		return PREMETHOD
	}

	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		p.req.Comments = append(p.req.Comments, strings.TrimSpace(line))
		return PREMETHOD
	}
	if strings.HasPrefix(strings.TrimSpace(line), "//") {
		p.req.Comments = append(p.req.Comments, strings.TrimSpace(line))
		return PREMETHOD
	}

	line = removeComment(line)
	if len(line) == 0 {
		return PREMETHOD
	}

	if strings.HasPrefix(line, "GET") {
		return METHOD
	}
	if strings.HasPrefix(line, "POST") {
		return METHOD
	}
	if strings.HasPrefix(line, "OPTIONS") {
		return METHOD
	}

	return PREMETHOD
}

// The Full GET or POST Statement
func (p *Parser) parseMethod(line string) int {
	//fmt.Println("Method:" + line)

	if !strings.HasPrefix(line, "GET") && !strings.HasPrefix(line, "POST") && !strings.HasPrefix(line, "OPTIONS") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
		return HEADER
	}
	if strings.HasPrefix(line, "###") {
		return PREMETHOD
	}

	if strings.HasPrefix(line, "GET ") {
		p.req.Method = "GET"
		line = trimLeftChars(line, 4)
	}
	if strings.HasPrefix(line, "OPTIONS ") {
		p.req.Method = "OPTIONS"
		line = trimLeftChars(line, 8)
	}
	if strings.HasPrefix(line, "POST") {
		p.req.Method = "POST"
		line = trimLeftChars(line, 5)
		/*
			if !strings.HasPrefix(line, "http") {
				line = "http://" + line
			}
		*/
	}
	p.req.URL += strings.TrimSpace(line)

	return METHOD
}

// The Headers after the GET or POST Statement
func (p *Parser) parseHeader(line string) int {
	//fmt.Println("Header:" + line)
	if strings.HasPrefix(line, "###") {
		return PREMETHOD
	}

	if len(strings.TrimSpace(line)) == 0 {
		return BODY
	}

	line = removeComment(line)
	if len(line) == 0 {
		return HEADER
	}

	kv := strings.Split(line, ":")
	if len(kv) != 2 {
		return HEADER
	}
	h := HTTPHeader{
		Key:   strings.TrimSpace(kv[0]),
		Value: strings.TrimSpace(kv[1]),
	}
	p.req.Header = append(p.req.Header, h)

	return HEADER
}

func (p *Parser) parseBody(line string) int {
	//fmt.Println("Body:" + line)

	if strings.HasPrefix(line, "###") {
		return PREMETHOD
	}
	if strings.HasPrefix(line, "> {%") {
		return RESPONSEFUNCTION
	}
	if len(strings.TrimSpace(line)) == 0 {
		return BODY
	}

	p.req.Body += line + "\n"

	return BODY
}

func (p *Parser) parseResponseFunction(line string) int {
	//fmt.Println("Responsefunction:" + line)

	if strings.HasPrefix(line, "###") {
		return PREMETHOD
	}
	if len(strings.TrimSpace(line)) == 0 {
		return BODY
	}
	p.req.ResponseFunction += line + "\n"

	return RESPONSEFUNCTION
}

func (p *Parser) parsePart(part int, line string) int {
	switch part {
	case PREMETHOD:
		return p.parsePre(line)
	case METHOD:
		return p.parseMethod(line)
	case HEADER:
		return p.parseHeader(line)
	case BODY:
		return p.parseBody(line)
	case RESPONSEFUNCTION:
		return p.parseResponseFunction(line)
	default:
		return PREMETHOD
	}
}

// Parses Parameters from an URL
//
//	https://xxx.xxxx.xx/abcd/efgh?a=b&c=e
//
// into an array
// a = b, c = e
func fillParameters(request *HTTPFile) {
	urlSplit := strings.Split(request.URL, "?")
	if len(urlSplit) != 2 {
		return
	}
	request.URL = urlSplit[0]
	params := strings.Split(urlSplit[1], "&")
	for _, kv := range params {
		keyvalue := strings.Split(kv, "=")
		if len(keyvalue) < 2 {
			//panic("URL Parameters do not contain key value pairs")
			keyvalue = []string{"", kv}
		} else if len(keyvalue) > 2 {
			v := strings.Join(keyvalue[1:], "=")
			keyvalue = []string{keyvalue[0], v}
		}
		request.Parameter = append(request.Parameter, HTTPParameter{Key: keyvalue[0], Value: keyvalue[1]})
	}
}

func (p *Parser) parse(addKeepAlive bool) {
	part := PREMETHOD

	for p.scanner.Scan() {
		line := p.scanner.Text()
		//fmt.Println(scanner.Text())

		newpart := p.parsePart(part, line)
		if part != PREMETHOD && newpart == PREMETHOD {
			fillParameters(&p.req)
			p.reqs = append(p.reqs, *PrepareRequest(p.req, addKeepAlive))
			p.req = NewHTTPFile()
		}
		if newpart != part {
			newpart = p.parsePart(newpart, line)
		}

		part = newpart
	}
	if len(p.req.Method) != 0 {
		fillParameters(&p.req)
		p.reqs = append(p.reqs, *PrepareRequest(p.req, addKeepAlive))
		p.req = NewHTTPFile()
	}

	if err := p.scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func HTTPFileParser(templatePath fs.FS, path string, addKeepAlive bool) []http.Request {
	file, err := templatePath.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file fs.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	p := newParser(file)
	p.parse(addKeepAlive)

	return p.reqs
}
