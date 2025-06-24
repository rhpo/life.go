package life

import (
	"image/color"
	"math"

	"github.com/ByteArena/box2d"
	"github.com/hajimehoshi/ebiten/v2"
)

// Border represents border properties
type Border struct {
	Width      float64
	Background color.Color
	Pattern    PatternType
}

// Shape represents a game object
type Shape struct {
	*EventEmitter

	// Basic properties
	ID            string
	Name          string
	Tag           string
	Type          ShapeType
	X, Y          float64
	Width         float64
	Height        float64
	Radius        float64
	RotationAngle float64
	RotationSpeed float64
	Mass          float64
	Density       float64
	ZIndex        int
	Scale         float64
	Opacity       float64

	// Visual properties
	Pattern    PatternType
	Background color.Color
	Image      *ebiten.Image
	Border     *Border
	Flip       struct{ X, Y bool }

	// Physics properties
	IsBody   bool
	Physics  bool
	Velocity Vector2
	Speed    float64
	Rebound  float64
	Friction float64
	Body     *box2d.B2Body

	// Collision properties
	NoCollisionWith  []string
	CollisionObjects []*Shape
	CacheDirection   string

	// State
	Hovered bool
	Clicked bool

	// Line coordinates (for line shapes)
	LineCoordinates struct{ X1, Y1, X2, Y2 float64 }

	// Callbacks
	OnCollisionFunc       func(*Shape)
	OnFinishCollisionFunc func(*Shape)

	// Reference to world
	world *World
}

// NewShape creates a new shape
func NewShape(props *ShapeProps) *Shape {
	if props == nil {
		props = &ShapeProps{}
	}

	// Set defaults
	if props.Type == "" {
		props.Type = ShapeRectangle
	}
	if props.Name == "" {
		props.Name = RandName()
	}
	if props.Tag == "" {
		props.Tag = "unknown"
	}
	if props.Width == 0 {
		props.Width = 10
	}
	if props.Height == 0 {
		props.Height = 10
	}
	if props.Speed == 0 {
		props.Speed = 3
	}
	if props.Rebound == 0 {
		props.Rebound = 0.7
	}
	if props.Friction == 0 {
		props.Friction = 0.5
	}
	if props.Scale == 0 {
		props.Scale = 1
	}
	if props.Opacity == 0 {
		props.Opacity = 1
	}
	if props.Background == nil {
		props.Background = color.RGBA{0, 0, 0, 255}
	}
	if props.Pattern == "" {
		props.Pattern = PatternColor
	}

	shape := &Shape{
		EventEmitter:          NewEventEmitter(),
		ID:                    ID(),
		Name:                  props.Name,
		Tag:                   props.Tag,
		Type:                  props.Type,
		X:                     props.X,
		Y:                     props.Y,
		Width:                 props.Width,
		Height:                props.Height,
		Radius:                props.Radius,
		RotationAngle:         props.Rotation,
		ZIndex:                props.ZIndex,
		Scale:                 props.Scale,
		Opacity:               props.Opacity,
		Pattern:               props.Pattern,
		Background:            props.Background,
		Image:                 props.Image,
		Border:                props.Border,
		IsBody:                props.IsBody,
		Physics:               props.Physics,
		Velocity:              props.Velocity,
		Speed:                 props.Speed,
		Rebound:               props.Rebound,
		Friction:              props.Friction,
		NoCollisionWith:       props.NoCollisionWith,
		OnCollisionFunc:       props.OnCollisionFunc,
		OnFinishCollisionFunc: props.OnFinishCollisionFunc,
		LineCoordinates:       props.LineCoordinates,
		Flip:                  props.Flip,
	}

	if props.Radius > 0 {
		shape.Width = props.Radius
		shape.Height = props.Radius
	}

	return shape
}

// ShapeProps contains properties for creating a shape
type ShapeProps struct {
	Type                  ShapeType
	X, Y                  float64
	Width, Height         float64
	Radius                float64
	ZIndex                int
	IsBody                bool
	Pattern               PatternType
	Background            color.Color
	Image                 *ebiten.Image
	Name                  string
	Rotation              float64
	Tag                   string
	OnCollisionFunc       func(*Shape)
	OnFinishCollisionFunc func(*Shape)
	Physics               bool
	Rebound               float64
	Friction              float64
	Speed                 float64
	Velocity              Vector2
	Border                *Border
	Flip                  struct{ X, Y bool }
	Opacity               float64
	NoCollisionWith       []string
	LineCoordinates       struct{ X1, Y1, X2, Y2 float64 }
	Scale                 float64
}

