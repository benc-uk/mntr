package types

// Monitor holds shared common config for all monitors
type Monitor struct {
	Name      string
	Plugin    string
	Frequency int
	Enabled   bool
	RunsOn    []string `yaml:"runsOn"`
}
