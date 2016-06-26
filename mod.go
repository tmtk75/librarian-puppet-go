package librarianpuppetgo

import (
	"fmt"
	"path/filepath"
)

type ModOpts map[string]string

// mod 'puppetlabs/stdlib', '4.1.0'
// mod 'fiz', :git => 'git@github.com:foo/bar.git', :ref => 'v0.4.1'
type Mod struct {
	name    string  // stdlib, fiz
	user    string  // puppetlabs
	version string  // 4.1.0
	opts    ModOpts // git => git@github.com:foo/bar.git, ref => v0.4.1
	cmd     string  // clone, fetch, checkout
	err     error
}

func (m Mod) Fullname() string {
	if m.user != "" {
		return m.user + "/" + m.name
	}
	return m.name
}

func (m Mod) String() string {
	return fmt.Sprintf("name:%v\topts:%v\tuser:%v\tversion:%v", m.name, m.opts, m.user, m.version)
}

func (m Mod) Dest() string {
	return filepath.Join(modulePath, m.name)
}

func (m *Mod) Replace(e *Mod) {
	m.user = e.user
	m.version = e.version
	for k, v := range e.opts {
		m.opts[k] = v
	}
}

func (m Mod) Format() string {
	if m.opts["git"] == "" && m.opts["ref"] == "" {
		if m.user != "" {
			if m.version != "" {
				return fmt.Sprintf("mod '%s/%s', '%s'", m.user, m.name, m.version)
			} else {
				return fmt.Sprintf("mod '%s/%s'", m.user, m.name)
			}
		} else {
			return fmt.Sprintf("mod '%s', '%s'", m.name, m.version)
		}
	}
	return fmt.Sprintf("mod '%s', :git => '%s', :ref => '%s'", m.name, m.opts["git"], m.Ref())
}

var defaultBranch string = ""

func (m Mod) Ref() string {
	if m.opts["ref"] == "" {
		if m.version == "" {
			return defaultBranch
		}
		return m.version
	}
	return m.opts["ref"]
}
