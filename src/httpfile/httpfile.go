package httpfile

type HTTPHeader struct {
	Key   string
	Value string
}

type HTTPParameter struct {
	Key   string
	Value string
}

type HTTPFile struct {
	Name string

	Comments []string

	Method    string
	URL       string
	Parameter []HTTPParameter
	Header    []HTTPHeader
	Body      string

	ResponseFunction string

	Tags []string
}

func NewHTTPFile() HTTPFile {
	var request HTTPFile
	request.Parameter = make([]HTTPParameter, 0)
	request.Header = make([]HTTPHeader, 0)
	request.Comments = make([]string, 0)
	return request
}
