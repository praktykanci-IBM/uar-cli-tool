package types

type RequestData struct {
	ID            string           `yaml:"id"`
	State         RequestUserState `yaml:"state"`
	Justification string           `yaml:"justification"`
	RequestedOn   string           `yaml:"requested_on"`
	RequestedBy   string           `yaml:"requested_by"`
}

type RequestDataCompleted struct {
	ID            string           `yaml:"id"`
	State         RequestUserState `yaml:"state"`
	Justification string           `yaml:"justification"`
	RequestedOn   string           `yaml:"requested_on"`
	RequestedBy   string           `yaml:"requested_by"`
	CompletedOn   string           `yaml:"completed_on"`
	CompletedBy   string           `yaml:"completed_by"`
}

type RequestUserState string

const (
	Granted   RequestUserState = "granted"
	Completed RequestUserState = "completed"
)

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

type CbnUserApproval string

const (
	Unset    CbnUserApproval = "unset"
	Aproved  CbnUserApproval = "approved"
	Rejected CbnUserApproval = "rejected"
)

type CbnUser struct {
	Name   string          `yaml:"name"`
	Status CbnUserApproval `yaml:"status"`
}
