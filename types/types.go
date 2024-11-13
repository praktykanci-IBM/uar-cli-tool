package types

type RequestData struct {
	ID            string `yaml:"id"`
	Added         bool   `yaml:"added"`
	Justification string `yaml:"justification"`
	RequestedOn   string `yaml:"requested_on"`
	RequestedBy   string `yaml:"requested_by"`
}

type CbnData struct {
	Owner string    `yaml:"owner"`
	Repo  string    `yaml:"repo"`
	Type  string    `yaml:"type"`
	Users []CbnUser `yaml:"users"`
}

type CbnDataCompleted struct {
	Owner        string    `yaml:"owner"`
	Repo         string    `yaml:"repo"`
	Type         string    `yaml:"type"`
	Users        []CbnUser `yaml:"users"`
	ExecutedBy   string    `yaml:"executed_by"`
	ExecutedOn   string    `yaml:"executed_on"`
	UsersChanged []CbnUser `yaml:"userschanged"`
}

type CbnUserApproval int

const (
	Unset CbnUserApproval = iota
	Aproved
	Rejected
)

type CbnUser struct {
	Name   string          `yaml:"name"`
	Status CbnUserApproval `yaml:"status"`
}
