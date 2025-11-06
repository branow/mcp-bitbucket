package util

import (
	"net/http"
)

// LogArgsExtractor collects key-value pairs for structured logging.
type LogArgsExtractor struct {
	args map[string]any
}

func NewLogArgsExtractor() *LogArgsExtractor {
	return &LogArgsExtractor{args: make(map[string]any)}
}

func (e *LogArgsExtractor) AddRequest(req *http.Request) *LogArgsExtractor {
	e.AddUrl(req.URL.String())
	e.addArg("method", req.Method)
	e.addArg("utl_host", req.URL.Host)
	e.addArg("url_path", req.URL.Path)
	if q := req.URL.RawQuery; q != "" {
		e.addArg("url_query", q)
	}
	return e
}

func (e *LogArgsExtractor) AddResponse(resp *http.Response) *LogArgsExtractor {
	if resp.Request != nil {
		e.AddRequest(resp.Request)
	}
	e.addArg("status_code", resp.StatusCode)
	e.addArg("status_text", http.StatusText(resp.StatusCode))
	return e
}

func (e *LogArgsExtractor) AddError(err error) *LogArgsExtractor {
	e.addArg("error", err.Error())
	return e
}

func (e *LogArgsExtractor) AddUrlBuilder(urlBuilder UrlBuilder) *LogArgsExtractor {
	e.addArg("base_url", urlBuilder.BaseUrl)
	e.addArg("url_path", urlBuilder.Path)
	e.addArg("query_params", urlBuilder.QueryParams)
	return e
}

func (e *LogArgsExtractor) AddUrl(url string) *LogArgsExtractor {
	e.addArg("url", url)
	return e
}

func (e *LogArgsExtractor) Extract() []any {
	argsList := make([]any, 0, len(e.args)*2)
	for key, value := range e.args {
		argsList = append(argsList, key)
		argsList = append(argsList, value)
	}
	return argsList
}

func (e *LogArgsExtractor) addArg(key string, value any) {
	e.args[key] = value
}
