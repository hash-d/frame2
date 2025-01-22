package f2general

import (
	"log"

	frame2 "github.com/hash-d/frame2/pkg"
)

type Dummy struct {
	Results []error
	round   int

	frame2.Log
}

func (d *Dummy) Validate() error {
	ret := d.Results[d.round%len(d.Results)]
	d.round++
	log.Printf("Dummy run %d", d.round)
	return ret
}
