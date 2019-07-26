package redactif_test

import (
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/bradleyjkemp/redactif"
	"strings"
	"testing"
)

type untagged struct {
	a string
	b int
}

type substruct struct {
	a       string `redactif:"a"`
	notB    int    `redactif:"!b"`
	aOrNotB bool   `redactif:"a,!b"`
}

type tagged struct {
	a              string    `redactif:"a"`
	notB           int       `redactif:"!b"`
	aOrNotB        bool      `redactif:"a,!b"`
	recurse        substruct `redactif:"recurse"`
	pointerRecurse *tagged
}

func TestIgnoresUntaggedFields(t *testing.T) {
	val := &untagged{
		a: "hello",
		b: 1337,
	}
	redactif.Redact(val, "!user", "snapshot")
	cupaloy.SnapshotT(t, val)
}

func TestBasicTaggedFields(t *testing.T) {
	cases := []string{"", "a", "a,b", "b"}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			sub := &tagged{
				a:       "world",
				notB:    16353,
				aOrNotB: true,
			}
			subsub := substruct{
				a:       "universe",
				notB:    8161,
				aOrNotB: true,
			}
			val := &tagged{
				a:              "hello",
				notB:           1337,
				aOrNotB:        true,
				recurse:        subsub,
				pointerRecurse: sub,
			}
			sub.pointerRecurse = val
			cupaloy.SnapshotT(t, redactif.Redact(val, strings.Split(tc, ",")...))
		})
	}
}
