package social

import "time"

type Post struct {
	ID        int64
	Platform  string
	PostType  string
	Caption   string
	Location  string
	Timestamp string
	Path      string
	MediaURL  string
	Metadata  string
}

type Like struct {
	ID             int64
	Platform       string
	Timestamp      string
	TargetUsername string
	TargetType     string
	Reaction       string
}

type Connection struct {
	ID             int64
	Platform       string
	ConnectionType string
	Username       string
	Timestamp      string
}

type Comment struct {
	ID        int64
	Platform  string
	Timestamp string
	Username  string
	Text      string
}

type Message struct {
	ID         int64
	Platform   string
	Timestamp  string
	Sender     string
	Receiver   string
	Text       string
	MediaURL   string
	StoryShare string
	Metadata   string
}

type ImportResult struct {
	TotalRows int
	Inserted  int
	Skipped   int
	Errors    []ImportError
	Duration  time.Duration
}

type ImportError struct {
	Row  int
	Col  int
	Err  error
	Data map[string]string
}

func (e ImportError) Error() string {
	return e.Err.Error()
}

func (r ImportResult) IsSuccess() bool {
	return len(r.Errors) == 0
}
