package life

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Game implements ebiten.Game interface
type Game struct {
	world *World
}

// NewGame creates a new game instance
func NewGame(world *World) *Game {
	return &Game{
		world: world,
	}
}

// Update is called every frame (60 FPS by default)
func (g *Game) Update() error {
	return g.world.Update()
}

// Draw is called every frame to render the game
func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
}

// Layout returns the game's screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.world.Width, g.world.Height
}

// Run starts the game
func (g *Game) Run() error {
	ebiten.SetWindowSize(g.world.Width, g.world.Height)
	ebiten.SetWindowTitle("Life Game Engine")
	return ebiten.RunGame(g)
}
