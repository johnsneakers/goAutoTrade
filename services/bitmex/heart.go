package bitmex

import "time"

const TFREQ = time.Second * 6

type Heart struct {
	cnt int
	timer *time.Timer
}

func NewHeart() *Heart {
	return &Heart{
		0,
		time.NewTimer(TFREQ),
	}
}


func (c *Heart) Reset() {
	c.cnt = 0
	c.timer.Reset(TFREQ)
}

