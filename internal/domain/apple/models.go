package apple

import "time"

type Contact struct {
	ID           int64
	Name         string
	Phone        string
	Email        string
	Organization string
	SourceFile   string
	CreatedAt    string
}

type Note struct {
	ID         int64
	Title      string
	Content    string
	Folder     string
	SourceFile string
	CreatedAt  string
	UpdatedAt  string
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
