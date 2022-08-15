package zendesk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Zendesk struct {
	Email string `json:"email,omitempty"`
	Token string `json:"token"`
	Key   string `json:"key"`
	Url   string `json:"url"`
}

type Requester struct {
	Name string `json:"name"`
}
type Comment struct {
	Body string `json:"body"`
}
type RequestObj struct {
	Requester Requester `json:"requester"`
	Subject   string    `json:"subject"`
	Comment   Comment   `json:"comment"`
	Username  string    `json:"username"`
	Service   string    `json:"service"`
}
type RequestJSON struct {
	Request *RequestObj `json:"request"`
}

type RequestErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description"`
	Details     struct {
		Base []struct {
			Description string `json:"description"`
			Error       string `json:"error"`
			FieldKey    int64  `json:"field_key"`
		} `json:"base"`
	} `json:"details"`
}

func NewZendesk(token, email, url string) *Zendesk {
	key := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + token))
	z := &Zendesk{
		Email: email,
		Token: token,
		Key:   key,
		Url:   url,
	}

	return z
}

func (z *Zendesk) NewRequest(username, userID, subject, comment, service string) (int, error) {
	//organize data
	request := &RequestObj{
		Requester: Requester{
			Name: userID,
		},
		Subject: subject,
		Comment: Comment{
			Body: comment,
		},
		Username: username,
		Service:  service,
	}

	reqJSON := &RequestJSON{Request: request}
	//marshall
	payloadBytes, err := json.Marshal(reqJSON)
	if err != nil {
		return http.StatusBadRequest, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, z.Url, body)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", z.Key))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody := &RequestErrorResponse{}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return http.StatusBadRequest, err
		}

		err = json.Unmarshal(bodyBytes, errorBody)
		if err != nil {
			return http.StatusBadRequest, err
		}

		return http.StatusBadRequest, fmt.Errorf(errorBody.Error)
	}

	return resp.StatusCode, nil
}
