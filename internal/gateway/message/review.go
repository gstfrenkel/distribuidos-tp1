package message

type ReviewMsg struct {
	appId       int64
	appName     string
	reviewText  string
	reviewScore int64
	reviewVotes int64
}

func (r *ReviewMsg) ToBytes() ([]byte, error) {
	//TODO implement
	panic("implement me")
}

func (r *ReviewMsg) FromBytes(bytes []byte) error {
	//TODO implement
	panic("implement me")
}
