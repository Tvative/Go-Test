package gotest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"
)

type ApiTestResult struct {
	TestStatus      bool        // TestStatus is the status of the test case.
	TestDescription string      // TestDescription is the description of the test case.
	TestError       interface{} // TestError is the error of the test case, if available.
	TestTime        interface{} // TestTime is the time of the test case.
}

type ApiTest struct {
	Tests       int64                   // Tests is the count fo total test cases.
	PassedTests int64                   // PassedTests is the count of passed test cases.
	FailedTests int64                   // FailedTests is the count of failed test cases.
	Result      map[int64]ApiTestResult // Result is the result of the test cases.
	Server      *httptest.Server        // Server is the server for the test cases.
	ServerMux   *http.ServeMux          // ServerMux is the mux for the server.
}

type ApiTestRequest struct {
	Details        string      // Details is the details like case of the API call.
	ReqParam       interface{} // ReqParam is the path parameters of the API call.
	ReqBody        interface{} // ReqBody is the body parameters of the API call.
	ApiUrl         string      // ApiUrl is the endpoint URL of the API call.
	ApiMethod      string      // ApiMethod is the method of the API call.
	ContentType    interface{} // ContentType is the content type of the API call.
	BearerToken    interface{} // BearerToken is the bearer token (like JWT token) of the API call.
	ExpectedStatus interface{} // ExpectedStatus is the expected status code of the response.
}

var (
	ContentTypeJson  = "application/json"                  // ContentTypeJson is for APIs with Json content.
	ContentTypeXml   = "application/xml"                   // ContentTypeXml is for APIs with Xml content.
	ContentTypeForm  = "application/x-www-form-urlencoded" // ContentTypeForm is for APIs with Form content.
	ContentTypeText  = "text/plain"                        // ContentTypeText is for APIs with Text content.
	ContentTypeHtml  = "text/html"                         // ContentTypeHtml is for APIs with Html content.
	ContentTypePdf   = "application/pdf"                   // ContentTypePdf is for APIs with Pdf content.
	ContentTypeZip   = "application/zip"                   // ContentTypeZip is for APIs with Zip content.
	ContentTypePng   = "image/png"                         // ContentTypePng is for APIs with Png content.
	ContentTypeJpg   = "image/jpeg"                        // ContentTypeJpg is for APIs with Jpg content.
	ContentTypeGif   = "image/gif"                         // ContentTypeGif is for APIs with Gif content.
	ContentTypeSvg   = "image/svg+xml"                     // ContentTypeSvg is for APIs with Svg content.
	ContentTypeBmp   = "image/bmp"                         // ContentTypeBmp is for APIs with Bmp content.
	ContentTypeTiff  = "image/tiff"                        // ContentTypeTiff is for APIs with Tiff content.
	ContentTypePpt   = "application/vnd.ms-powerpoint"     // ContentTypePpt is for APIs with Ppt content.
	ContentTypeDoc   = "application/msword"                // ContentTypeDoc is for APIs with Doc content.
	ContentTypeXls   = "application/vnd.ms-excel"          // ContentTypeXls is for APIs with Xls content.
	ContentTypeCsv   = "text/csv"                          // ContentTypeCsv is for APIs with Csv content.
	ContentTypeXml2  = "application/xml; charset=utf-8"    // ContentTypeXml2 is for APIs with Xml2 content.
	ContentTypeHtml2 = "text/html; charset=utf-8"          // ContentTypeHtml2 is for APIs with Html2 content.
)

// InitApiTest function initializes an instance of the ApiTest struct and returns a pointer to it.
//
// Example usage:
//
// ```
// var T *ApiTest
// T = InitApiTest()
// // Initialize main route
// // More process...
// T.DumpApiTestResult(true)
// ```
func InitApiTest() *ApiTest {
	mux := http.NewServeMux()
	httptest.NewServer(mux)

	return &ApiTest{
		Tests:       0,
		PassedTests: 0,
		FailedTests: 0,
		Result:      make(map[int64]ApiTestResult),
		Server:      httptest.NewServer(mux),
		ServerMux:   mux,
	}
}

