package grub2

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"pkg.deepin.io/lib/dbus"
	"pkg.deepin.io/lib/log"
	"regexp"
	"strings"
	"unicode"
)

var logger *log.Logger

func SetLogger(v *log.Logger) {
	logger = v
}

type Grub2 struct {
	configSaveChan  chan int
	mkconfigManager *MkconfigManager
	entries         []Entry
	theme           *Theme

	// props:
	DefaultEntry string
	EnableTheme  bool
	Resolution   string
	Timeout      uint32

	Updating bool
}

func (g *Grub2) saveLoop() {
	for {
		select {
		case <-g.configSaveChan:
			cfg := g.newConfig()
			err := cfg.Save()
			if err != nil {
				logger.Warning(err)
			}
			g.mkconfigManager.Change(cfg)
		}
	}
}

func (g *Grub2) saveConfig() {
	g.configSaveChan <- 1
}

// return -1 for failed
func (g *Grub2) defaultEntryStr2Idx(str string) int {
	entriesLv1 := g.getEntryTitlesLv1()
	return getStringIndexInArray(str, entriesLv1)
}

func (g *Grub2) defaultEntryIdx2Str(idx int) (string, error) {
	entriesLv1 := g.getEntryTitlesLv1()
	length := len(entriesLv1)
	if length == 0 {
		return "", errors.New("no entry")
	}
	if 0 <= idx && idx < length {
		return entriesLv1[idx], nil
	} else {
		return "", errors.New("index out of range")
	}
}

// Config >>> Props
func (g *Grub2) applyConfig(c *Config) {
	g.DefaultEntry, _ = g.defaultEntryIdx2Str(c.DefaultEntry)
	g.EnableTheme = c.EnableTheme
	g.Resolution = c.Resolution
	g.Timeout = c.Timeout
}

// Props >>> Config
func (g *Grub2) newConfig() *Config {
	c := NewConfig()
	c.DefaultEntry = g.defaultEntryStr2Idx(g.DefaultEntry)
	c.EnableTheme = g.EnableTheme
	c.Resolution = g.Resolution
	c.Timeout = g.Timeout
	return c
}

func loadConfig() (config *Config, saveCfg bool) {
	// read config file
	config = NewConfig()
	err := config.Load()
	if err != nil {
		saveCfg = true
		// try load old config
		c0 := NewConfigV0()
		err = c0.Load()
		if err == nil {
			config = c0.Upgrade()
		} else {
			// load old config failed
			config.UseDefault()
		}
	}
	logger.Debugf("loadConfig: %#v, saveCfg: %v", config, saveCfg)
	return
}

func (g *Grub2) fixConfig(config *Config) (saveCfg bool) {
	// fix config.DefaultEntry
	_, err := g.defaultEntryIdx2Str(config.DefaultEntry)
	if err != nil {
		logger.Warningf("config default entry %d is invalid", config.DefaultEntry)
		config.DefaultEntry = 0
		saveCfg = true
	}

	// fix config.Resolution
	err = checkResolution(config.Resolution)
	if err != nil {
		logger.Warningf("config resolution %q is invalid", config.Resolution)
		config.Resolution = defaultResolution
		saveCfg = true
	}
	// no fix config.Timeout
	// no fix config.EnableTheme
	return
}

func New() *Grub2 {
	g := &Grub2{}

	g.readEntries()

	config, saveCfg := loadConfig()
	saveCfg = g.fixConfig(config)

	// check log
	l, err := loadLog()
	if err != nil {
		logger.Warning("loadLog failed:", err)
		saveCfg = true
	} else {
		// load log success
		logger.Debugf("log: %#v", l)
		if ok, _ := l.Verify(config); !ok {
			logger.Warning("log verify failed")
			saveCfg = true
		}
	}

	g.applyConfig(config)

	g.configSaveChan = make(chan int)
	go g.saveLoop()

	g.mkconfigManager = newMkconfigManager(func(running bool) {
		// state change callback
		if g.Updating != running {
			g.Updating = running
			dbus.NotifyChange(g, "Updating")
		}
	})

	// init theme
	g.theme = NewTheme(g)
	g.theme.initTheme()
	go g.theme.regenerateBackgroundIfNeed()

	if saveCfg {
		g.saveConfig()
	}
	return g
}

