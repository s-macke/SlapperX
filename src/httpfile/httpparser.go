package httpfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var HTTPMethods = map[string]bool{
	"GET":     true,
	"HEAD":    true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"CONNECT": true,
	"OPTIONS": true,
	"TRACE":   true,
	"PATCH":   true,
}

// parserState represents the current state of the parser
type parserState int

const (
	StatePreMethod        parserState = iota
	StateMethod           parserState = iota
	StateHeader           parserState = iota
	StateBody             parserState = iota
	StateResponseFunction parserState = iota
)

type Parser struct {
	reqs    []http.Request
	req     HTTPFile
	content string
}

func newParser(r string) (p *Parser) {
	_p := new(Parser)
	_p.content = r
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
	s = strings.Split(s, " //")[0]
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

// Prüft, ob eine Zeile mit einer gültigen HTTP-Methode oder URL beginnt
func isValidMethodLine(line string) bool {
	if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		return true
	}
	for method := range HTTPMethods {
		if strings.HasPrefix(line, method+" ") || strings.HasPrefix(line, method+"\t") {
			return true
		}
	}
	return false
}

// Everything before GET and POST Statements
func (p *Parser) parsePre(line string) (parserState, error) {
	//fmt.Println("Pre:" + line)

	if strings.HasPrefix(line, "// @Name ") {
		p.req.Name = strings.TrimSpace(line[8:])
		return StatePreMethod, nil
	}

	if strings.HasPrefix(line, "// @Tags ") {
		p.req.Tags = strings.Split(strings.TrimSpace(line[8:]), ",")
		for idx := range p.req.Tags {
			p.req.Tags[idx] = strings.TrimSpace(p.req.Tags[idx])
		}
		return StatePreMethod, nil
	}

	// this might from pevious request
	if strings.HasPrefix(strings.TrimSpace(line), "###") {
		return StatePreMethod, nil
	}

	if strings.HasPrefix(strings.TrimSpace(line), "#") {
		p.req.Comments = append(p.req.Comments, strings.TrimSpace(line))
		return StatePreMethod, nil
	}
	if strings.HasPrefix(strings.TrimSpace(line), "//") {
		p.req.Comments = append(p.req.Comments, strings.TrimSpace(line))
		return StatePreMethod, nil
	}

	line = removeComment(line)
	if len(line) == 0 {
		return StatePreMethod, nil
	}

	if isValidMethodLine(line) {
		return StateMethod, nil
	}

	// If we have non-empty, non-comment content that's not a valid method line,
	// it's likely malformed content
	if len(strings.TrimSpace(line)) > 0 {
		return StatePreMethod, errors.New("unexpected content: line does not match expected format")
	}

	return StatePreMethod, nil
}

// The Full GET or POST Statement
func (p *Parser) parseMethod(line string) (parserState, error) {
	//fmt.Println("Method:" + line)

	if strings.HasPrefix(line, "###") {
		return StatePreMethod, nil
	}

	if !isValidMethodLine(line) {
		return StateHeader, nil
	}

	if strings.HasPrefix(line, "http") {
		p.req.Method = "GET"
	} else {
		for method := range HTTPMethods {
			methodWithSpace := method + " "
			methodWithTab := method + "\t"
			if strings.HasPrefix(line, methodWithSpace) {
				p.req.Method = method
				line = trimLeftChars(line, len(methodWithSpace))
				break
			} else if strings.HasPrefix(line, methodWithTab) {
				p.req.Method = method
				line = trimLeftChars(line, len(methodWithTab))
				break
			}
		}
	}

	// Validate that URL exists after method
	url := strings.TrimSpace(line)
	if len(url) == 0 {
		return StateMethod, errors.New("HTTP method must be followed by a URL")
	}

	p.req.URL += url

	return StateMethod, nil
}