// generateApiUrl function takes a httptest.Server and a getPath string as inputs and returns a string that
// represents the complete URL for an API call.
func generateApiUrl(server *httptest.Server, getPath string) string {
	return server.URL + getPath
}

// addTestResult function adds a test result to the ApiTest struct.
func (h *ApiTest) addTestResult(description string, reqError interface{}, isTestPassed bool, processTime interface{}) {
	h.Tests++

	if isTestPassed {
		h.PassedTests++
	} else {
		h.FailedTests++
	}

	h.Result[h.Tests] = ApiTestResult{
		TestStatus:      isTestPassed,
		TestDescription: description,
		TestError:       reqError,
		TestTime:        processTime.(time.Duration),
	}
}

// CreateTest function creates a new test case for an API call.
func (h *ApiTest) CreateTest(httpReq ApiTestRequest) {
	var reqParam string
	var reqBody io.Reader

	if httpReq.ReqParam != nil {
		reqParam = httpReq.ReqParam.(string)
	} else {
		reqParam = ""
	}

	if httpReq.ReqBody != nil {
		jsonBytes, err := json.Marshal(httpReq.ReqBody)
		if err != nil {
			h.addTestResult(httpReq.Details, err.Error(), false, nil)
			return
		}

		reqBody = bytes.NewBufferString(string(jsonBytes))
	}

	fmt.Println(httpReq.ApiUrl + reqParam)
	req, err := http.NewRequest(httpReq.ApiMethod, generateApiUrl(h.Server, httpReq.ApiUrl)+reqParam, reqBody)
	if err != nil {
		h.addTestResult(httpReq.Details, err.Error(), false, nil)
		return
	}

	if httpReq.ContentType != nil {
		req.Header.Set("Content-Type", httpReq.ContentType.(string))
	}

	if httpReq.BearerToken != nil {
		req.Header.Set("Authorization", "Bearer "+httpReq.BearerToken.(string))
	}

	startTime := time.Now()
	resp, respErr := http.DefaultClient.Do(req)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	if respErr != nil {
		h.addTestResult(httpReq.Details, respErr.Error(), false, duration)
		return
	}

	if resp.StatusCode != httpReq.ExpectedStatus.(int) {
		h.addTestResult(httpReq.Details, resp, false, duration)
		return
	}

	h.addTestResult(httpReq.Details, nil, true, duration)
	return
}

// DumpApiTestResult function prints the result of the API test cases in to the terminal.
func (h *ApiTest) DumpApiTestResult(needExit bool) {
	defer h.Server.Close()

	fmt.Printf("\nAPI Test Result:\n\n")
	fmt.Printf("┌──────┬──────────┬─────────────────┬─────────────────────--------------►\n")
	fmt.Printf("│ %-4s │ %-8s │ %-15s │ %s\n", "No", "Status", "Time", "Description")
	fmt.Printf("├──────┼──────────┼─────────────────┼─────────────────────--------------►\n")

	for i, result := range h.Result {
		fmt.Printf("│ %-4d │ %-8s │ %-15s │ %s", i, strconv.FormatBool(result.TestStatus),
			result.TestTime, result.TestDescription)

		if result.TestError != nil {
			fmt.Print("\u001B[1;31m [ Error:\033[0;0m ", result.TestError, "\u001B[1;31m ]\u001B[0;0m")
		}

		fmt.Printf("\n")
	}

	fmt.Printf("└──────┴──────────┴─────────────────┴─────────────────────--------------►\n")

	fmt.Printf("\n%-40s : \033[1;36m%d\033[0;0m\n", "Total white box API test cases", h.Tests)
	fmt.Printf("%-40s : \033[1;32m%d/%d\033[0;0m\n", "Total passed white box API test cases", h.PassedTests, h.Tests)
	fmt.Printf("%-40s : \033[1;31m%d/%d\033[0;0m\n\n", "Total failed white box API test cases", h.FailedTests, h.Tests)

	if needExit {
		if h.FailedTests > 0 {
			os.Exit(1)
		}

		os.Exit(0)
	}
}
