package zendesk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type Zendesk struct {
	Email string `json:"email,omitempty"`
	Token string `json:"token"`
	Key   string `json:"key"`
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

func NewZendesk(token, email string) *Zendesk {
	key := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + token))
	z := &Zendesk{
		Email: email,
		Token: token,
		Key:   key,
	}

	return z
}

func (z *Zendesk) NewRequest(username, userID, subject, comment, service string) (int, error) {
	fmt.Println(z.Key)
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

	req, err := http.NewRequest("POST", "https://supremacyhelp.zendesk.com/api/v2/requests.json", body)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req.Header.Set("Authorization", "Basic "+z.Key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
