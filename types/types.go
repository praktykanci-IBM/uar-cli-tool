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
	Removed   RequestUserState = "removed"
)

type CbnData struct {
	StartedBy    string    `yaml:"started_by"`
	StartedOn    string    `yaml:"started_on"`
	Org          string    `yaml:"org"`
	Type         string    `yaml:"type"`
	ExtractedBy  string    `yaml:"extracted_by"`
	ExtractedOn  string    `yaml:"extracted_on"`
	Users        []CbnUser `yaml:"users"`
	ExecutedBy   string    `yaml:"executed_by"`
	ExecutedOn   string    `yaml:"executed_on"`
	UsersChanged []CbnUser `yaml:"userschanged"`
}

// type CbnDataCompleted struct {
// 	StartedBy    string    `yaml:"started_by"`
// 	StartedOn    string    `yaml:"started_on"`
// 	Repo         string    `yaml:"repo"`
// 	Type         string    `yaml:"type"`
// 	ExtractedBy  string    `yaml:"extracted_by"`
// 	ExtractedOn  string    `yaml:"extracted_on"`
// 	Users        []CbnUser `yaml:"users"`
// 	ExecutedBy   string    `yaml:"executed_by"`
// 	ExecutedOn   string    `yaml:"executed_on"`
// 	UsersChanged []CbnUser `yaml:"userschanged"`
// }

type CbnUserApproval string

const (
	Pending  CbnUserApproval = "pending"
	Aproved  CbnUserApproval = "approved"
	Rejected CbnUserApproval = "rejected"
)

type UserAccess struct {
	AccessType    string `yaml:"access_type"`
	AccessTo      string `yaml:"access_to"`
	Justification string `yaml:"justification"`
}

type CbnUser struct {
	Name           string          `yaml:"name"`
	State          CbnUserApproval `yaml:"state"`
	ListOfAccesses []UserAccess    `yaml:"list_of_accesses"`
	ValidatedOn    string          `yaml:"validated_on"`
	ValidatedBy    string          `yaml:"validated_by"`
}
