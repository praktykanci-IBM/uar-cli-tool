package enums

type RequestState int

const (
	Requested RequestState = iota
	Granted
	Added
)

var stateName = map[RequestState]string{
	Requested: "requested",
	Granted:   "granted",
	Added:     "added",
}

func (rs RequestState) String() string {
	if name, ok := stateName[rs]; ok {
		return name
	}
	return "unknown"
}
