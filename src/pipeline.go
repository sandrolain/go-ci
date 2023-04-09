package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type RefType int

const (
	RefTypeCommit RefType = iota
	RefTypeBranch
	RefTypeTag
)

type Pipeline struct {
	URL     string
	RefType RefType
	Ref     string
}

func (p *Pipeline) Clone(dest string) (err error) {
	repo, err := git.PlainClone(dest, false, &git.CloneOptions{
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