func (grub *Grub2) readEntries() (err error) {
	fileContent, err := ioutil.ReadFile(grubScriptFile)
	if err != nil {
		logger.Error(err)
		return
	}
	err = grub.parseEntries(string(fileContent))
	if err != nil {
		logger.Error(err)
		return
	}
	if len(grub.entries) == 0 {
		logger.Warning("there is no menu entry in %s", grubScriptFile)
	}
	return
}

func (grub *Grub2) resetEntries() {
	grub.entries = make([]Entry, 0)
}

// getAllEntriesLv1 return all entires titles in level one.
func (grub *Grub2) getEntryTitlesLv1() (entryTitles []string) {
	for _, entry := range grub.entries {
		if entry.parentSubMenu == nil {
			entryTitles = append(entryTitles, entry.getFullTitle())
		}
	}
	return
}

func (grub *Grub2) parseEntries(fileContent string) (err error) {
	grub.resetEntries()

	inMenuEntry := false
	level := 0
	numCount := make(map[int]int)
	numCount[0] = 0
	parentMenus := make([]*Entry, 0)
	parentMenus = append(parentMenus, nil)
	sl := bufio.NewScanner(strings.NewReader(fileContent))
	sl.Split(bufio.ScanLines)
	for sl.Scan() {
		line := sl.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "menuentry ") {
			if inMenuEntry {
				grub.resetEntries()
				err = fmt.Errorf("a 'menuentry' directive was detected inside the scope of a menuentry")
				return
			}
			title, ok := parseTitle(line)
			if ok {
				entry := Entry{MENUENTRY, title, numCount[level], parentMenus[len(parentMenus)-1]}
				grub.entries = append(grub.entries, entry)
				logger.Debugf("found entry: [%d] %s %s", level, strings.Repeat(" ", level*2), title)

				numCount[level]++
				inMenuEntry = true
				continue
			} else {
				grub.resetEntries()
				err = fmt.Errorf("parse entry title failed from: %q", line)
				return
			}
		} else if strings.HasPrefix(line, "submenu ") {
			if inMenuEntry {
				grub.resetEntries()
				err = fmt.Errorf("a 'submenu' directive was detected inside the scope of a menuentry")
				return
			}
			title, ok := parseTitle(line)
			if ok {
				entry := Entry{SUBMENU, title, numCount[level], parentMenus[len(parentMenus)-1]}
				grub.entries = append(grub.entries, entry)
				parentMenus = append(parentMenus, &entry)
				logger.Debugf("found entry: [%d] %s %s", level, strings.Repeat(" ", level*2), title)

				level++
				numCount[level] = 0
				continue
			} else {
				grub.resetEntries()
				err = fmt.Errorf("parse entry title failed from: %q", line)
				return
			}
		} else if line == "}" {
			if inMenuEntry {
				inMenuEntry = false
			} else if level > 0 {
				level--

				// delete last parent submenu
				i := len(parentMenus) - 1
				copy(parentMenus[i:], parentMenus[i+1:])
				parentMenus[len(parentMenus)-1] = nil
				parentMenus = parentMenus[:len(parentMenus)-1]
			}
		}
	}
	err = sl.Err()
	if err != nil {
		return
	}
	return
}

var (
	entryRegexpSingleQuote = regexp.MustCompile(`^ *(menuentry|submenu) +'(.*?)'.*$`)
	entryRegexpDoubleQuote = regexp.MustCompile(`^ *(menuentry|submenu) +"(.*?)".*$`)
)

func parseTitle(line string) (string, bool) {
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	if entryRegexpSingleQuote.MatchString(line) {
		return entryRegexpSingleQuote.FindStringSubmatch(line)[2], true
	} else if entryRegexpDoubleQuote.MatchString(line) {
		return entryRegexpDoubleQuote.FindStringSubmatch(line)[2], true
	} else {
		return "", false
	}
}

func (g *Grub2) getScreenWidthHeight() (w, h uint16, err error) {
	return parseResolution(g.Resolution)
}

func (g *Grub2) canSafelyExit() bool {
	logger.Debug("call canSafelyExit")
	if g.Updating || g.theme.Updating {
		return false
	}
	return true
}
