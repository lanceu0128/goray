package main

import (
	"fmt"
	"math"
	"time"

	"github.com/fogleman/gg"
)

type Color struct {
	red   float64
	green float64
	blue  float64
}

type Coords struct {
	x float64
	y float64
	z float64
}

type Sphere struct {
	center     Coords
	radius     float64
	color      Color
	specular   float64 // fancy word for shininess
	reflective float64
}

type Light interface {
	GetIntensity() float64
}

type AmbientLight struct {
	intensity float64
}

func (light AmbientLight) GetIntensity() float64 {
	return light.intensity
}

type PointLight struct {
	intensity float64
	position  Coords
}

func (light PointLight) GetIntensity() float64 {
	return light.intensity
}

type DirectionalLight struct {
	intensity float64
	direction Coords
}

func (light DirectionalLight) GetIntensity() float64 {
	return light.intensity
}

type Scene struct {
	viewport_size      [2]int
	projection_plane_d int
	spheres            []Sphere
	lights             []Light
}

const (
	// Canvas Settings
	C_Width  float64 = 500
	C_Height float64 = 500

	// Viewport Settings
	V_Height float64 = 1
	V_Width  float64 = 1
	V_Dist   float64 = 1

	Recursion_Depth int = 3 // determines how many iterations of recursion are used for reflection
)

/*
	COORDINATE / VECTOR FUNCTIONS
*/

func Add(coords1, coords2 Coords) Coords {
	coords := Coords{
		x: coords1.x + coords2.x,
		y: coords1.y + coords2.y,
		z: coords1.z + coords2.z,
	}
	return coords
}

func Subtract(coords1, coords2 Coords) Coords {
	coords := Coords{
		x: coords1.x - coords2.x,
		y: coords1.y - coords2.y,
		z: coords1.z - coords2.z,
	}
	return coords
}

func Multiply(coords Coords, scalar float64) Coords {
	result := Coords{
		x: coords.x * scalar,
		y: coords.y * scalar,
		z: coords.z * scalar,
	}
	return result
}

func Divide(coords Coords, scalar float64) Coords {
	result := Coords{
		x: coords.x / scalar,
		y: coords.y / scalar,
		z: coords.z / scalar,
	}
	return result
}

func Dot(coords1, coords2 Coords) float64 {
	dotProduct := coords1.x*coords2.x + coords1.y*coords2.y + coords1.z*coords2.z
	return dotProduct
}

func Length(vec Coords) float64 {
	magnitude := math.Sqrt((vec.x * vec.x) + (vec.y * vec.y) + (vec.z * vec.z))
	return magnitude
}

func IntensifyColor(color Color, intensity float64) Color {
	result := Color{
		red:   color.red * intensity,
		green: color.green * intensity,
		blue:  color.blue * intensity,
	}
	return result
}

func AddColor(color1, color2 Color) Color {
	result := Color{
		red:   color1.red + color2.red,
		green: color1.green + color2.green,
		blue:  color1.blue + color2.blue,
	}
	return result
}

func NormalizeColor(color Color) (float64, float64, float64) {
	// all colors "intensified" after 255 need to be flattened to 255 for proper rendering
	if color.red > 255 {
		color.red = 255
	}
	if color.green > 255 {
		color.green = 255
	}
	if color.blue > 255 {
		color.blue = 255
	}

	// graphics library requires color channel values to be from 0-1
	return color.red / 255, color.green / 255, color.blue / 255
}

func ReflectRay(ray, normal Coords) Coords {
	return Subtract(Multiply(normal, Dot(normal, ray)*2), ray)
}

/*
	CANVAS / RAY TRACING OPERATIONS
*/

func DrawPixel(canvas *gg.Context, C_x float64, C_y float64, color Color) {
	// Scene Coordinates
	S_x := (C_Width / 2) + C_x
	S_y := (C_Height / 2) - C_y

	canvas.SetRGB(NormalizeColor(color))
	canvas.DrawPoint(S_x, S_y, 1)
	canvas.Fill()
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

	discriminant := (b * b) - (4 * a * c)

	if discriminant < 0 {
		return math.Inf(1), math.Inf(1)
	}

	t1 := (-b + math.Sqrt(discriminant)) / (2 * a)
	t2 := (-b - math.Sqrt(discriminant)) / (2 * a)

	return t1, t2
}

