package zendesk

type Zendesk struct {
	Email string `json:"email,omitempty"`
	Token string `json:"token"`
}

type Requester struct {
	Name string `json:"name"`
}
type Comment struct {
	Body string `json:"body"`
}
type Request struct {
	Requester Requester `json:"requester"`
	Subject   string    `json:"subject"`
	Comment   Comment   `json:"comment"`
	Username  string    `json:"username"`
	Service   string    `json:"service"`
}

func NewZendesk(token, email string) *Zendesk {
	z := &Zendesk{
		Email: email,
		Token: token,
	}

	return z
}

//func NewRequest(api *api.API) {
//
//}
