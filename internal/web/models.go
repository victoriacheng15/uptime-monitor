package web

import "time"

type SiteConfig struct {
	Header       HeaderConfig       `yaml:"header"`
	LLMS         LLMSConfig         `yaml:"llms"`
	Architecture ArchitectureConfig `yaml:"architecture"`
	Tech         []ComponentConfig  `yaml:"tech"`
	Proof        []ProofDetail      `yaml:"proof"`
	Reach        ReachConfig        `yaml:"reach"`
	Footer       FooterConfig       `yaml:"footer"`
	MonitorSpecs []MonitorSpec      `yaml:"monitor_specs"`
}

type HeaderConfig struct {
	ProjectName string `yaml:"project_name"`
	SiteURL     string `yaml:"site_url"`
}

type LLMSConfig struct {
	Objective           string `yaml:"objective"`
	Stack               string `yaml:"stack"`
	Pattern             string `yaml:"pattern"`
	EntryPoint          string `yaml:"entry_point"`
	PersistenceStrategy string `yaml:"persistence_strategy"`
	Observability       string `yaml:"observability"`
}

type ArchitectureConfig struct {
	DiagramASCII string `yaml:"diagram_ascii"`
}

type ComponentConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type ProofDetail struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type ReachConfig struct {
	HumblePivots      []PivotConfig            `yaml:"humble_pivots"`
	ObjectiveClarity  ObjectiveClarityConfig   `yaml:"objective_clarity"`
	VerifiableOutputs []VerifiableOutputConfig `yaml:"verifiable_outputs"`
}

type PivotConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}

type ObjectiveClarityConfig struct {
	Description string `yaml:"description"`
}

type VerifiableOutputConfig struct {
	Title          string `yaml:"title"`
	TerminalOutput string `yaml:"terminal_output"`
}

type MonitorSpec struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

type FooterConfig struct {
	Author       string `yaml:"author"`
	GithubLink   string `yaml:"github_link"`
	LinkedinLink string `yaml:"linkedin_link"`
}

type EvolutionConfig struct {
	PageTitle string    `yaml:"page_title"`
	IntroText string    `yaml:"intro_text"`
	Chapters  []Chapter `yaml:"chapters"`
}

type Chapter struct {
	Title    string     `yaml:"title"`
	Intro    string     `yaml:"intro"`
	Timeline []Timeline `yaml:"timeline"`
}

type Timeline struct {
	Date        string     `yaml:"date"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Artifacts   []Artifact `yaml:"artifacts"`
}

type Artifact struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// Uptime data from Lambda (for build-time hydration)
type CheckResult struct {
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	IsUp       bool      `json:"is_up"`
	LatencyMS  int       `json:"latency_ms"`
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error"`
}

type LatestResponse struct {
	Sites     []CheckResult `json:"sites"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type HistoryResponse struct {
	History map[string][]CheckResult `json:"history"`
}

type LatencyAverage struct {
	URL              string
	AverageLatencyMS int
	History          []CheckResult
}

type TemplateData struct {
	Landing         *SiteConfig
	Evolution       *EvolutionConfig
	Uptime          *LatestResponse
	LatencyAverages []LatencyAverage
	Year            int
	APIBaseURL      string
	PageName        string
}