func (s *Shape) UpdatePhysicsInfo() {
	center := s.Body.GetPosition()
	s.X = center.X - s.Width/2
	s.Y = center.Y - s.Height/2
	s.RotationAngle = s.Body.GetAngle()
	s.RotationSpeed = s.Body.GetAngularVelocity()
	s.Velocity = NewVector2(s.Body.GetLinearVelocity().X, s.Body.GetLinearVelocity().Y)
	if s.Body.GetMass() > 0 {
		s.Mass = s.Body.GetMass()
	} else {
		s.Mass = 1
	}
}

func (s *Shape) SetX(x float64) {
	s.X = x
	centerX := x + s.Width/2
	s.Body.SetTransform(box2d.MakeB2Vec2(centerX, s.Y+s.Height/2), s.RotationAngle)
}

func (s *Shape) SetY(y float64) {
	s.Y = y
	centerY := y + s.Height/2
	s.Body.SetTransform(box2d.MakeB2Vec2(s.X+s.Width/2, centerY), s.RotationAngle)
}

func (s *Shape) SetPosition(x, y float64) {
	s.X = x
	s.Y = y
	centerX := x + s.Width/2
	centerY := y + s.Height/2
	s.Body.SetTransform(box2d.MakeB2Vec2(centerX, centerY), s.RotationAngle)
}

// SetRotation sets the rotation angle of the shape
func (s *Shape) SetRotation(angle float64) {
	s.RotationAngle = angle

	s.Body.SetTransform(box2d.MakeB2Vec2(s.X, s.Y), angle*Deg)
}

// SetScale sets the scale of the shape
func (s *Shape) SetScale(scale float64) {
	s.Scale = scale

	// Box2D does not support scaling directly, so we need to recreate the body with the new scale
	// This is a simplified approach; in a real scenario, you would need to destroy the old body and create a new one
	s.Body.SetTransform(box2d.MakeB2Vec2(s.X, s.Y), s.RotationAngle*Deg)
}

// Draw renders the shape
func (s *Shape) Draw(screen *ebiten.Image) {
	s.UpdatePhysicsInfo()

	if s.Opacity <= 0 {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Step 1: Move origin to the center of the shape (relative to image)
	op.GeoM.Translate(-s.Width/2, -s.Height/2)

	// Step 2: Flip if needed
	scaleX, scaleY := 1.0, 1.0
	if s.Image != nil {
		imgWidth, imgHeight := s.Image.Size()

		scaleX = float64(imgWidth) / s.Width * 0.1
		scaleY = float64(imgHeight) / s.Height * 0.2
	}

	if s.Flip.X {
		scaleX *= -1
	}
	if s.Flip.Y {
		scaleY *= -1
	}

	op.GeoM.Scale(s.Scale, s.Scale)
	op.GeoM.Scale(scaleX, scaleY)

	op.GeoM.Rotate(s.RotationAngle)

	op.GeoM.Translate(s.X+s.Width/2, s.Y+s.Height/2)

	if s.Border != nil && s.Border.Width > 0 {
		s.drawBorder(screen)
	}

	switch s.Type {
	case ShapeRectangle:
		s.drawRectangle(screen, op)
	case ShapeSquare:
		s.drawSquare(screen, op)
	case ShapeCircle:
		s.drawCircle(screen, op)
	case ShapeDot:
		s.drawDot(screen, op)
	case ShapeLine:
		s.drawLine(screen, op)
	}

}

func (s *Shape) drawRectangle(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	switch s.Pattern {
	case PatternColor:
		// Create a colored rectangle
		img := ebiten.NewImage(int(s.Width), int(s.Height))
		img.Fill(s.Background)
		screen.DrawImage(img, op)
	case PatternImage:
		// make a resized image from s.Image
		if s.Image != nil {
			screen.DrawImage(s.Image, op)
		}
	}
}

func (s *Shape) drawSquare(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	// For squares, ensure width and height are equal
	size := s.Width
	if s.Height > s.Width {
		size = s.Height
	}

	switch s.Pattern {
	case PatternColor:
		// Create a colored square
		img := ebiten.NewImage(int(size), int(size))
		img.Fill(s.Background)
		screen.DrawImage(img, op)
	case PatternImage:
		if s.Image != nil {
			screen.DrawImage(s.Image, op)
		}
	}
}

func (s *Shape) drawCircle(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	// Create a circle image
	size := int(s.Radius * 2)
	img := ebiten.NewImage(size, size)

	// Simple circle drawing (can be optimized)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - s.Radius
			dy := float64(y) - s.Radius
			if dx*dx+dy*dy <= s.Radius*s.Radius {
				img.Set(x, y, s.Background)
			}
		}
	}

	screen.DrawImage(img, op)
}

