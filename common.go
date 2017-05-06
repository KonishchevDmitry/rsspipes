package rsspipes

import "github.com/KonishchevDmitry/rsspipes/util"

type temporary interface {
	Temporary() bool
}

var log = util.MustGetLogger("rsspipes")