/*
Computes intensity at a certain point's color
- scene = 3D scene
- point = coordinates in 3D scene where light hits
- normal = unit vector perpendicular to the surface of point
*/
func ComputeLighting(scene Scene, point Coords, normal Coords, view Coords, specularity float64) float64 {
	// intensity = multiplier given to color in rendering to simulate light
	intensity := 0.00
	var t_max float64

	for _, light := range scene.lights {
		if ambient_light, ok := light.(AmbientLight); ok {
			intensity += ambient_light.intensity
		} else {
			// light = directional vector of the incoming light ray
			var light_vec Coords

			if point_light, ok := light.(PointLight); ok {
				light_vec = Subtract(point_light.position, point)
				t_max = 1
			} else if directional_light, ok := light.(DirectionalLight); ok {
				light_vec = directional_light.direction
				t_max = math.Inf(1)
			}

			// SHADOW CHECK (skip reflections if shadow is being cased on point)
			sphere_shadow, _ := ClosestIntersection(scene, point, light_vec, 0.001, t_max)
			if sphere_shadow != (Sphere{}) {
				continue
			}

			// DIFFUSE REFLECTION (reflection for matte objects) CALCULATION
			n_dot_l := Dot(normal, light_vec)
			// don't add negative values to intensity
			if n_dot_l > 0 {
				intensity += light.GetIntensity() * (n_dot_l / (Length(normal) * Length(light_vec)))
			}

			// SPECULAR REFLECTION (reflection for shiny objects) CALCULATION
			if specularity != -1 {
				reflection := ReflectRay(light_vec, normal)
				r_dot_v := Dot(reflection, view)
				if r_dot_v > 0 {
					intensity += light.GetIntensity() * math.Pow(r_dot_v/(Length(reflection)*Length(view)), specularity)
				}
			}
		}
	}
	return intensity
}

/*
Find the intersections between the camera and objects in a certain direction
- scene = 3D scene containing objects
- camera = coordinates of the camera
- dir = coordinates of the direction being drawn based on X and Y coordinates
- t_min = minimum magnitude of vector extending from viewport
- t_max = maximum magnitude of vector extending from viewport
*/
func ClosestIntersection(scene Scene, camera Coords, dir Coords, t_min float64, t_max float64) (Sphere, float64) {
	t_closest := math.Inf(1)
	var sphere_closest Sphere

	// finds closest sphere by intersecting rays and spheres and finding minimum within bounds t_min and t_max
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

	return sphere_closest, t_closest
}

/*
Performs calculations to return color based on a given direction
- scene = 3D scene containing objects
- camera = coordinates of the camera
- dir = coordinates of the direction being drawn based on X and Y coordinates
- t_min = minimum magnitude of vector extending from viewport
- t_max = maximum magnitude of vector extending from viewport
*/
func TraceRay(bg_color Color, scene Scene, camera Coords, dir Coords, t_min float64, t_max float64, recursion_depth int) Color {

	sphere_closest, t_closest := ClosestIntersection(scene, camera, dir, t_min, t_max)

	// checks for empty Sphere (null but Go doesn't have null)
	if sphere_closest == (Sphere{}) {
		return bg_color
	}

	// INTENSITY CALCULATIONS:

	// point = coordinates in 3D scene where light hits (OOO unclear...)
	point := Add(camera, Multiply(dir, t_closest))
	// normal = unit vector perpendicular to the surface of point
	normal := Subtract(point, sphere_closest.center)
	normal = Divide(normal, Length(normal))
	intensity := ComputeLighting(scene, point, normal, Multiply(dir, -1), sphere_closest.specular)
	local_color := IntensifyColor(sphere_closest.color, intensity)

	// REFLECTION CALCULATIONS:

	reflective := sphere_closest.reflective
	if recursion_depth <= 0 || reflective <= 0 { // base case
		return local_color
	}

	ray := ReflectRay(Multiply(dir, -1), normal)
	reflected_color := TraceRay(bg_color, scene, point, ray, 0.001, math.Inf(1), recursion_depth-1) // recursive case

	return AddColor(IntensifyColor(local_color, 1-reflective), IntensifyColor(reflected_color, reflective))
}

func main() {
	start := time.Now()

	canvas := gg.NewContext(int(C_Width), int(C_Height))

	// Set the background color
	canvas.SetRGB(0, 0, 0) // black
	canvas.Clear()

	camera := Coords{0, 0, 0}
	bg_color := Color{0, 0, 0}

	scene := Scene{
		viewport_size:      [2]int{1, 1},
		projection_plane_d: 100,
		spheres: []Sphere{
			{center: Coords{0, -1, 3}, radius: 1, color: Color{255, 0, 0}, specular: 500, reflective: 0.2},
			{center: Coords{2, 0, 4}, radius: 1, color: Color{0, 0, 255}, specular: 500, reflective: 0.3},
			{center: Coords{-2, 0, 4}, radius: 1, color: Color{0, 255, 0}, specular: 10, reflective: 0.4},
			{center: Coords{0, -5001, 0}, radius: 5000, color: Color{255, 255, 0}, specular: 1000, reflective: 0.5},
		},
		lights: []Light{
			AmbientLight{intensity: 0.2},
			PointLight{intensity: 0.6, position: Coords{2, 1, 0}},
			DirectionalLight{intensity: 0.2, direction: Coords{1, 4, 4}},
		},
	}

	for x := -(C_Width / 2); x < (C_Width / 2); x++ {
		for y := -(C_Height / 2); y < (C_Height / 2); y++ {
			dir := CanvasToViewPort(x, y)
			color := TraceRay(bg_color, scene, camera, dir, 1, math.Inf(1), Recursion_Depth)
			DrawPixel(canvas, x, y, color)
		}
	}

	save_location := "img.png"
	canvas.SavePNG(save_location)

	elapsed := time.Since(start)
	fmt.Printf("Image size: (%.0f, %.0f), Execution time: %s, Saved to: %s", C_Width, C_Height, elapsed, save_location)
}
