package toolspec

// SafetyLevel classifies how risky a command invocation is.
type SafetyLevel string

const (
	SafetyLevelSafe      SafetyLevel = "safe"
	SafetyLevelCaution   SafetyLevel = "caution"
	SafetyLevelDangerous SafetyLevel = "dangerous"
)

// Contract describes behavioral guarantees of a command.
type Contract struct {
	Idempotent    bool     `json:"idempotent,omitempty"`
	SideEffects   []string `json:"side_effects,omitempty"`
	Retryable     bool     `json:"retryable,omitempty"`
	PreConditions []string `json:"pre_conditions,omitempty"`
}

// Safety captures risk metadata for a command.
type Safety struct {
	Level                SafetyLevel `json:"level"`
	RequiresConfirmation bool        `json:"requires_confirmation,omitempty"`
	Permissions          []string    `json:"permissions,omitempty"`
}

// OutputSchema describes the expected output of a command.
type OutputSchema struct {
	Format  string   `json:"format,omitempty"`
	Fields  []string `json:"fields,omitempty"`
	Example string   `json:"example,omitempty"`
}

// StateIntrospection lists commands/vars for discovering tool state.
type StateIntrospection struct {
	ConfigCommands []string `json:"config_commands,omitempty"`
	EnvVars        []string `json:"env_vars,omitempty"`
	AuthCommands   []string `json:"auth_commands,omitempty"`
}

// Provenance records where a piece of spec data came from.
type Provenance struct {
	Source      string  `json:"source,omitempty"`
	RetrievedAt string  `json:"retrieved_at,omitempty"`
	Confidence  float32 `json:"confidence,omitempty"`
}

// Intent classifies a command's purpose.
type Intent struct {
	Domain   string   `json:"domain,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// ToolSpec captures everything known about a single CLI tool.
type ToolSpec struct {
	Name               string              `json:"name"`
	SchemaVersion      string              `json:"schema_version,omitempty"`
	Commands           []Command           `json:"commands,omitempty"`
	Flags              []Flag              `json:"flags,omitempty"`
	ErrorPatterns      []ErrorPattern      `json:"error_patterns,omitempty"`
	Workflows          []Workflow          `json:"workflows,omitempty"`
	StateIntrospection *StateIntrospection `json:"state_introspection,omitempty"`
}

// Command is a (sub)command in a CLI tool's command tree.
type Command struct {
	Name            string        `json:"name"`
	Aliases         []string      `json:"aliases,omitempty"`
	Flags           []Flag        `json:"flags,omitempty"`
	Children        []Command     `json:"children,omitempty"`
	Contract        *Contract     `json:"contract,omitempty"`
	Safety          *Safety       `json:"safety,omitempty"`
	PreviewModes    []string      `json:"preview_modes,omitempty"`
	OutputSchema    *OutputSchema `json:"output_schema,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty"`
	DeprecatedSince string        `json:"deprecated_since,omitempty"`
	ReplacedBy      string        `json:"replaced_by,omitempty"`
	Intent          *Intent       `json:"intent,omitempty"`
	SuggestedNext   []string      `json:"suggested_next,omitempty"`
}

// Flag describes a single CLI flag.
type Flag struct {
	Name        string `json:"name"`
	Short       string `json:"short,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty"`
	ReplacedBy  string `json:"replaced_by,omitempty"`
}

// ErrorPattern maps a known error output to a fix.
type ErrorPattern struct {
	Pattern    string      `json:"pattern"`
	Fix        string      `json:"fix"`
	Source     string      `json:"source,omitempty"`
	Cause      string      `json:"cause,omitempty"`
	Fixes      []string    `json:"fixes,omitempty"`
	Confidence float32     `json:"confidence,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
}

// Workflow describes a common multi-step sequence.
type Workflow struct {
	Name       string              `json:"name"`
	Steps      []string            `json:"steps"`
	After      map[string][]string `json:"after,omitempty"`
	Provenance *Provenance         `json:"provenance,omitempty"`
}

// FindCommand walks the command tree breadth-first and returns the
// shallowest Command whose Name matches name, or nil if not found.
func (ts *ToolSpec) FindCommand(name string) *Command {
	queue := make([]*Command, 0, len(ts.Commands))
	for i := range ts.Commands {
		queue = append(queue, &ts.Commands[i])
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c.Name == name {
			return c
		}
		for i := range c.Children {
			queue = append(queue, &c.Children[i])
		}
	}
	return nil
}
