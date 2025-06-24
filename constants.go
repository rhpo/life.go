package life

import (
	"math"
	"math/rand"
	"time"
)

// Constants and enums
const (
	Deg = math.Pi / 180
)

// Shape types
type ShapeType string

const (
	ShapeCircle    ShapeType = "circle"
	ShapeSquare    ShapeType = "square"
	ShapeRectangle ShapeType = "rectangle"
	ShapeRect      ShapeType = "rectangle"
	ShapeLine      ShapeType = "line"
	ShapeDot       ShapeType = "dot"
)

// Pattern types
type PatternType string

const (
	PatternImage      PatternType = "image"
	PatternColor      PatternType = "color"
	PatternSolidColor PatternType = "color"
	PatternGradient   PatternType = "gradient"
)

// Event types
type EventType string

const (
	EventMouseDown  EventType = "mousedown"
	EventMouseUp    EventType = "mouseup"
	EventMouseMove  EventType = "mousemove"
	EventMouseEnter EventType = "mouseenter"
	EventMouseLeave EventType = "mouseleave"
	EventHover      EventType = "hover"
	EventUnHover    EventType = "unhover"
	EventClick      EventType = "click"
	EventCollision  EventType = "collision"
)

// Cursor types
type CursorType string

const (
	CursorDefault   CursorType = "default"
	CursorPointer   CursorType = "pointer"
	CursorCrosshair CursorType = "crosshair"
	CursorMove      CursorType = "move"
	CursorText      CursorType = "text"
)

// Utility functions
func ID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 7)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func RandName() string {
	names := []string{"James", "Robert", "John", "Michael", "David", "William", "Richard", "Thomas", "Charles", "Islam", "Mohammed", "Ramy"}
	return names[rand.Intn(len(names))]
}

func Defined(v interface{}) bool {
	return v != nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
