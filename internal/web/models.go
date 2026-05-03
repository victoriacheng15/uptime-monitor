package web

import "time"

type SiteConfig struct {
	Header       HeaderConfig        `yaml:"header"`
	SystemSpec   SystemSpecification `yaml:"system_specification"`
	Hero         HeroConfig          `yaml:"hero"`
	WhatIsUptime WhatIsUptimeMonitor `yaml:"what_is_uptime_monitor"`
	KeyFeatures  KeyFeaturesConfig   `yaml:"key_features"`
	WhyItMatters WhyItMattersConfig  `yaml:"why_it_matters"`
	MonitorSpecs []MonitorSpec       `yaml:"monitor_specs"`
	Footer       FooterConfig        `yaml:"footer"`
}

type MonitorSpec struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

type HeaderConfig struct {
	ProjectName string `yaml:"project_name"`
	SiteURL     string `yaml:"site_url"`
}

type SystemSpecification struct {
	Objective           string `yaml:"objective"`
	Stack               string `yaml:"stack"`
	Pattern             string `yaml:"pattern"`
	EntryPoint          string `yaml:"entry_point"`
	PersistenceStrategy string `yaml:"persistence_strategy"`
	Observability       string `yaml:"observability"`
	MachineRegistry     string `yaml:"machine_registry"`
}

type HeroConfig struct {
	Headline         string `yaml:"headline"`
	SubHeadline      string `yaml:"sub_headline"`
	BriefDescription string `yaml:"brief_description"`
	CTAText          string `yaml:"cta_text"`
	CTALink          string `yaml:"cta_link"`
	SecondaryCTAText string `yaml:"secondary_cta_text"`
	SecondaryCTALink string `yaml:"secondary_cta_link"`
	TertiaryCTAText  string `yaml:"tertiary_cta_text"`
	TertiaryCTALink  string `yaml:"tertiary_cta_link"`
}

type WhatIsUptimeMonitor struct {
	Title   string   `yaml:"title"`
	Content []string `yaml:"content"`
}

type KeyFeaturesConfig struct {
	Title    string    `yaml:"title"`
	Features []Feature `yaml:"features"`
}

type Feature struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

type WhyItMattersConfig struct {
	Title  string   `yaml:"title"`
	Points []string `yaml:"points"`
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

type TemplateData struct {
	Landing    *SiteConfig
	Evolution  *EvolutionConfig
	Uptime     *LatestResponse
	Year       int
	APIBaseURL string
}
