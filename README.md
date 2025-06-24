# Life Game Engine - Go Port

A Go port of the LifeJS game engine, using Box2D for physics and Ebiten for rendering.

## Features

- 2D game world with physics simulation
- Shape-based game objects (rectangles, circles, lines)
- Collision detection and response
- Event system for user interactions
- Animation support
- Input handling (keyboard and mouse)
- Sprite and image support

## Installation

\`\`\`bash
go mod init your-game
go get github.com/ByteArena/box2d
go get github.com/hajimehoshi/ebiten/v2
\`\`\`

## Quick Start

\`\`\`go
package main

import (
    "image/color"
    "log"
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

    // Create a shape
    player := life.NewShape(&life.ShapeProps{
        Type:       life.ShapeRectangle,
        X:          100,
        Y:          100,
        Width:      50,
        Height:     50,
        Background: color.RGBA{255, 0, 0, 255},
        Physics:    true,
    })

    world.Register(player)

    // Create and run the game
    game := life.NewGame(world)
    if err := game.Run(); err != nil {
        log.Fatal(err)
    }
}
\`\`\`

## API Reference

### World

The `World` is the main container for your game:

\`\`\`go
world := life.NewWorld(&life.WorldProps{
    Width:      800,
    Height:     600,
    Background: color.RGBA{0, 0, 0, 255},
    HasLimits:  true,
    G:          life.NewVector2(0, 0.5), // Gravity
})
\`\`\`

### Shape

Shapes are the game objects in your world:

\`\`\`go
shape := life.NewShape(&life.ShapeProps{
    Type:       life.ShapeRectangle,
    X:          100,
    Y:          100,
    Width:      50,
    Height:     50,
    Background: color.RGBA{255, 0, 0, 255},
    Physics:    true,
    Speed:      5,
})
\`\`\`

### Movement

\`\`\`go
shape.Move("up")
shape.Move("down") 
shape.Move("left")
shape.Move("right")
shape.MoveTheta(45) // Move at 45 degree angle
shape.Follow(targetShape) // Follow another shape
\`\`\`

### Physics

\`\`\`go
shape.SetVelocity(5, -10)
shape.Jump(15)
shape.SetRotation(45)
\`\`\`

### Events

\`\`\`go
shape.On(life.EventCollision, func(data interface{}) {
    other := data.(*life.Shape)
    log.Printf("Collided with: %s", other.Name)
})

world.OnMouseDown = func(x, y float64) {
    log.Printf("Mouse clicked at: %.2f, %.2f", x, y)
}
\`\`\`

### Animation

\`\`\`go
// Load sprite frames
frames := []*ebiten.Image{frame1, frame2, frame3}

// Create animation
anim := life.NewAnimation(shape, 100*time.Millisecond, true, frames...)
anim.Start()
\`\`\`

## Differences from LifeJS

- Uses Go's strong typing system
- Integrates Box2D for realistic physics simulation
- Uses Ebiten's efficient 2D rendering
- Event system uses Go channels and interfaces
- Memory management handled by Go's garbage collector

## License

MIT License - see LICENSE file for details.
