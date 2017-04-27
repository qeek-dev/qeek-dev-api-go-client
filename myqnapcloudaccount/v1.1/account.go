package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func New(client *http.Client) *Service {
	if client == nil {
		client = http.DefaultClient
	}
	s := &Service{client: client, BasePath: basePath}
	s.Me = NewMeService(s)
	return s
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Me     *MeService
	Friend *FriendService
	User   *UserService

	// Set to true to output debugging logs during API calls
	Debug bool
}

func versioned(path string) string {
	return fmt.Sprintf("/%s/%s", apiVersion, strings.Trim(path, "/"))
}

// NewRequest creates an API request.
// The path is expected to be a relative path and will be resolved
// according to the BaseURL of the Client. Paths should always be specified without a preceding slash.
func (c *Service) doRequest(method, path string, payload interface{}) (*http.Request, error) {
	url := c.BasePath + path

	body := new(bytes.Buffer)
	if payload != nil {
		err := json.NewEncoder(body).Encode(payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	//req.Header.Add("User-Agent", formatUserAgent(c.UserAgent))

	return req, nil
}

func (c *Service) get(path string, obj interface{}) (*http.Response, error) {
	req, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, obj)
}

func (c *Service) post(path string, payload, obj interface{}) (*http.Response, error) {
	req, err := c.doRequest("POST", path, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, obj)
}

func (c *Service) put(path string, payload, obj interface{}) (*http.Response, error) {
	req, err := c.doRequest("PUT", path, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, obj)
}

func (c *Service) patch(path string, payload, obj interface{}) (*http.Response, error) {
	req, err := c.doRequest("PATCH", path, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, obj)
}

func (c *Service) delete(path string, payload interface{}, obj interface{}) (*http.Response, error) {
	req, err := c.doRequest("DELETE", path, payload)
	if err != nil {
		return nil, err
	}

	return c.do(req, obj)
}

// Do sends an API request and returns the API response.
//
// The API response is JSON decoded and stored in the value pointed by obj,
// or returned as an error if an API error has occurred.
// If obj implements the io.Writer interface, the raw response body will be written to obj,
// without attempting to decode it.
func (c *Service) do(req *http.Request, obj interface{}) (*http.Response, error) {
	if c.Debug {
		log.Printf("Executing request (%v): %#v", req.URL, req)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.Debug {
		log.Printf("Response received: %#v", resp)
	}

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	// If obj implements the io.Writer,
	// the response body is decoded into v.
	if obj != nil {
		if w, ok := obj.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(obj)
		}
	}

	return resp, err
}

type GetUserResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Result  struct {
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		DisplayName  string `json:"display_name"`
		Subscribed   bool   `json:"subscribed"`
		Language     string `json:"language"`
		Gender       int    `json:"gender"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		PortalNotify bool   `json:"portal_notify"`
		SimpleToken  string `json:"simple_token"`
		Brithday     string `json:"brithday"`
		MobileNumber string `json:"mobile_number"`
		UserId       string `json:"user_id"`
		Email        string `json:"email"`
	} `json:"result"`
}

type MeService struct {
	s *Service

	Activity *ActivityService
	Password *PasswordService
	Avatar   *AvatarService
}

func NewMeService(s *Service) *MeService {
	rs := &MeService{s: s}
	rs.Activity = NewActivityService(s)
	rs.Password = NewPasswordService(s)
	rs.Avatar = NewAvatarService(s)
	return rs
}

type MeGetCall struct {
	s *Service
}

func (r *MeService) Get() *MeGetCall {
	c := &MeGetCall{s: r.s}
	return c
}

func (c *MeGetCall) Do() (*GetUserResponse, error) {
	path := versioned("me")
	ret := &GetUserResponse{}
	_, err := c.s.get(path, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type ActivityService struct {
	s *Service
}

func NewActivityService(s *Service) *ActivityService {
	rs := &ActivityService{s: s}
	return rs
}

type PasswordService struct {
	s *Service
}

func NewPasswordService(s *Service) *PasswordService {
	rs := &PasswordService{s: s}
	return rs
}

type AvatarService struct {
	s *Service
}

func NewAvatarService(s *Service) *AvatarService {
	rs := &AvatarService{s: s}
	return rs
}

type FriendService struct {
	s *Service
}

func NewFriendService(s *Service) *FriendService {
	rs := &FriendService{s: s}
	return rs
}

type UserService struct {
	s *Service
}

func NewUserService(s *Service) *UserService {
	rs := &UserService{s: s}
	return rs
}

//-----------------------------------------------------------------------------
// A Response represents an API response.
type Response struct {
	// HTTP response
	HttpResponse *http.Response
}

// An ErrorResponse represents an API response that generated an error.
type ErrorResponse struct {
	Response

	// human-readable message
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Error implements the error interface.
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %v %v",
		r.HttpResponse.Request.Method, r.HttpResponse.Request.URL,
		r.HttpResponse.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if present.
// A response is considered an error if the status code is different than 2xx. Specific requests
// may have additional requirements, but this is sufficient in most of the cases.
func CheckResponse(resp *http.Response) error {
	if code := resp.StatusCode; 200 <= code && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{}
	errorResponse.HttpResponse = resp

	err := json.NewDecoder(resp.Body).Decode(errorResponse)
	if err != nil {
		return err
	}

	return errorResponse
}
