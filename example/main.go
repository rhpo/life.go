package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yourname/life"
)

func main() {
	// Create a world
	world := life.NewWorld(&life.WorldProps{
		Width:      800,
		Height:     600,
		Background: color.RGBA{50, 50, 50, 255},
		HasLimits:  true,
		G:          life.NewVector2(0, 0.5), // Gravity
	})

	// Create a player shape
	player := life.NewShape(&life.ShapeProps{
		Type:       life.ShapeRectangle,
		X:          100,
		Y:          100,
		Width:      50,
		Height:     50,
		Background: color.RGBA{255, 0, 0, 255},
		Name:       "player",
		Tag:        "player",
		Physics:    true,
		Speed:      5,
	})

	// Add collision handler
	player.OnCollisionFunc = func(self, other *life.Shape) {
		if other.Tag == "border" {
			log.Printf("Player hit border!")
		}
	}

	// Register the player with the world
	world.Register(player)

	// Create some obstacles
	for i := 0; i < 5; i++ {
		obstacle := life.NewShape(&life.ShapeProps{
			Type:       life.ShapeCircle,
			X:          float64(200 + i*100),
			Y:          300,
			Radius:     25,
			Background: color.RGBA{0, 255, 0, 255},
			Name:       "obstacle",
			Tag:        "obstacle",
			Physics:    false,
		})
		world.Register(obstacle)
	}

	// Handle input
	world.OnMouseDown = func(x, y float64) {
		log.Printf("Mouse clicked at: %.2f, %.2f", x, y)
	}

	// Create and run the game
	game := life.NewGame(world)
	
	// Add a simple game loop for player movement
	go func() {
		for {
			if world.IsKeyPressed(ebiten.KeyArrowLeft) || world.IsKeyPressed(ebiten.KeyA) {
				player.Move("left")
			}
			if world.IsKeyPressed(ebiten.KeyArrowRight) || world.IsKeyPressed(ebiten.KeyD) {
				player.Move("right")
			}
			if world.IsKeyPressed(ebiten.KeyArrowUp) || world.IsKeyPressed(ebiten.KeyW) {
				player.Move("up")
			}
			if world.IsKeyPressed(ebiten.KeyArrowDown) || world.IsKeyPressed(ebiten.KeyS) {
				player.Move("down")
			}
			if world.IsKeyPressed(ebiten.KeySpace) {
				player.Jump(10)
			}
		}
	}()

	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}
