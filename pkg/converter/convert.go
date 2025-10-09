package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"li-acc/internal/errs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Converter struct manages API operations of the converter.
// The methods are general for all kinds of conversion, except ExcelToPdf.
type Converter struct {
	apiKey  string // auth token of API
	baseUrl string
	client  HTTPClient
}

// NewConverter is initializer of the Converter object, generates auth token and stores in field `token`
func NewConverter(baseUrl, publicKey string) *Converter {
	c := Converter{
		baseUrl: strings.TrimRight(baseUrl, "/"),
		client:  RealClient{},
		apiKey:  publicKey,
	}
	return &c
}

// endpoint returns the complete API url for given path, using the base Converter.baseUrl
func (c *Converter) endpoint(path string) (string, error) {
	return url.JoinPath(c.baseUrl, path)
}

// Convert converts given file to another format. Formats are given in [conversionType] argument.
// Source file's path is passed as [inFilepath], result file is stored as [outFilepath].
// Used here ComPDFKit API requires to perform following steps to convert the file:
// Create Task, Upload File, Execute the conversion of the uploaded file, Get the download link of the converted file.
func (c *Converter) Convert(inFilepath, outFilepath string, conversionType Conversion) error {
	fileInfo, err := c.processConversion(inFilepath, conversionType)
	if err != nil {
		return errs.Wrap(errs.System, "failed to upload file", err)
	}

	fileUrl, err := getFileUrl(fileInfo)
	if err != nil {
		return errs.Wrap(errs.System, "failed to get file download url", err)
	}

	// download file from the obtained link and store to [outFilepath]
	err = downloadFile(fileUrl, outFilepath)
	if err != nil {
		return errs.Wrap(errs.System, "failed to download file", err)
	}
	return nil
}

// processConversion perform file conversion using API. It uploads file to the server, sets api key in headers
// and performs request.
func (c *Converter) processConversion(filepath string, conversion Conversion) (ProcessConversionResponse, error) {
	body, formDataType, err := buildMultipartBody(filepath)
	if err != nil {
		return ProcessConversionResponse{}, err
	}

	// Prepare request headers (add multipart content type)
	headers := map[string]string{
		"x-api-key":    c.apiKey,
		"Content-Type": formDataType,
	}

	var resp ProcessConversionResponse

	// Send the multipart request and decode JSON response into `resp`
	err = c.doJSONRequest(http.MethodPost, conversion.ConversionFormatEndpoint(), body, headers, nil, &resp)
	if err != nil {
		return ProcessConversionResponse{}, err
	}

	// Return the API response model
	return resp, nil
}

// buildMultipartBody creates a multipart/form-data request body for file upload.
// The function opens the file at the specified `filepath`, reads its contents, and forms
// a multipart body with a "file" field.
// Returns the finished request body, the corresponding 'Content-Type' header and an error if the operation fails.
func buildMultipartBody(filepath string) (io.Reader, string, error) {
	// Open the file that will be uploaded to the converter API
	file, err := os.Open(filepath)
	if err != nil {
		return nil, "", errs.WrapIOError("open file to upload it to converter API", filepath, err)
	}
	defer file.Close()

	// Prepare multipart/form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Attach the file under the form field "file"
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		return nil, "", fmt.Errorf("create multipart form file: %w", err)
	}

	// Copy file contents into the form
	if _, err = io.Copy(part, file); err != nil {
		return nil, "", fmt.Errorf("copy file contents into the multipart form: %w", err)
	}

	// Finalize the multipart body before sending
	if err = writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func getFileUrl(fileInfo ProcessConversionResponse) (string, error) {
	fileUrl := fileInfo.Data.FileInfo[0].DownloadUrl
	status := fileInfo.Data.FileInfo[0].Status

	if status != "success" {
		return "", fmt.Errorf("file conversion status is not 'success'")
	} else if fileUrl == "" {
		return "", fmt.Errorf("file download url from API response is empty")
	}

	return fileUrl, nil
}

// downloadFile gets file download URL as [url] parameter, gets its binary data and writes to the file [filename].
func downloadFile(url string, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot dowload file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("converted file downloading failed failed. URL: %s, Code: %d",
			url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open the file '%s': %w", filename, err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write file data to the file '%s': %w", filename, err)
	}

	return nil

}

// doJSONRequest is a helper function for methods of Converter.
// It forms the request with headers, performs the API call, checks response status for success and
// decodes json response into `out` (expected that it is XxxxResponse structure from models.go)
// `method`: HTTP method;
// `endpoint`: an endpoint for the current API call (without base URL);
// `body`: a body of the request;
// `headers`: a map of HTTP request headers;
// `params`: map of query params.
func (c *Converter) doJSONRequest(
	method, endpoint string,
	body io.Reader,
	headers map[string]string,
	params map[string]string,
	out any,
) error {

	// generate the complete API url
	fullURL, _ := c.endpoint(endpoint)

	// Build the base URL
	reqURL, err := url.Parse(fullURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Attach query parameters if provided
	if len(params) > 0 {
		q := reqURL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		reqURL.RawQuery = q.Encode()
	}

	// Build the HTTP request
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Send and process response
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if err := checkResponseStatus(resp, fullURL); err != nil {
		return err
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
