// Package vfssrv provides a client for interacting with a Virtual File System (VFS) service.
// It supports authentication, file uploads, and file downloads with proper error handling.
package vfssrv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmkteam/appkit"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("invalid token")
	ErrInternal     = errors.New("internal error")
)

const (
	uploadFormName = "Filedata"
	authHeader     = "AuthorizationJWT"
	defaultTimeout = 5 * time.Second

	// API endpoints
	authTokenURL  = "/auth-token"
	hashUploadURL = "/upload/hash"
)

// RequestError records an error, URL and code.
type RequestError struct {
	Code int
	URL  string
	Err  error
}

func (e *RequestError) Error() string {
	return strconv.Itoa(e.Code) + " " + e.URL + ": " + e.Err.Error()
}

func (e *RequestError) Unwrap() error { return e.Err }

func newRequestError(code int, url string, err error) *RequestError {
	return &RequestError{Code: code, URL: url, Err: err}
}

type Opts struct {
	ApiURL         string        // Base URL for API endpoints
	PublicURL      string        // Base URL for Public Files
	Timeout        time.Duration // HTTP request timeout, default is 5s
	UploadFormName string        // Form field name for file uploads, default is Filedata
	AuthHeader     string        // Header name for authentication, default is AuthorizationJWT
	Client         *http.Client  // Custom HTTP client (optional)
}

type Client struct {
	opts Opts
}

func NewClient(opts Opts) *Client {
	if opts.Timeout == 0 {
		opts.Timeout = defaultTimeout
	}

	if opts.PublicURL == "" {
		opts.PublicURL = opts.ApiURL
	}

	if opts.UploadFormName == "" {
		opts.UploadFormName = uploadFormName
	}

	if opts.AuthHeader == "" {
		opts.AuthHeader = authHeader
	}

	if opts.Client == nil {
		opts.Client = &http.Client{
			Timeout: opts.Timeout,
		}
	}

	return &Client{
		opts: opts,
	}
}

// apiURL returns vsfHost + authTokenURL.
func (c *Client) apiURL(basePath string) string {
	p, _ := url.JoinPath(c.opts.ApiURL, basePath)
	return p
}

// setHeaders adds additional passed headers and X-Request-Id from context.
func (c *Client) setHeaders(ctx context.Context, req *http.Request) {
	appkit.SetXRequestIDFromCtx(ctx, req)
}

// AuthToken returns vfs auth token for further requests.
func (c *Client) AuthToken(ctx context.Context) (string, error) {
	u := c.apiURL(authTokenURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("authtoken new reqeust failed: %w", err)
	}

	c.setHeaders(ctx, req)

	resp, err := c.opts.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("authtoken do: %w", newRequestError(0, u, err))
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authtoken request failed: %w", newRequestError(resp.StatusCode, u, ErrInternal))
	}

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("authtoken read failed: %w", err)
	}

	return string(body), nil
}

type HashUploadResponse struct {
	Error     string `json:"error,omitempty"`   // error message
	Hash      string `json:"hash,omitempty"`    // for hash
	Extension string `json:"ext,omitempty"`     // vfs file ext
	WebPath   string `json:"webPath,omitempty"` // for hash
	FileID    int    `json:"id,omitempty"`      // vfs file id
	Name      string `json:"name,omitempty"`    // vfs file name
}

// UploadFile uploads File to VFS. Filename with extension.
// Use empty namespace for default.
func (c *Client) UploadFile(ctx context.Context, token, namespace, filename string, file io.Reader) (*HashUploadResponse, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	// fill multipart form
	_ = w.WriteField("ns", namespace)
	_ = w.WriteField("ext", strings.TrimLeft(filepath.Ext(filename), "."))

	// write file to form
	part, err := w.CreateFormFile(c.opts.UploadFormName, filename)
	if err != nil {
		return nil, fmt.Errorf("upload file form: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("upload file copy: %w", err)
	}

	_ = w.Close()

	// create request
	u := c.apiURL(hashUploadURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &body)
	if err != nil {
		return nil, fmt.Errorf("upload file request: %w", err)
	}

	// set headers
	c.setHeaders(ctx, req)
	req.Header.Add("Content-Type", w.FormDataContentType())
	req.Header.Add(c.opts.AuthHeader, token)

	// do request
	resp, err := c.opts.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload file do: %w", newRequestError(0, u, err))
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// check response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("upload file read: %w", newRequestError(resp.StatusCode, u, err))
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("upload file: %w", newRequestError(resp.StatusCode, u, ErrUnauthorized))
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload file: %w body=%s", newRequestError(resp.StatusCode, u, ErrInternal), respBody)
	}

	var h = new(HashUploadResponse)
	if err = json.Unmarshal(respBody, h); err != nil {
		return nil, fmt.Errorf("upload file json: %w", newRequestError(resp.StatusCode, u, err))
	}

	if h.Error != "" {
		return nil, fmt.Errorf("upload file vfs: %w: %s", ErrInternal, h.Error)
	}

	return h, nil
}

// HashURL converts a 32-character hash into a hierarchical file path structure.
// It returns the original hash unchanged if the input is not exactly 32 characters.
// The resulting path format is: first_char/next_two_chars/full_hash
func HashURL(hash string) string {
	if len(hash) != 32 {
		return hash
	}

	return path.Join(
		hash[:1],
		hash[1:3],
		hash,
	)
}

// FilePath constructs a URL for accessing a media image in the VFS.
// It returns an empty string and no error if the hash is not 32 characters.
// The URL format is: base_url/namespace/size/hash_path.jpg
func (c *Client) FilePath(namespace, hash, size, ext string) (string, error) {
	if len(hash) != 32 {
		return "", nil
	}

	if ext == "" {
		ext = "jpg"
	}

	return url.JoinPath(c.opts.PublicURL,
		namespace,
		size,
		HashURL(hash)+"."+ext,
	)
}

// DownloadImage DownloadFile downloads vfs image with default extension and returns bytes.
func (c *Client) DownloadImage(ctx context.Context, namespace, hash, size string) ([]byte, error) {
	return c.DownloadFile(ctx, namespace, hash, "", size)
}

// DownloadFile downloads vfs file and returns bytes.
// Size and extension could be empty.
func (c *Client) DownloadFile(ctx context.Context, namespace, hash, ext, size string) ([]byte, error) {
	u, err := c.FilePath(namespace, hash, size, ext)
	if err != nil {
		return nil, fmt.Errorf("download construct url failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("vfs download new request failed: %w", err)
	}

	c.setHeaders(ctx, req)

	// do request
	resp, err := c.opts.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download do: %w", newRequestError(0, u, err))
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// check result
	switch resp.StatusCode {
	case http.StatusOK:
		return io.ReadAll(resp.Body)
	case http.StatusNotFound:
		return nil, newRequestError(resp.StatusCode, u, ErrNotFound)
	default:
		return nil, fmt.Errorf("vfs download failed: %w", newRequestError(resp.StatusCode, u, ErrInternal))
	}
}
