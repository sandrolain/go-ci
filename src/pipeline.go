package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"
)

type RefType int

const (
	RefTypeCommit RefType = iota
	RefTypeBranch
	RefTypeTag
)

type Pipeline struct {
	URL      string
	RefType  RefType
	Ref      string
	CI       PipelineCI
	workpath string
}

func (p *Pipeline) SetupWorkpath() (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	temp := fmt.Sprintf("%s/temp", cwd)

	t := time.Now()
	ts := t.Format("20060102150405")
	dest := fmt.Sprintf("%s/p_%s", temp, ts)

	err = os.MkdirAll(dest, 0700)
	if err != nil {
		return
	}
	p.workpath = dest
	return
}

func (p *Pipeline) GetWorkpath() string {
	return p.workpath
}

func (p *Pipeline) Cleanup() error {
	return os.RemoveAll(p.workpath)
}

func (p *Pipeline) Clone() (err error) {
	repo, err := git.PlainClone(p.workpath, false, &git.CloneOptions{
		URL:               p.URL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})
	if err != nil {
		err = fmt.Errorf("cannot clone repo: %w", err)
		return
	}
	w, err := repo.Worktree()
	if err != nil {
		err = fmt.Errorf("cannot obtain worktree: %w", err)
		return
	}
	if p.RefType == RefTypeCommit {
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(p.Ref),
		})
		if err != nil {
			err = fmt.Errorf("cannot checkout hash: %w", err)
			return
		}
	} else {
		var branch plumbing.ReferenceName
		switch p.RefType {
		case RefTypeBranch:
			branch = plumbing.NewBranchReferenceName(p.Ref)
		case RefTypeTag:
			branch = plumbing.NewTagReferenceName(p.Ref)
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: branch,
		})
		if err != nil {
			err = fmt.Errorf("cannot checkout branch: %w", err)
			return
		}
	}

	return
}

func (p *Pipeline) GetCI() (*PipelineCI, error) {
	ciPath := path.Join(p.workpath, "go-ci.yaml")
	data, err := os.ReadFile(ciPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &p.CI, nil
		}
		return nil, err
	}
	var ci PipelineCI
	err = yaml.Unmarshal(data, &ci)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

type PipelineCI struct {
	Image   string
	Tag     string
	Workdir string
	Volumes map[string]string
	Steps   []Step
}

func (p *PipelineCI) GetWorkdir() string {
	if p.Workdir == "" {
		return "/workdir"
	}
	return p.Workdir
}
