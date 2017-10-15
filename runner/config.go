package runner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

const dftConf = ".air.conf"

type config struct {
	Root     string   `toml:"root"`
	WatchDir string   `toml:"watch_dir"`
	TmpDir   string   `toml:"tmp_dir"`
	Build    cfgBuild `toml:"build"`
	Color    cfgColor `toml:"color"`
}

type cfgBuild struct {
	Bin        string   `toml:"bin"`
	Cmd        string   `toml:"cmd"`
	Log        string   `toml:"log"`
	IncludeExt []string `toml:"include_ext"`
	ExcludeDir []string `toml:"exclude_dir"`
	Delay      int      `toml:"delay"`
}

type cfgColor struct {
	Main    string `toml:"main"`
	Watcher string `toml:"watcher"`
	Build   string `toml:"build"`
	Runner  string `toml:"runner"`
	App     string `toml:"app"`
}

// InitConfig loads config info
func InitConfig(path string) (*config, error) {
	var err error
	var useDftCfg bool
	dft := defaultConfig()
	if path == "" {
		useDftCfg = true
		// when path is blank, first find `.air.conf` in root directory, if not found, use defaults
		path, err = dftConfPath()
		if err != nil {
			return &dft, nil
		}
	}
	cfg, err := readConfig(path)
	if err != nil {
		if !useDftCfg {
			return nil, err
		}
		cfg = &dft
	}
	err = cfg.preprocess()
	return cfg, err
}

func defaultConfig() config {
	build := cfgBuild{
		Bin:        "tmp/main",
		Cmd:        "go build -o ./tmp/main main.go",
		Log:        "build-errors.log",
		IncludeExt: []string{"go", "tpl", "tmpl", "html"},
		ExcludeDir: []string{"assets", "tmp", "vendor"},
		Delay:      1000,
	}
	color := cfgColor{
		Main:    "magenta",
		Watcher: "cyan",
		Build:   "yellow",
		Runner:  "green",
		App:     "white",
	}
	return config{
		Root:     ".",
		TmpDir:   "tmp",
		WatchDir: "",
		Build:    build,
		Color:    color,
	}
}

func readConfig(path string) (*config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := config{}
	if err = toml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *config) preprocess() error {
	// TODO: merge defaults if some options are not set
	var err error
	c.Root, err = expandPath(c.Root)
	if c.TmpDir == "" {
		c.TmpDir = "tmp"
	}
	if err != nil {
		return err
	}
	ed := c.Build.ExcludeDir
	for i := range ed {
		ed[i] = cleanPath(ed[i])
	}
	c.Build.ExcludeDir = ed
	return nil
}

func (c *config) colorInfo() map[string]string {
	return map[string]string{
		"main":    c.Color.Main,
		"build":   c.Color.Build,
		"runner":  c.Color.Runner,
		"watcher": c.Color.Watcher,
		"app":     c.Color.App,
	}
}

func dftConfPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, dftConf), nil
}

func (c *config) WatchDirRoot() string {
	if c.WatchDir != "" {
		return c.FullPath(c.WatchDir)
	}
	return c.Root
}

func (c *config) BuildLogPath() string {
	return filepath.Join(c.TmpPath(), c.Build.Log)
}

func (c *config) BuildDelay() time.Duration {
	return time.Duration(c.Build.Delay) * time.Millisecond
}

func (c *config) FullPath(path string) string {
	return filepath.Join(c.Root, path)
}

func (c *config) BinPath() string {
	return filepath.Join(c.Root, c.Build.Bin)
}

func (c *config) TmpPath() string {
	return filepath.Join(c.Root, c.TmpDir)
}

func (c *config) Rel(path string) string {
	s, err := filepath.Rel(c.Root, path)
	if err != nil {
		return ""
	}
	return s
}