func (s *Shape) drawLine(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	// Line drawing implementation would go here
	// This is a simplified version
	img := ebiten.NewImage(int(s.Width), int(s.Height))
	img.Fill(s.Background)
	screen.DrawImage(img, op)
}

func (s *Shape) drawDot(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	s.drawCircle(screen, op)
}

func (s *Shape) drawBorder(screen *ebiten.Image) {
	// Border drawing implementation
	borderImg := ebiten.NewImage(int(s.Width+s.Border.Width*2), int(s.Height+s.Border.Width*2))
	borderImg.Fill(s.Border.Background)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.X-s.Border.Width, s.Y-s.Border.Width)
	screen.DrawImage(borderImg, op)
}

func (s *Shape) MoveTheta(angle float64, optionalSpeed ...float64) {
	speed := s.Speed
	if len(optionalSpeed) > 0 {
		speed = optionalSpeed[0]
	}

	// Apply force in the specified direction
	impulse := box2d.MakeB2Vec2(
		math.Cos(angle*Deg)*speed,
		math.Sin(angle*Deg)*speed,
	)
	s.Body.ApplyLinearImpulse(impulse, s.Body.GetWorldCenter(), true)

	s.X += math.Cos(angle*Deg) * speed
	s.Y += math.Sin(angle*Deg) * speed
}

func (s *Shape) Follow(target *Shape) {
	dx := target.X - s.X
	dy := target.Y - s.Y
	angle := math.Atan2(dy, dx)

	s.SetX(math.Cos(angle) * s.Speed)
	s.SetY(s.Y + math.Sin(angle)*s.Speed)

	impulse := box2d.MakeB2Vec2(
		math.Cos(angle)*s.Speed,
		math.Sin(angle)*s.Speed,
	)
	s.Body.ApplyLinearImpulse(impulse, s.Body.GetWorldCenter(), true)
}

// Physics methods
func (s *Shape) SetVelocity(x, y float64) {
	s.Body.SetLinearVelocity(box2d.MakeB2Vec2(x, y))
}

// setXVelocity sets the X component of the velocity
func (s *Shape) SetXVelocity(x float64) {

	vel := s.Body.GetLinearVelocity()
	s.Body.SetLinearVelocity(box2d.MakeB2Vec2(x, vel.Y))

}

// setYVelocity sets the Y component of the velocity
func (s *Shape) SetYVelocity(y float64) {

	vel := s.Body.GetLinearVelocity()
	s.Body.SetLinearVelocity(box2d.MakeB2Vec2(vel.X, y))

}

func (s *Shape) Jump(howHigh float64) {
	s.SetYVelocity(-howHigh)
}

func (s *Shape) Rotate(angle float64) {
	s.Body.SetTransform(s.Body.GetPosition(), s.Body.GetAngle()+angle*Deg)
}

