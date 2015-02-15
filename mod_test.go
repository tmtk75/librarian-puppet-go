package main

import (
	"io"
	"strings"
	"testing"
)

func r(s string) io.Reader {
	return strings.NewReader(s)
}

func TestParsePuppetfile(t *testing.T) {
	var mods []Mod

	mods = parsePuppetfile(r(``))
	if !(len(mods) == 0) {
		t.Errorf("should be 0")
	}

	mods = parsePuppetfile(r(`
#
mod 'foo/bar'
mod 'foo/bar', '0.1.2'

mod 'foo', :git => 'git-url'
mod 'foo', :git => 'git-url', :ref => 'v0.1.1'  #
`))
	if !(len(mods) == 4) {
		t.Errorf("should be 0")
	}
}

func TestParseMod(t *testing.T) {
	var m Mod
	var err error

	if _, err = parseMod(``); !(err != nil) {
		t.Errorf("should return err")
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

	m, err = parseMod(`mod 'bar',:git=>'a-git-url'`) // mo spacing
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar', :git => 'a-git-url', :ref => '0.1.1'`)
	if !(err == nil && m.name == "bar" && m.version == "" && m.user == "" && m.opts["git"] == "a-git-url" && m.opts["ref"] == "0.1.1") {
		t.Errorf("%v %v", m, err)
	}

	m, err = parseMod(`mod 'bar',:git=>'a-git-url',:ref=>'0.1.1'`) // mo spacing
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
}
