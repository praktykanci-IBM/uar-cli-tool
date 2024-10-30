package cmd

type GitResponseData struct {
	Content string `json:"content"`
	Sha     string `json:"sha"`
}

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
	ApproverID    string `json:"approver"`
	AdminID       string `json:"admin"`
	WhenCompleted int64  `json:"whenCompleted"`
}
type GrantedRequests struct {
	Requests []GrantedRequest `json:"grantedRequests"`
}
