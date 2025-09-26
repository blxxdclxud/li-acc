package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var apiUrl = "https://api-server.compdf.com/server/v1/"
var (
	endpointToken        = "oauth/token"
	endpointUploadFile   = "file/upload"
	endpointConvert      = "execute/start"
	endpointGetConverted = "file/fileInfo"
)

// Conversion defines converter's conversion kinds.
// e.g. from DOC to PDF, etc.
type Conversion struct {
	From string
	To   string
}

func (c *Conversion) endpoint() string {
	return fmt.Sprintf("task/%s/%s", strings.ToLower(c.From), strings.ToLower(c.To))
}

// Converter struct manages API operations of the converter.
// The methods are general for all kinds of conversion, except ExcelToPdf.
type Converter struct {
	token   string // auth token of API
	baseUrl string
}

// NewConverter is initializer of the Converter object, generates auth token and stores in field `token`
func NewConverter() (*Converter, error) {
	c := Converter{}
	c.baseUrl = strings.TrimRight(apiUrl, "/")

	token, err := c.generateToken()
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

// newRequest is util function that creates http.Request object
// using given http method, endpoint path and request body.
func (c *Converter) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	// generate the complete API url
	fullURL, err := c.endpoint(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Convert converts given file to another format. Formats are given in [conversionType] argument.
// Source file's path is passed as [inFilepath], result file is stored as [outFilepath].
// Used here ComPDFKit API requires to perform following steps to convert the file:
// Create Task, Upload File, Execute the conversion of the uploaded file, Get the download link of the converted file.
func (c *Converter) Convert(inFilepath, outFilepath string, conversionType Conversion) error {
	// taskId is the ID of created task, used in next requests
	taskId, err := c.createTask(conversionType)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// fileKey is ID of uploaded file
	fileKey, err := c.uploadFile(taskId, inFilepath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	err = c.executeConversion(taskId)
	if err != nil {
		return fmt.Errorf("failed to execute conversion: %w", err)
	}

	downloadUrl, err := c.getConvertedFileUrl(fileKey)
	fmt.Println(downloadUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch url of converted file: %w", err)
	}

	// download file from the obtained link and store to [outFilepath]
	err = downloadFile(downloadUrl, outFilepath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return err
}

// generateToken generates temporary auth token via API call. It is used then passing in requests header.
func (c *Converter) generateToken() (string, error) {
	publicKey := os.Getenv("PUBLIC_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	if publicKey == "" || secretKey == "" {
		return "", fmt.Errorf("missing PUBLIC_KEY or SECRET_KEY in environment variables")
	}

	content := map[string]string{
		"publicKey": publicKey,
		"secretKey": secretKey,
	}

	contentJson, _ := json.Marshal(content)

	req, _ := c.newRequest(http.MethodPost, endpointToken, bytes.NewReader(contentJson))
	req.Header.Add("Content-Type", "application/json")

	resp, err := performRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}

	return tr.Data.AccessToken, nil
}

// createTask creates the task via API call, when task is opened, we can attach the files and perform conversions.
func (c *Converter) createTask(conversion Conversion) (string, error) {
	req, _ := c.newRequest(http.MethodGet, conversion.endpoint(), nil)
	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := performRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tr createTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	return tr.Data.TaskId, nil
}

// uploadFile uploads files to the server and attaches them to recently created task by [taskId]
func (c *Converter) uploadFile(taskId string, filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(part, file); err != nil {
		return "", err
	}
	err = writer.WriteField("taskId", taskId)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, _ := c.newRequest(http.MethodPost, endpointUploadFile, body)

	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := performRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var fr uploadFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return "", err
	}
	return fr.Data.FileKey, nil
}

// executeConversion converts uploaded Excel file to PDF
func (c *Converter) executeConversion(taskId string) error {
	req, _ := c.newRequest(http.MethodGet, endpointConvert, nil)

	params := req.URL.Query()
	params.Add("taskId", taskId)
	req.URL.RawQuery = params.Encode()

	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := performRequest(req)
	resp.Body.Close()

	return err
}

// getConvertedFileUrl fetches converted file download URL from json response of the API
func (c *Converter) getConvertedFileUrl(fileKey string) (string, error) {
	req, _ := c.newRequest(http.MethodGet, endpointGetConverted, nil)

	params := req.URL.Query()
	params.Add("fileKey", fileKey)
	req.URL.RawQuery = params.Encode()

	req.Header.Add("Authorization", "Bearer "+c.token)

	resp, err := performRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var gr getConvertedResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	return gr.Data.DownloadUrl, nil
}

// getJsonResponse fetches the Json format response of the server and returns it as map of interface{} objects.
// Passed parameter [body] - API response body. Returns error if API response is not successful.
func getJsonResponse(body io.Reader) (map[string]interface{}, error) {
	var data map[string]interface{}
	byteData, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(byteData, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return data, nil
}

// downloadFile gets file download URL as [url] parameter, gets its binary data and writes to the file [filename].
func downloadFile(url string, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("converted file downloading failed failed. URL: %s, Code: %d",
			url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, 0606)
	if err != nil {
		return fmt.Errorf("failed to open the file '%s': %w", filename, err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write pdf data to the file '%s': %w", filename, err)
	}

	return nil

}

// performRequest creates default HTTP Client and performs the request [req]. Returns error if it is unsuccessful,
// the response object otherwise.
func performRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request %s: %w", req.URL.String(), err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api request failed. URL: %s, Code: %d, response: %s",
			req.URL.String(), resp.StatusCode, string(body))
	}

	return resp, nil
}
