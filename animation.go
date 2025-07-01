package life

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
	target       *Shape
	frames       []*ebiten.Image
	currentFrame int
	speed        time.Duration
	loop         bool
	isPlaying    bool
	ticker       *time.Ticker
	stopCh       chan bool
	onFinish     func(*Shape)
}

func NewAnimation(target *Shape, speed time.Duration, loop bool, frames ...*ebiten.Image) *Animation {
	if speed == 0 {
		speed = 100 * time.Millisecond
	}

	anim := &Animation{
		target:       target,
		frames:       frames,
		currentFrame: 0,
		speed:        speed,
		loop:         loop,
		isPlaying:    false,
		stopCh:       make(chan bool),
		onFinish:     func(*Shape) {},
	}

	return anim
}

func (a *Animation) Start() *Animation {
	if a.isPlaying {
		return a
	}

	a.isPlaying = true
	a.ticker = time.NewTicker(a.speed)

	go func() {
		for {
			select {
			case <-a.ticker.C:
				if !a.isPlaying {
					continue
				}

				a.currentFrame++
				if a.currentFrame >= len(a.frames) {
					a.currentFrame = 0
					if !a.loop {
						a.isPlaying = false
						a.onFinish(a.target)
						return
					}
				}

				if a.currentFrame < len(a.frames) {
					a.target.Image = a.frames[a.currentFrame]
				}

			case <-a.stopCh:
				return
			}
		}
	}()

	return a
}

func (a *Animation) Stop() *Animation {
	a.isPlaying = false
	if a.ticker != nil {
		a.ticker.Stop()
	}
	a.stopCh <- true
	return a
}

func (a *Animation) OnFinish(callback func(*Shape)) *Animation {
	a.onFinish = callback
	return a
}

func (a *Animation) IsPlaying() bool {
	return a.isPlaying
}
