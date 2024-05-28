package frame2

var id string
var smallId string

// Returns the string form of the UUID that identifies this test execution
//
// Every execution has a unique ID, that is logged at the very start of the
// execution.
func GetId() string {
	return id
}

// Returns the small format of GetId().  It's the three first hex characters
// of the MD5 form of GetId()
func GetSmallId() string {
	return smallId
}
