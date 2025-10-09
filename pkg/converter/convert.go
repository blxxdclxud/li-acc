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
	"time"
)

// Converter struct manages API operations of the converter.
// The methods are general for all kinds of conversion, except ExcelToPdf.
type Converter struct {
	token   string // auth token of API
	baseUrl string
	client  HTTPClient
}

// NewConverter is initializer of the Converter object, generates auth token and stores in field `token`
func NewConverter(baseUrl, publicKey, privateKey string) (*Converter, error) {
	c := Converter{
		baseUrl: strings.TrimRight(baseUrl, "/"),
		client:  RealClient{},
	}

	token, err := c.generateToken(publicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}
	c.token = token
	return &c, nil
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
	// taskId is the ID of created task, used in next requests
	taskId, err := c.createTask(conversionType)
	if err != nil {
		return errs.Wrap(errs.System, "failed to create task", err)
	}

	// fileKey is ID of uploaded file
	fileKey, err := c.uploadFile(taskId, inFilepath)
	if err != nil {
		return errs.Wrap(errs.System, "failed to upload file", err)
	}

	err = c.executeConversion(taskId)
	if err != nil {
		return errs.Wrap(errs.System, "failed to execute conversion", err)
	}

	downloadUrl, err := c.getConvertedFileUrl(fileKey)

	if err != nil {
		return errs.Wrap(errs.System, "failed to fetch url of converted file", err)
	}
	if downloadUrl == "" {
		return errs.Wrap(errs.System, "fetched empty downloadUrl, err is nil", nil)
	}

	// download file from the obtained link and store to [outFilepath]
	err = downloadFile(downloadUrl, outFilepath)
	if err != nil {
		return errs.Wrap(errs.System, "failed to download file", err)
	}
	return nil
}

// generateToken generates temporary auth token via API call. Token places in other requests' headers.
func (c *Converter) generateToken(publicKey, secretKey string) (string, error) {
	if publicKey == "" || secretKey == "" {
		return "", fmt.Errorf("missing API keys in environment variables")
	}

	content := map[string]string{
		"publicKey": publicKey,
		"secretKey": secretKey,
	}

	contentJson, _ := json.Marshal(content)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var tr TokenResponse

	err := c.doJSONRequest(http.MethodPost, EndpointToken, bytes.NewReader(contentJson), headers, nil, &tr)
	if err != nil {
		return "", err
	}

	return tr.Data.AccessToken, nil
}

// createTask creates the task via API call, when task is opened, we can attach the files and perform conversions.
func (c *Converter) createTask(conversion Conversion) (string, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + c.token,
	}

	var tr CreateTaskResponse
	err := c.doJSONRequest(http.MethodGet, conversion.CreateTaskEndpoint(), nil, headers, nil, &tr)
	if err != nil {
		return "", err
	}

	return tr.Data.TaskId, nil
}

// uploadFile uploads files to the server and attaches them to recently created task by [taskId]
func (c *Converter) uploadFile(taskId string, filepath string) (string, error) {
	// Open the file that will be uploaded to the converter API
	file, err := os.Open(filepath)
	if err != nil {
		return "", errs.WrapIOError("open file to upload it to converter API", filepath, err)
	}
	defer file.Close()

	// Prepare multipart/form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Attach the file under the form field "file"
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		return "", fmt.Errorf("create multipart form file: %w", err)
	}

	// Copy file contents into the form
	if _, err = io.Copy(part, file); err != nil {
		return "", err
	}

	// Add the task ID as a form field
	if err = writer.WriteField("taskId", taskId); err != nil {
		return "", err
	}

	// Finalize the multipart body before sending
	if err = writer.Close(); err != nil {
		return "", err
	}

	// Prepare request headers (add multipart content type)
	headers := map[string]string{
		"Authorization": "Bearer " + c.token,
		"Content-Type":  writer.FormDataContentType(),
	}

	var fr UploadFileResponse

	// Send the multipart request and decode JSON response into `fr`
	err = c.doJSONRequest(http.MethodPost, EndpointUploadFile, body, headers, nil, &fr)
	if err != nil {
		return "", err
	}

	// Return the file key from API response
	return fr.Data.FileKey, nil
}

// executeConversion converts uploaded Excel file to PDF
func (c *Converter) executeConversion(taskId string) error {
	params := map[string]string{
		"taskId": taskId,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + c.token,
	}

	err := c.doJSONRequest(http.MethodGet, EndpointConvert, nil, headers, params, nil)
	if err != nil {
		return err
	}

	return nil
}

// getConvertedFileUrl fetches converted file download URL from json response of the API
func (c *Converter) getConvertedFileUrl(fileKey string) (string, error) {

	var gr GetConvertedResponse

	// file conversion can be not finished, when fetching its dowloadUrl
	// so try to fetch it again, if downloadUrl=""
	for {
		gr = GetConvertedResponse{}
		params := map[string]string{
			"fileKey": fileKey,
		}
		headers := map[string]string{
			"Authorization": "Bearer " + c.token,
		}

		err := c.doJSONRequest(http.MethodGet, EndpointGetConverted, nil, headers, params, &gr)
		if err != nil {
			return "", err
		}

		// ensure that file indeed converted and download link is ready
		if gr.Data.FileUrl != "" {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}
	return gr.Data.FileUrl, nil
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
	fullURL, err := c.endpoint(endpoint)

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
