package enums

type requestState int

const (
	Requested requestState = iota
	Granted
	Added
)

var stateName = map[requestState]string{
	Requested: "requested",
	Granted:   "granted",
	Added:     "added",
}

func (rs requestState) String() string {
	if name, ok := stateName[rs]; ok {
		return name
	}
	return "unknown"
}
