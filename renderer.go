package main

import (
	"fmt"
	"image"
	img_color "image/color"
	"math"
	"sync"
	"time"
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

type RayMap struct { // for use to send ray and matching XY coordinates through channels in one
	X   float64
	Y   float64
	ray Coords
}

type ColorMap struct { // for use to send colro and matching XY coordinates through channels in one
	X     float64
	Y     float64
	color Color
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

type Camera struct {
	position        Coords
	rotation_matrix [3][3]float64 // fancy matrix multiplication stuff that lets you rotate the view
}

const (
	// Canvas Settings
	C_Width  float64 = 1080
	C_Height float64 = 1080

	// Viewport Settings
	V_Height float64 = 1
	V_Width  float64 = 1
	V_Dist   float64 = 1

	Recursion_Depth int = 1 // determines how many iterations of recursion are used for reflection

	Intensity_Threshold float64 = 1.2 // threshold used to discretize colors for cel shading
)

var inf = math.Inf(1)

func ReflectRay(ray, normal Coords) Coords {
	return Subtract(Multiply(normal, Dot(normal, ray)*2), ray)
}

func DrawPixel(canvas *image.RGBA, C_x float64, C_y float64, color Color) {
	// Scene Coordinates
	S_x := (C_Width / 2) + C_x
	S_y := (C_Height / 2) - C_y

	normal_color := NormalizeColor(color)

	canvas.Set(int(S_x), int(S_y), img_color.RGBA{R: uint8(normal_color.red), G: uint8(normal_color.green), B: uint8(normal_color.blue), A: 255})

	// canvas.SetRGB(NormalizeColor(color))
	// canvas.DrawPoint(S_x, S_y, 1)
	// canvas.Fill()
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
		return inf, inf
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
				t_max = inf
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

	// weight intensity against thresholds to quantize colors; creating "cell" shading
	if intensity > Intensity_Threshold*0.95 {
		intensity = Intensity_Threshold * 0.95
	} else if intensity > Intensity_Threshold*0.50 {
		intensity = Intensity_Threshold * 0.50
	} else if intensity > Intensity_Threshold*0.30 {
		intensity = Intensity_Threshold * 0.30
	} else {
		intensity = Intensity_Threshold * 0.05
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
	t_closest := inf
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
	reflected_color := TraceRay(bg_color, scene, point, ray, 0.001, inf, recursion_depth-1) // recursive case

	return AddColor(IntensifyColor(local_color, 1-reflective), IntensifyColor(reflected_color, reflective))
}

func main() {
	start := time.Now()

	// canvas := gg.NewContext(int(C_Width), int(C_Height))

	canvas := image.NewRGBA(image.Rect(0, 0, int(C_Width), int(C_Height)))

	// Set the background color
	// canvas.SetRGB(0, 0, 0) // black
	// canvas.Clear()

	camera := Camera{
		rotation_matrix: [3][3]float64{
			{1, 0, 0},
			{0, 1, 0},
			{0, 0, 1},
		},
		position: Coords{0, 0, 0},
	}
	bg_color := Color{255, 255, 255}

	scene := Scene{
		viewport_size:      [2]int{1, 1},
		projection_plane_d: 100,
		spheres: []Sphere{
			{center: Coords{0, -1, 3}, radius: 1, color: Color{255, 0, 0}, specular: 250, reflective: 0},
			{center: Coords{2, 0, 4}, radius: 1, color: Color{0, 0, 255}, specular: 250, reflective: 0.25},
			{center: Coords{-2, 0, 4}, radius: 1, color: Color{0, 255, 0}, specular: 5, reflective: 0},
			{center: Coords{0, -5001, 0}, radius: 5000, color: Color{255, 255, 0}, specular: 500, reflective: 0},
		},
		lights: []Light{
			AmbientLight{intensity: 0.2},
			PointLight{intensity: 0.6, position: Coords{2, 1, 0}},
			DirectionalLight{intensity: 0.2, direction: Coords{1, 4, 4}},
		},
	}

	var wg sync.WaitGroup
	wg.Add(3)

	ray_chan := make(chan RayMap, int(C_Width*C_Height))
	color_chan := make(chan ColorMap, int(C_Width*C_Height))

	go func(ray_chan chan RayMap) {
		defer wg.Done()

		for x := -(C_Width / 2); x < (C_Width / 2); x++ {
			for y := -(C_Height / 2); y < (C_Height / 2); y++ {
				ray := MultiplyMatrixVector(camera.rotation_matrix, CanvasToViewPort(x, y))

				ray_chan <- RayMap{
					X:   x,
					Y:   y,
					ray: ray,
				}

				// fmt.Printf("\n\nSetting ray of (%f, %f) to hit (%f, %f, %f)", x, y, ray.x, ray.y, ray.z)
			}
		}

		close(ray_chan)
	}(ray_chan)

	go func(ray_chan chan RayMap, color_chan chan ColorMap) {
		defer wg.Done()

		for ray := range ray_chan {
			color := TraceRay(bg_color, scene, camera.position, ray.ray, 1, inf, Recursion_Depth)

			color_chan <- ColorMap{
				X:     ray.X,
				Y:     ray.Y,
				color: color,
			}

			// fmt.Printf("\n\nSetting color at (%f, %f) as (%f, %f, %f)", ray.X, ray.Y, color.red, color.green, color.blue)
		}

		close(color_chan)
	}(ray_chan, color_chan)

	go func(color_chan chan ColorMap) {
		defer wg.Done()

		for color := range color_chan {
			// fmt.Printf("\n\nDrawing pixel at (%f, %f) as (%f, %f, %f)", color.X, color.Y, color.color.red, color.color.green, color.color.blue)

			DrawPixel(canvas, color.X, color.Y, color.color)
		}
	}(color_chan)

	wg.Wait()

	EdgeDetection(canvas) // post processing to add image edges for cel shading

	SavePNG("img.png", canvas)

	elapsed := time.Since(start)
	fmt.Printf("Image size: %.0fx%.0f, Execution time: %s", C_Width, C_Height, elapsed)
}
