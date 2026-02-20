package conversation

import "time"

type SourceType string

const (
	SourceTypingMind SourceType = "typingmind"
	SourceOpenAI     SourceType = "openai"
)

type Source struct {
	ID           string
	SourceType   SourceType
	SourcePath   string
	ImportedAt   time.Time
	MetadataJSON string
}

type Folder struct {
	ID             string
	SourceID       string
	SourceFolderID string
	Title          string
	ParentID       *string
	SortOrder      int
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	SettingsJSON   string
	MetadataJSON   string
}

type Character struct {
	ID                string
	SourceID          string
	SourceCharacterID string
	Name              string
	Description       *string
	AvatarURL         *string
	Instruction       *string
	Categories        string
	SettingsJSON      string
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	MetadataJSON      string
}

type Prompt struct {
	ID             string
	SourceID       string
	SourcePromptID string
	Name           string
	Content        *string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	MetadataJSON   string
}

type Setting struct {
	Key       string
	ValueJSON string
	SourceID  *string
}

type Conversation struct {
	ID                   string
	SourceID             string
	SourceConversationID string
	SourceType           SourceType
	Title                *string
	Model                *string
	Preview              *string
	TotalTokens          int
	FolderID             *string
	CharacterID          *string
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
	MetadataJSON         string
}

type Message struct {
	ID              string
	ConversationID  string
	SourceMessageID string
	ParentID        *string
	Role            string
	Content         *string
	ContentJSON     string
	Model           *string
	TokenCount      int
	UsageJSON       string
	CreatedAt       *time.Time
	MetadataJSON    string
}

type File struct {
	ID             string
	SourceID       string
	SourceFileID   *string
	FileName       string
	FileType       *string
	FilePath       *string
	SourcePath     *string
	ContentText    *string
	MetadataJSON   string
	ConversationID *string
	CharacterID    *string
	CreatedAt      *time.Time
}

type ImportResult struct {
	TotalConversations int
	TotalMessages      int
	TotalFolders       int
	TotalCharacters    int
	TotalPrompts       int
	TotalFiles         int
	Inserted           int
	Skipped            int
	Errors             []ImportError
	Duration           time.Duration
}

type ImportError struct {
	Type string
	ID   string
	Err  error
	Data map[string]string
}

func (e ImportError) Error() string {
	return e.Err.Error()
}

func (r ImportResult) IsSuccess() bool {
	return len(r.Errors) == 0
}