// The Headers after the GET or POST Statement
func (p *Parser) parseHeader(line string) (parserState, error) {
	//fmt.Println("Header:" + line)
	if strings.HasPrefix(line, "###") {
		return StatePreMethod, nil
	}

	if len(strings.TrimSpace(line)) == 0 {
		return StateBody, nil
	}

	line = removeComment(line)
	if len(line) == 0 {
		return StateHeader, nil
	}

	kv := strings.Split(line, ":")
	if len(kv) != 2 {
		return StateHeader, errors.New("invalid header format: headers must contain a colon separator")
	}
	h := HTTPHeader{
		Key:   strings.TrimSpace(kv[0]),
		Value: strings.TrimSpace(kv[1]),
	}
	p.req.Header = append(p.req.Header, h)

	return StateHeader, nil
}

func (p *Parser) parseBody(line string) (parserState, error) {
	//fmt.Println("Body:" + line)

	if strings.HasPrefix(line, "###") {
		return StatePreMethod, nil
	}
	if strings.HasPrefix(line, "> {%") {
		return StateResponseFunction, nil
	}
	if len(strings.TrimSpace(line)) == 0 {
		return StateBody, nil
	}

	p.req.Body += line + "\n"

	return StateBody, nil
}

func (p *Parser) parseResponseFunction(line string) (parserState, error) {
	//fmt.Println("Responsefunction:" + line)

	if strings.HasPrefix(line, "###") {
		return StatePreMethod, nil
	}
	if len(strings.TrimSpace(line)) == 0 {
		return StateBody, nil
	}
	p.req.ResponseFunction += line + "\n"

	return StateResponseFunction, nil
}

func (p *Parser) parsePart(part parserState, line string) (parserState, error) {
	switch part {
	case StatePreMethod:
		return p.parsePre(line)
	case StateMethod:
		return p.parseMethod(line)
	case StateHeader:
		return p.parseHeader(line)
	case StateBody:
		return p.parseBody(line)
	case StateResponseFunction:
		return p.parseResponseFunction(line)
	default:
		return StatePreMethod, nil
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

func (p *Parser) parse(addKeepAlive bool) error {
	part := StatePreMethod

	for _, line := range strings.Split(strings.ReplaceAll(p.content, "\r\n", "\n"), "\n") {
		//fmt.Println(scanner.Text())

		newpart, err := p.parsePart(part, line)
		if err != nil {
			return err
		}
		if part != StatePreMethod && newpart == StatePreMethod {
			// Validate the request before adding it
			if len(p.req.Method) == 0 {
				return errors.New("incomplete request: no HTTP method found")
			}
			if len(p.req.URL) == 0 {
				return errors.New("incomplete request: no URL found")
			}
			fillParameters(&p.req)
			req, err := PrepareRequest(p.req, addKeepAlive)
			if err != nil {
				return err
			}
			p.reqs = append(p.reqs, *req)
			p.req = NewHTTPFile()
		}
		if newpart != part {
			newpart, err = p.parsePart(newpart, line)
			if err != nil {
				return err
			}
		}

		part = newpart
	}
	if len(p.req.Method) != 0 {
		fillParameters(&p.req)
		req, err := PrepareRequest(p.req, addKeepAlive)
		if err != nil {
			return err
		}
		p.reqs = append(p.reqs, *req)
		p.req = NewHTTPFile()
	}
	return nil
}

func HTTPFileParser(path string, overridesPath string, addKeepAlive bool) ([]http.Request, error) {
	httpFile, err := template.ParseGlob(path)
	if err != nil {
		return nil, errors.New("failed to parse HTTP template file: " + err.Error())
	}
	var overrides any = nil
	overridesFile, err := os.ReadFile(overridesPath)
	if err == nil {
		err := json.Unmarshal(overridesFile, &overrides)
		if err != nil {
			return nil, errors.New("failed to unmarshal JSON overrides: " + err.Error())
		}
	}
	var buff bytes.Buffer
	err = httpFile.Execute(&buff, overrides)
	if err != nil {
		return nil, errors.New("failed to execute template: " + err.Error())
	}

	p := newParser(buff.String())
	err = p.parse(addKeepAlive)
	if err != nil {
		return nil, errors.New("failed to parse HTTP content: " + err.Error())
	}

	return p.reqs, nil
}
