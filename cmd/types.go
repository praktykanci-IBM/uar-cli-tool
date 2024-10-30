package cmd

type Request struct {
	Name          string `json:"name"`
	When          int64  `json:"when"`
	Justification string `json:"justification"`
	Repo          string `json:"repo"`
	ID            string `json:"id"`
}
type Requests struct {
	Requests []Request `json:"requests"`
}

type GrantedRequest struct {
	Name          string `json:"name"`
	WhenRequested int64  `json:"whenRequested"`
	WhenAccepted  int64  `json:"whenAccepted"`
	Justification string `json:"justification"`
	Repo          string `json:"repo"`
	ID            string `json:"id"`
	Approver      string `json:"approver"`
}
type GrantedRequests struct {
	Requests []GrantedRequest `json:"grantedRequests"`
}
