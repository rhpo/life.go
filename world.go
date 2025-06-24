package life

import (
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ByteArena/box2d"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ContactListener struct {
	box2d.B2ContactListenerInterface
	world *World
}

func (cl *ContactListener) PreSolve(contact box2d.B2ContactInterface, oldManifold box2d.B2Manifold) {
	// No-op, required by Box2D interface
}

func (cl *ContactListener) PostSolve(contact box2d.B2ContactInterface, impulse *box2d.B2ContactImpulse) {
	// No-op, required by Box2D interface
}

func (cl *ContactListener) BeginContact(contact box2d.B2ContactInterface) {
	cl.world.mutex.Lock() //  Lock the world mutex to ensure thread safety
	defer cl.world.mutex.Unlock()

	// Handle contact begin
	fixtureA := contact.GetFixtureA()
	fixtureB := contact.GetFixtureB()

	if fixtureA == nil || fixtureB == nil {
		return
	}

	bodyA := fixtureA.GetBody()
	bodyB := fixtureB.GetBody()

	// get from the world's objects, the bodies that are involved in the contact
	var shapeA, shapeB *Shape
	for _, obj := range cl.world.Objects {
		if obj.Body == bodyA {
			shapeA = obj
		} else if obj.Body == bodyB {
			shapeB = obj
		}
	}

	if shapeA == nil || shapeB == nil {
		return
	}

	// Emit collision event
	cl.world.Emit("collision", map[string]interface{}{
		"shapeA": shapeA,
		"shapeB": shapeB,
	})

	shapeA.CollideWith(shapeB)
	shapeB.CollideWith(shapeA)

	if shapeA.OnCollisionFunc != nil {
		shapeA.OnCollisionFunc(shapeB)
	}

	if shapeB.OnCollisionFunc != nil {
		shapeB.OnCollisionFunc(shapeA)
	}
}

func (cl *ContactListener) EndContact(contact box2d.B2ContactInterface) {
	cl.world.mutex.Lock() // Lock the world mutex to ensure thread safety
	defer cl.world.mutex.Unlock()

	// Handle contact end
	fixtureA := contact.GetFixtureA()
	fixtureB := contact.GetFixtureB()

	if fixtureA == nil || fixtureB == nil {
		return
	}

	bodyA := fixtureA.GetBody()
	bodyB := fixtureB.GetBody()

	// get from the world's objects, the bodies that are involved in the contact
	var shapeA, shapeB *Shape
	for _, obj := range cl.world.Objects {
		if obj.Body == bodyA {
			shapeA = obj
		} else if obj.Body == bodyB {
			shapeB = obj
		}
	}

	if shapeA == nil || shapeB == nil {
		return
	}

	// Emit collision end event
	cl.world.Emit("collision-end", map[string]interface{}{
		"shapeA": shapeA,
		"shapeB": shapeB,
	})

	shapeA.FinishCollideWith(shapeB)
	shapeB.FinishCollideWith(shapeA)

}

// World represents the game world
type World struct {
	*EventEmitter

	// Display properties
	Width  int
	Height int

	// Physics
	PhysicsWorld    *box2d.B2World
	contactListener *ContactListener
	G               Vector2 // Gravity

	Screen *ebiten.Image // Screen to draw on

	Logic      *GameLoop
	lastUpdate time.Time

	// Visual properties
	Pattern    PatternType
	Background color.Color
	Border     *Border

	// Game objects
	Objects []*Shape
	mutex   sync.RWMutex

	// Input
	Mouse struct {
		X, Y                          float64
		IsLeftClicked, IsRightClicked bool
		IsMiddleClicked               bool
	}
	Keys      map[ebiten.Key]bool
	keysMutex sync.RWMutex // Add mutex for keys map

	// State
	HasLimits bool
	Paused    bool
	Cursor    CursorType

	// Callbacks
	OnMouseDown func(x, y float64)
	OnMouseUp   func(x, y float64)
	OnMouseMove func(x, y float64)
}