// Collision methods
func (s *Shape) IsCollidingWith(target *Shape) bool {
	// Simple AABB collision detection for rectangles and circles
	if s == nil || target == nil {
		return false
	}

	switch s.Type {
	case ShapeRectangle, ShapeSquare:
		sx1, sy1 := s.X, s.Y
		sx2, sy2 := s.X+s.Width, s.Y+s.Height
		tx1, ty1 := target.X, target.Y
		tx2, ty2 := target.X+target.Width, target.Y+target.Height
		return sx1 < tx2 && sx2 > tx1 && sy1 < ty2 && sy2 > ty1
	case ShapeCircle, ShapeDot:
		dx := (s.X + s.Radius) - (target.X + target.Radius)
		dy := (s.Y + s.Radius) - (target.Y + target.Radius)
		distance := math.Sqrt(dx*dx + dy*dy)
		return distance < (s.Radius + target.Radius)
	default:
		// Fallback: bounding box
		sx1, sy1 := s.X, s.Y
		sx2, sy2 := s.X+s.Width, s.Y+s.Height
		tx1, ty1 := target.X, target.Y
		tx2, ty2 := target.X+target.Width, target.Y+target.Height
		return sx1 < tx2 && sx2 > tx1 && sy1 < ty2 && sy2 > ty1
	}
}
func (s *Shape) IsCollidingWithAngle(target *Shape) bool {
	if s == nil || target == nil {
		return false
	}

	// For rectangles with rotation, use Separating Axis Theorem (SAT)
	if (s.Type == ShapeRectangle || s.Type == ShapeSquare) &&
		(target.Type == ShapeRectangle || target.Type == ShapeSquare) &&
		(s.RotationAngle != 0 || target.RotationAngle != 0) {

		// Get corners of both shapes
		sCorners := getRotatedRectCorners(s.X, s.Y, s.Width, s.Height, s.RotationAngle)
		tCorners := getRotatedRectCorners(target.X, target.Y, target.Width, target.Height, target.RotationAngle)

		return polygonsOverlap(sCorners, tCorners)
	}

	// Fallback to AABB/circle as before
	return s.IsCollidingWith(target)
}

// Helper: get rectangle corners after rotation
func getRotatedRectCorners(x, y, w, h, angle float64) [4][2]float64 {
	cx := x + w/2
	cy := y + h/2
	rad := angle * Deg

	cosA := math.Cos(rad)
	sinA := math.Sin(rad)

	// Rectangle corners relative to center
	pts := [4][2]float64{
		{-w / 2, -h / 2},
		{w / 2, -h / 2},
		{w / 2, h / 2},
		{-w / 2, h / 2},
	}
	for i := range pts {
		px := pts[i][0]
		py := pts[i][1]
		pts[i][0] = cx + px*cosA - py*sinA
		pts[i][1] = cy + px*sinA + py*cosA
	}
	return pts
}

// Helper: SAT polygon overlap test
func polygonsOverlap(a, b [4][2]float64) bool {
	polys := [][][2]float64{a[:], b[:]}
	for _, poly := range polys {
		for i := 0; i < 4; i++ {
			j := (i + 1) % 4
			edgeX := poly[j][0] - poly[i][0]
			edgeY := poly[j][1] - poly[i][1]
			// Perpendicular axis
			axisX := -edgeY
			axisY := edgeX

			// Project both polygons onto axis
			var minA, maxA, minB, maxB float64
			for k, pt := range a {
				proj := pt[0]*axisX + pt[1]*axisY
				if k == 0 || proj < minA {
					minA = proj
				}
				if k == 0 || proj > maxA {
					maxA = proj
				}
			}
			for k, pt := range b {
				proj := pt[0]*axisX + pt[1]*axisY
				if k == 0 || proj < minB {
					minB = proj
				}
				if k == 0 || proj > maxB {
					maxB = proj
				}
			}
			// If projections do not overlap, no collision
			if maxA < minB || maxB < minA {
				return false
			}
		}
	}
	return true
}
func (s *Shape) CollideWith(object *Shape) {
	s.CollisionObjects = append(s.CollisionObjects, object)
}

func (s *Shape) FinishCollideWith(object *Shape) {
	for i, obj := range s.CollisionObjects {
		if obj.ID == object.ID {
			s.CollisionObjects = append(s.CollisionObjects[:i], s.CollisionObjects[i+1:]...)
			break
		}
	}
}

func (s *Shape) CCHas(target *Shape) bool {
	for _, obj := range s.CollisionObjects {
		if obj.ID == target.ID {
			return true
		}
	}
	return false
}

// Utility methods
func (s *Shape) Remove() {
	if s.world != nil {
		s.world.Unregister(s)
	}
}

func (s *Shape) IsOutOfMap() bool {
	if s.world == nil {
		return false
	}
	return s.X < 0 || s.X > float64(s.world.Width) || s.Y < 0 || s.Y > float64(s.world.Height)
}

func (s *Shape) SetProps(props map[string]interface{}) {
	// Reflection-based property setting could be implemented here
	// For now, this is a placeholder
}

func (s *Shape) Get(property string) interface{} {
	// Property getter implementation
	return nil
}

func (s *Shape) Set(property string, value interface{}) {
	// Property setter implementation
}
