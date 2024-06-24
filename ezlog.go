package ezlog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

type sourceFile struct {
	fullPath string
	source   map[int]string
}

var sourceFiles = make(map[string]*sourceFile)

type EventDetail struct {
	Level int    `json:"level,omitempty"`
	Line  int    `json:"line,omitempty"`
	Msg   string `json:"msg,omitempty"`
	File  string `json:"file,omitempty"`
	Func  string `json:"func,omitempty"`
}

type EventTraceSource struct {
	Line   int    `json:"line,omitempty"`
	Source string `json:"source,omitempty"`
}

type EventTrace struct {
	File    string             `json:"file,omitempty"`
	Func    string             `json:"func,omitempty"`
	Line    int                `json:"line,omitempty"`
	Sources []EventTraceSource `json:"sources,omitempty"`
}

type EventRuntime struct {
	Platform int    `json:"platform,omitempty"`
	Ver      string `json:"ver,omitempty"`
}

type EventSDK struct {
	Ver string `json:"ver,omitempty"`
}

type EventCreateRequest struct {
	Detail  EventDetail             `json:"detail,omitempty"`
	Traces  []EventTrace            `json:"traces,omitempty"`
	Runtime EventRuntime            `json:"runtime,omitempty"`
	SDK     EventSDK                `json:"sdk,omitempty"`
	Tags    *map[string]interface{} `json:"tags,omitempty"`
}

func Exception(err error, messages ...interface{}) {
	if clientOptions.ServiceKey == "" {
		return
	}

	request := EventCreateRequest{
		Traces: []EventTrace{},
		SDK: EventSDK{
			Ver: SDK_VERSION,
		},
	}

	// messages を解析
	for _, message := range messages {
		if req, ok := message.(*http.Request); ok {
			tags := map[string]interface{}{}

			scheme := "http"
			if req.TLS != nil {
				scheme = "https"
			}
			tags["requestURL"] = fmt.Sprintf("%s://%s%s", scheme, req.Host, req.RequestURI)
			tags["requestMethod"] = req.Method
			tags["userAgent"] = req.UserAgent()

			request.Tags = &tags
		}
	}

	for i := 2; i <= 6; i++ {
		field := GetField(i)

		if field.File == "" {
			break
		}

		var source *map[int]string

		if _, ok := sourceFiles[field.File]; ok {
			source = &sourceFiles[field.File].source
		} else {
			_source, err := readFileWithLine(field.File)
			if err != nil {
				fmt.Println(err)
				continue
			}

			sourceFiles[field.File] = &sourceFile{
				fullPath: field.File,
				source:   *_source,
			}
			source = _source
		}

		if i == 2 {
			request.Detail.Level = 31 // error
			request.Detail.Msg = err.Error()
			request.Detail.File = field.File
			request.Detail.Line = field.Line
			request.Detail.Func = field.Func
		}

		trace := EventTrace{
			File: field.File,
			Func: field.Func,
			Line: field.Line,
		}

		startLine := field.Line - 5
		if startLine < 1 {
			startLine = 1
		}
		endLine := field.Line + 5
		if endLine > len(*source) {
			endLine = len(*source)
		}

		for i := startLine; i <= endLine; i++ {
			traceSource := EventTraceSource{
				Line:   i,
				Source: (*source)[i],
			}

			trace.Sources = append(trace.Sources, traceSource)
		}

		request.Traces = append(request.Traces, trace)
	}

	request.Runtime.Platform = 1101         // go
	request.Runtime.Ver = runtime.Version() // "go1.16.3"
	sendEvents(&request)
}

func readFileWithLine(filePath string) (*map[int]string, error) {
	source := map[int]string{}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	line := 1
	for scanner.Scan() {
		source[line] = scanner.Text()
		line++
	}

	return &source, nil
}

// sendEvents イベントを送信
func sendEvents(createRequest *EventCreateRequest) error {
	endpoint := clientOptions.Endpoint + "/events/"

	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return err
	}
	urlStr := fmt.Sprintf("%v", u)

	byteArr, err := json.Marshal(createRequest)
	if err != nil {
		return err
	}

	collectRequest := url.Values{}
	collectRequest.Set("data", string(byteArr))

	// リクエストデータをgzip圧縮
	compressedData, err := GzipCompress(PointerConvert([]byte(collectRequest.Encode())))
	if err != nil {
		fmt.Println(err)
		return err
	}

	client := &http.Client{}
	request, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(*compressedData))
	if err != nil {
		fmt.Println(err)
		return err
	}

	request.Header.Add("X-EZLOG-SERVICE-KEY", clientOptions.ServiceKey)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Encoding", "gzip")

	resp, err := client.Do(request)

	// TODO: ネットワークエラーが発生した場合はリトライ
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer resp.Body.Close()
	contents, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println(fmt.Errorf("failed to post message: %s", contents))
		return err
	}

	return nil
}