// WorldProps contains properties for creating a world
type WorldProps struct {
	Width      int
	Height     int
	G          Vector2
	Pattern    PatternType
	Background color.Color
	HasLimits  bool
	Border     *Border
	Paused     bool
	Cursor     CursorType
}

// NewWorld creates a new world
func NewWorld(props *WorldProps) *World {
	if props == nil {
		props = &WorldProps{}
	}

	// Set defaults
	if props.Width == 0 {
		props.Width = 800
	}
	if props.Height == 0 {
		props.Height = 600
	}
	if props.Background == nil {
		props.Background = color.RGBA{0, 0, 0, 255}
	}
	if props.Pattern == "" {
		props.Pattern = PatternColor
	}
	if props.Cursor == "" {
		props.Cursor = CursorDefault
	}

	// Create Box2D world
	contactListener := ContactListener{}

	gravity := box2d.MakeB2Vec2(props.G.X, props.G.Y)
	physicsWorld := box2d.MakeB2World(gravity)
	physicsWorld.SetContactListener(&contactListener)

	world := &World{
		EventEmitter:    NewEventEmitter(),
		contactListener: &contactListener,
		PhysicsWorld:    &physicsWorld,
		Width:           props.Width,
		Height:          props.Height,
		G:               props.G,
		Logic:           nil, // Set later
		Pattern:         props.Pattern,
		Background:      props.Background,
		Border:          props.Border,
		Paused:          props.Paused,
		Cursor:          props.Cursor,
		Keys:            make(map[ebiten.Key]bool),
		lastUpdate:      time.Now(),
	}

	contactListener.world = world

	return world
}

func (w *World) CreateBorders() {
	borderWidth := 8.0
	if w.Border != nil && w.Border.Width > 0 {
		borderWidth = w.Border.Width
	}

	var borderColor color.Color = color.RGBA{0, 0, 0, 255}
	if w.Border != nil && w.Border.Background != nil {
		borderColor = w.Border.Background
	}

	// Top border - static physics body
	topBorder := NewShape(&ShapeProps{
		Type:       ShapeRectangle,
		X:          0,
		Y:          0,
		Width:      float64(w.Width),
		Height:     borderWidth,
		Background: borderColor,
		Tag:        "border",
		Name:       "borderX",
		Physics:    true,
		IsBody:     true,
	})

	// Bottom border - static physics body
	bottomBorder := NewShape(&ShapeProps{
		Type:       ShapeRectangle,
		X:          0,
		Y:          float64(w.Height) - borderWidth,
		Width:      float64(w.Width),
		Height:     borderWidth,
		Background: borderColor,
		Tag:        "border",
		Name:       "borderXW",
		Physics:    true,
		IsBody:     true,
	})

	// Left border - static physics body
	leftBorder := NewShape(&ShapeProps{
		Type:       ShapeRectangle,
		X:          0,
		Y:          0,
		Width:      borderWidth,
		Height:     float64(w.Height),
		Background: borderColor,
		Tag:        "border",
		Name:       "borderY",
		Physics:    true,
		IsBody:     true,
	})

	// Right border - static physics body
	rightBorder := NewShape(&ShapeProps{
		Type:       ShapeRectangle,
		X:          float64(w.Width) - borderWidth,
		Y:          0,
		Width:      borderWidth,
		Height:     float64(w.Height),
		Background: borderColor,
		Tag:        "border",
		Name:       "borderYW",
		Physics:    true,
		IsBody:     true,
	})

	w.Register(topBorder)
	w.Register(bottomBorder)
	w.Register(leftBorder)
	w.Register(rightBorder)
}

// Register adds an object to the world
func (w *World) Register(object *Shape) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	object.world = w
	w.Objects = append(w.Objects, object)

	w.createPhysicsBody(object)
}

// Unregister removes an object from the world
func (w *World) Unregister(object *Shape) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for i, obj := range w.Objects {
		if obj.ID == object.ID {
			// Destroy Box2D body if it exists
			if obj.Body != nil {
				w.PhysicsWorld.DestroyBody(obj.Body)
			}

			w.Objects = append(w.Objects[:i], w.Objects[i+1:]...)
			break
		}
	}
}

