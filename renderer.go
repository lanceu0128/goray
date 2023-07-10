package main

import (
	"math"
	"github.com/fogleman/gg"
)

type Color struct {
	red float64
	green float64
	blue float64
}

type Coords struct {
	x float64
	y float64
	z float64
}

type Sphere struct {
	center Coords
	radius float64
	color Color
}

type Scene struct {
	viewport_size [2]int
	projection_plane_d int
	spheres []Sphere 
}

const (
	// Canvas Settings
	C_Width float64 = 250
	C_Height float64 = 250

	// Viewport Settings
	V_Height float64 = 1
	V_Width float64 = 1
	V_Dist float64 = 1
)

func DrawPixel(canvas *gg.Context, C_x float64, C_y float64, color Color) {
	// Scene Coordinates
	S_x := (C_Width / 2) + C_x 
	S_y := (C_Height / 2) - C_y 

	canvas.SetRGB(color.red, color.green, color.blue)
	canvas.DrawPoint(S_x, S_y, 1) 
	canvas.Fill()
}

func Subtract(coords1, coords2 Coords) Coords {
	result := Coords{
		x: coords1.x - coords2.x,
		y: coords1.y - coords2.y,
		z: coords1.z - coords2.z,
	}

	return result
}

func Dot(coords1, coords2 Coords) float64 {
	dotProduct := coords1.x*coords2.x + coords1.y*coords2.y + coords1.z*coords2.z
	return dotProduct
}

func CanvasToViewPort(C_x, C_y float64) Coords {
	V_x := C_x * V_Width / C_Width
	V_y := C_y * V_Height / C_Height
	V_z := V_Dist

	return Coords{V_x, V_y, V_z}
}

func IntersectRaySphere(camera Coords, dir Coords, sphere Sphere) (float64, float64) {
	rad := sphere.radius
	dist := Subtract(camera, sphere.center)

	a := Dot(dir, dir)	
	b := 2 * Dot(dist, dir)
	c := Dot(dist, dist) - (rad * rad)

	discriminant := (b*b) - (4*a*c)

	if discriminant < 0 {
		return math.Inf(1), math.Inf(1)
	}

	t1 := (-b + math.Sqrt(discriminant)) / (2*a)
	t2 := (-b - math.Sqrt(discriminant)) / (2*a)

	return t1, t2
}

func TraceRay(bg_color Color, scene Scene, camera Coords, dir Coords, t_min float64, t_max float64) Color {
	t_closest := math.Inf(1)
	var sphere_closest Sphere

	for _, sphere := range scene.spheres {
		t1, t2 := IntersectRaySphere(camera, dir, sphere)

		if t1 >= t_min && t1 <= t_max && t1 < t_closest {
			t_closest = t1
			sphere_closest = sphere
		}
		if t2 >= t_min && t2 <= t_max && t2 < t_closest {
			t_closest = t2
			sphere_closest = sphere
		}
	}

	if sphere_closest == (Sphere{}) {
		return bg_color
	}

	return sphere_closest.color
}

func main() {
    canvas := gg.NewContext(int(C_Width), int(C_Height))

    // Set the background color
	canvas.SetRGB(0, 0, 0) // black
	canvas.Clear()

	camera := Coords{0, 0, 0}

	bg_color := Color{1, 1, 1}

	scene := Scene {
		viewport_size: [2]int{1, 1},
		projection_plane_d: 100,
		spheres: []Sphere{
			Sphere{center: Coords{0, -1, 3}, radius: 1, color: Color{1, 0, 0}},
			Sphere{center: Coords{2, 0, 4}, radius: 1, color: Color{0, 0, 1}},
			Sphere{center: Coords{-2, 0, 4}, radius: 1, color: Color{0, 1, 0}},
		},
	}

	for x := - (C_Width / 2); x < (C_Width / 2); x++ {
		for y := - (C_Height / 2); y < (C_Height / 2); y++ {
			dir := CanvasToViewPort(x, y)
			color := TraceRay(bg_color, scene, camera, dir, 1, math.Inf(1))	
			DrawPixel(canvas, x, y, color)
		} 
	} 

	canvas.SavePNG("img.png")
}