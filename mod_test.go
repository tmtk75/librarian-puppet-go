package main

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func r(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func TestParsePuppetfile(t *testing.T) {
	var mods []Mod
	var err error

	mods, _ = parsePuppetfile(r(``))
	if !(len(mods) == 0) {
		t.Errorf("should be 0")
	}

	_, err = parsePuppetfile(r(`mod ""`))
	if !(err != nil) {
		t.Errorf("should be error")
	}

	mods, _ = parsePuppetfile(r(`
#
mod 'foo/bar0'
mod 'foo/bar1', '0.1.2'

mod 'foo2', :git => 'git-url'
mod 'foo3', :git => 'git-url', :ref => 'v0.1.1'  #
`))
	if !(len(mods) == 4) {
		t.Errorf("should be 4: %v", len(mods))
	}

	// For testing of include
	fname2body := map[string]string{
		"Puppetfile.1": `mod 'foo/bar', '0.0.1'`,
		"Puppetfile.2": `mod 'fiz/biz', '0.0.1'`,
		"Puppetfile.3": `mod 'my_api', :git => 'github/000', :ref => 'dev'`,
		"Puppetfile.4": `mod 'aaa/bbb'`,
	}
	newReader = func(n string) io.ReadCloser { return r(fname2body[n]) }

	mods, err = parsePuppetfile(r(`
mod 'fiz/bar', '0.1.0'
include "hello"
`))
	if !(err != nil) {
		t.Errorf("should be error")
	}

	mods, _ = parsePuppetfile(r(`
include "Puppetfile.1"
mod 'fiz/biz', '0.1.0'
`))
	if !(len(mods) == 2) {
		t.Errorf("should be 2: %v", len(mods))
	}

	// override if name matches
	mods, _ = parsePuppetfile(r(`
include "Puppetfile.2"
mod 'fiz/biz', '0.1.0'
`))
	if !(len(mods) == 1) {
		t.Errorf("should be 1: %v", len(mods))
	}
	if !(mods[0].version == "0.1.0") {
		t.Errorf("should be '0.1.0': %v", mods[0].version)
	}

	mods, _ = parsePuppetfile(r(`
include "Puppetfile.3"
`))
	if !(len(mods) == 1) {
		t.Errorf("should be 1: %v", len(mods))
	}
	if !(mods[0].name == "my_api" && mods[0].opts["ref"] == "dev" && mods[0].opts["git"] == "github/000" && mods[0].version == "") {
		t.Errorf("should be '0.1.0': %v", mods[0].version)
	}

	mods, _ = parsePuppetfile(r(`
include "Puppetfile.4"
mod 'ccc/ddd'
mod 'eee/fff'
`))
	if !(len(mods) == 3) {
		t.Errorf("should be 3: %v", len(mods))
	}
	if !(mods[0].name == "bbb" && mods[1].name == "ddd" && mods[2].name == "fff") {
		t.Errorf("%v", mods)
	}

	mods, _ = parsePuppetfile(r(`
include "Puppetfile.1"
include "Puppetfile.2"
include "Puppetfile.3"
include "Puppetfile.4"
mod 'xxx/yyy'
`))
	if !(len(mods) == 5) {
		t.Errorf("should be 5: %v", len(mods))
	}
	if !(mods[0].name == "bar" && mods[1].name == "biz" && mods[2].name == "my_api" && mods[3].name == "bbb" && mods[4].name == "yyy") {
		t.Errorf("%v", mods)
	}

	mods, _ = parsePuppetfile(r(`
# include "Puppetfile.1"
`))
	if !(len(mods) == 0) {
		t.Errorf("should be 0: %v", len(mods))
	}

	mods, _ = parsePuppetfile(r(`
mod 'xxx/yyy', '0.0.1'
mod 'xxx/yyy', '0.0.2'
mod 'xxx/yyy', '0.0.3'
`))
	if !(len(mods) == 1 && mods[0].version == "0.0.3") {
		t.Errorf("%v", mods)
	}

	mods, err = parsePuppetfile(r(`
forge "http://forge.puppetlabs.com"
mod 'xxx/yyy', '0.0.1'
`))
	if !(err == nil) {
		t.Errorf("%v", err)
	}
	if !(len(mods) == 1) {
		t.Errorf("%v", mods)
	}
}

func TestParseMod(t *testing.T) {
	var m Mod
	var err error

	if _, err = parseMod(``); !(err != nil) {
		t.Errorf("should return err")
	}

	m, err = parseMod(`forge "http://forge.puppetlabs.com"`)
	if !(err != nil) {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar', '0.1.0'`)
	if !(err != nil) {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'foo/bar'`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "foo" && m.opts["git"] == "" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'foo/bar','0.1.2'`)
	if !(err == nil && m.name == "bar" && m.version == "0.1.2" && m.user == "foo" && m.opts["git"] == "" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'foo/bar','0.1'`)
	if !(err == nil && m.name == "bar" && m.version == "0.1" && m.user == "foo" && m.opts["git"] == "" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar', :git => 'a-git-url'`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar',:git=>'a-git-url'`) // no spacing
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar', :git => 'a-git-url', :ref => '0.1.1'`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "0.1.1") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar',:git=>'a-git-url',:ref=>'0.1.1'`) // no spacing
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "0.1.1") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar', :ref => '0.2.1', :git => 'a-url'`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-url" && m.opts["ref"] == "0.2.1") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod "foo/bar", "0.1.2"`)
	if !(err == nil && m.name == "bar" && m.version == "0.1.2" && m.user == "foo" && m.opts["git"] == "" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod "bar", :ref => "0.2.1", :git => "a-url"`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-url" && m.opts["ref"] == "0.2.1") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'garethr/erlang' #, :git => 'git://github.com/garethr/garethr-erlang.git'`)
	if !(err == nil && m.name == "erlang" && m.version == "" && m.user == "garethr" && m.opts["git"] == "" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}
}

func TestIsInclude(t *testing.T) {
	m := isInclude(`include "hello"  `)
	if !(m == "hello") {
		t.Errorf("'%v' should be 'hello'", m)
	}

	m = isInclude(`  include  'world'  `)
	if !(m == "world") {
		t.Errorf("'%v' should be 'world'", m)
	}
}