func (w *World) createPhysicsBody(object *Shape) {
	// Create body definition
	bodyDef := box2d.MakeB2BodyDef()

	if !object.IsBody || object.Tag == "border" {
		bodyDef.Type = box2d.B2BodyType.B2_staticBody // Static bodies for borders
	} else {
		bodyDef.Type = box2d.B2BodyType.B2_dynamicBody // Dynamic bodies for other objects
	}

	if !object.Physics {
		bodyDef.GravityScale = 0
	}

	centerX := object.X + object.Width/2
	centerY := object.Y + object.Height/2
	bodyDef.Position.Set(centerX, centerY)

	// Create body
	body := w.PhysicsWorld.CreateBody(&bodyDef)

	// Create and assign shape based on object type
	// set width and height for rectangle shapes
	var shape box2d.B2ShapeInterface
	switch object.Type {
	case ShapeCircle:
		circleShape := box2d.MakeB2CircleShape()
		circleShape.SetRadius(object.Radius)
		shape = &circleShape
	default: // Rectangle
		boxShape := box2d.MakeB2PolygonShape()
		// Set box shape dimensions based on object width and height
		if object.Width <= 0 || object.Height <= 0 {
			panic("Width and Height must be greater than 0 for rectangle shapes")
		}

		boxShape.SetAsBox(object.Width/2, object.Height/2)
		shape = &boxShape
	}

	// Create fixture with shape and density
	density := 1.0
	if object.Tag == "border" {
		density = 0.0 // Static bodies should have 0 density
	}
	fixture := body.CreateFixture(shape, density)

	// Set additional properties
	fixture.SetFriction(object.Friction)
	fixture.SetRestitution(object.Rebound)

	object.Body = body
}

// Update updates the world state
func (w *World) Update() error {
	if w.Paused {
		return nil
	}

	// Calculate deltaTime
	now := time.Now()
	var deltaTime float64
	if !w.lastUpdate.IsZero() {
		deltaTime = now.Sub(w.lastUpdate).Seconds()
	} else {
		deltaTime = 1.0 / 60.0 // default on first frame
	}
	w.lastUpdate = now

	if w.Logic != nil {
		(*w.Logic)(LoopData{
			Time:  now,
			Delta: deltaTime,
		})
	}

	// Update input
	w.updateInput()

	// Step physics simulation with real deltaTime
	velocityIterations := 1
	positionIterations := 1
	w.PhysicsWorld.Step(deltaTime, velocityIterations, positionIterations)

	return nil
}

func (w *World) updateInput() {
	// Update mouse position
	x, y := ebiten.CursorPosition()
	w.Mouse.X = float64(x)
	w.Mouse.Y = float64(y)

	// Update mouse buttons
	w.Mouse.IsLeftClicked = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	w.Mouse.IsRightClicked = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	w.Mouse.IsMiddleClicked = ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)

	// Update keyboard with proper synchronization
	w.keysMutex.Lock()
	for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
		w.Keys[key] = ebiten.IsKeyPressed(key)
	}
	w.keysMutex.Unlock()

	// Handle mouse events
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		w.handleMouseDown(w.Mouse.X, w.Mouse.Y)
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		w.handleMouseUp(w.Mouse.X, w.Mouse.Y)
	}
}

func (w *World) handleMouseDown(x, y float64) {
	hoveredObjects := w.HoveredObjects()
	for _, obj := range hoveredObjects {
		if !obj.Clicked {
			obj.Clicked = true
			obj.Emit(EventMouseDown, map[string]float64{"x": x, "y": y})
		}
	}

	if w.OnMouseDown != nil {
		w.OnMouseDown(x, y)
	}
}

func (w *World) handleMouseUp(x, y float64) {
	hoveredObjects := w.HoveredObjects()
	for _, obj := range hoveredObjects {
		obj.Emit(EventMouseUp, map[string]float64{"x": x, "y": y})
		obj.Emit(EventClick, map[string]float64{"x": x, "y": y})
		if obj.Clicked {
			obj.Clicked = false
		}
	}

	if w.OnMouseUp != nil {
		w.OnMouseUp(x, y)
	}
}

