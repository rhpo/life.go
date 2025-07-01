package life

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	world *World
}

func NewGame(world *World) *Game {
	return &Game{
		world: world,
	}
}

func (g *Game) Update() error {
	return g.world.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.world.Draw(screen)
	g.world.Render(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.world.Width, g.world.Height
}

func (g *Game) Run() error {
	ebiten.SetWindowSize(g.world.Width, g.world.Height)
	ebiten.SetWindowTitle(g.world.Title)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if g.world.Init != nil {
		g.world.SelectLevel(0)
	}

	return ebiten.RunGame(g)
}
