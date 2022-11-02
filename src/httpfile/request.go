package httpfile

import (
	"encoding/base64"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Transforms request
//
//	Authentification: Basic abcd efgh
//
// to
//
//	Authentification: Basic Base64(abcd:efgh)
func AuthToBase64(header HTTPHeader) HTTPHeader {
	if header.Key != "Authorization" {
		return header
	}
	if !strings.HasPrefix(header.Value, "Basic") {
		return header
	}
	header.Value = strings.TrimPrefix(header.Value, "Basic")
	header.Value = strings.TrimSpace(header.Value)
	//replace multiple spaces
	space := regexp.MustCompile(`\s+`)
	header.Value = space.ReplaceAllString(header.Value, " ")
	userPass := strings.Split(header.Value, " ")

	if len(userPass) != 2 { // Probably already base 64 encoded
		header.Value = "Basic " + userPass[0]
		return header
	}
	header.Value = "Basic " + base64.StdEncoding.EncodeToString([]byte(userPass[0]+":"+userPass[1]))
	return header
}

func PrepareRequest(r HTTPFile, addKeepAlive bool) *http.Request {
	req, err := http.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		log.Fatal(err)
	}
	q := req.URL.Query()
	for _, p := range r.Parameter {
		q.Add(p.Key, p.Value)
	}
	req.URL.RawQuery = q.Encode()

	for _, h := range r.Header {
		h = AuthToBase64(h)
		if h.Key == "Host" {
			req.Host = h.Value
		}
		req.Header.Set(h.Key, h.Value)
	}
	if addKeepAlive {
		req.Header.Set("Connection", "keep-alive")
	}
	/*
	   bytes2, err := httputil.DumpRequest(req, true)
	   fmt.Println(string(bytes2))
	*/
	return req
}

/*
func DoRequest(r HTTPRequest, showdetails bool) string {
	substituteVariables(&r, r.Variables)
	client := &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: redirectPolicy,
	}
	req, _ := http.NewRequest(r.Request.Method, r.Request.URL, strings.NewReader(r.Request.Body))

	q := req.URL.Query()
	for _, p := range r.Request.Parameter {
		q.Add(p.Key, p.Value)
	}
	req.URL.RawQuery = q.Encode()

	for _, h := range r.Request.Header {
		h = AuthToBase64(h)
		req.Header.Set(h.Key, h.Value)
	}
	/*
		bytes2, err := httputil.DumpRequest(req, true)
		fmt.Println(string(bytes2))
*/
/*
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}

	if showdetails {
		bytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bytes))
	}
	defer resp.Body.Close()
	return resp.Status
	//io.Copy(os.Stdout, resp.Body)

}
*/