// Draw renders the world
func (w *World) Draw(screen *ebiten.Image) {

	if w.Screen != screen {
		w.Screen = screen
	}

	// Clear screen with background
	screen.Fill(w.Background)

	// Sort objects by Z-index
	w.mutex.RLock()
	objects := make([]*Shape, len(w.Objects))
	copy(objects, w.Objects)
	w.mutex.RUnlock()

	sort.Slice(objects, func(i, j int) bool {
		if objects[i].Tag == "border" && objects[j].Tag != "border" {
			return true
		}
		if objects[i].Tag != "border" && objects[j].Tag == "border" {
			return false
		}
		return objects[i].ZIndex < objects[j].ZIndex
	})

	// Draw all objects
	for _, obj := range objects {
		obj.Draw(screen)
	}
}

// Utility methods
func (w *World) Center(obj *Shape, resetVelocity bool) {
	obj.SetX(float64(w.Width)/2 - obj.Width/2)
	obj.SetY(float64(w.Height)/2 - obj.Height/2)

	if resetVelocity {
		obj.SetVelocity(0, 0)
	}
}

func (w *World) CenterX(obj *Shape, resetVelocity bool) {
	obj.SetX(float64(w.Width)/2 - obj.Width/2)
	if resetVelocity {
		obj.SetVelocity(0, 0)
	}
}

func (w *World) CenterY(obj *Shape, resetVelocity bool) {
	obj.SetY(float64(w.Height)/2 - obj.Height/2)
	if resetVelocity {
		obj.SetVelocity(0, 0)
	}
}

func (w *World) GetAngleBetween(a, b *Shape) float64 {
	return math.Atan2(b.Y-a.Y, b.X-a.X) * 180 / math.Pi
}

func (w *World) HoveredObjects() []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var hovered []*Shape
	for _, obj := range w.Objects {
		if w.Mouse.X >= obj.X && w.Mouse.X <= obj.X+obj.Width &&
			w.Mouse.Y >= obj.Y && w.Mouse.Y <= obj.Y+obj.Height {
			hovered = append(hovered, obj)
		}
	}
	return hovered
}

func (w *World) UnhoveredObjects() []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var unhovered []*Shape
	for _, obj := range w.Objects {
		if !(w.Mouse.X >= obj.X && w.Mouse.X <= obj.X+obj.Width &&
			w.Mouse.Y >= obj.Y && w.Mouse.Y <= obj.Y+obj.Height) {
			unhovered = append(unhovered, obj)
		}
	}
	return unhovered
}

// Object queries
func (w *World) GetAllElements() []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	result := make([]*Shape, len(w.Objects))
	copy(result, w.Objects)
	return result
}

func (w *World) GetElementsByTagName(tag string) []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var result []*Shape
	for _, obj := range w.Objects {
		if obj.Tag == tag {
			result = append(result, obj)
		}
	}
	return result
}

func (w *World) GetElementByName(name string) *Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	for _, obj := range w.Objects {
		if obj.Name == name {
			return obj
		}
	}
	return nil
}

func (w *World) GetElementsByName(name string) []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var result []*Shape
	for _, obj := range w.Objects {
		if obj.Name == name {
			result = append(result, obj)
		}
	}
	return result
}

func (w *World) GetElementsByType(shapeType ShapeType) []*Shape {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	var result []*Shape
	for _, obj := range w.Objects {
		if obj.Type == shapeType {
			result = append(result, obj)
		}
	}
	return result
}

// State control
func (w *World) Pause() {
	w.Paused = true
}

func (w *World) Resume() {
	w.Paused = false
}

func (w *World) GetCursorPosition() Vector2 {
	return Vector2{X: w.Mouse.X, Y: w.Mouse.Y}
}

// Input utilities - Thread-safe key checking
func (w *World) IsKeyPressed(key ebiten.Key) bool {
	w.keysMutex.RLock()
	defer w.keysMutex.RUnlock()
	return w.Keys[key]
}

func (w *World) OncePressed(key ebiten.Key, callback func()) {
	// This would need to be implemented with a key state tracker
	// For now, it's a placeholder
}
