package enums

type RequestState int

const (
	Granted RequestState = iota
	Added
)

var stateName = map[RequestState]string{
	Granted: "granted",
	Added:   "added",
}

func (rs RequestState) String() string {
	if name, ok := stateName[rs]; ok {
		return name
	}
	return "unknown"
}
