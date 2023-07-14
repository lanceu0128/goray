package main

import (
	"image"
	"image/png"
	"math"
	"os"
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
	new_coords := Coords{
		x: coords.x * scalar,
		y: coords.y * scalar,
		z: coords.z * scalar,
	}
	return new_coords
}

func Divide(coords Coords, scalar float64) Coords {
	new_coords := Coords{
		x: coords.x / scalar,
		y: coords.y / scalar,
		z: coords.z / scalar,
	}
	return new_coords
}

func Dot(coords1, coords2 Coords) float64 {
	dotProduct := coords1.x*coords2.x + coords1.y*coords2.y + coords1.z*coords2.z
	return dotProduct
}

func Length(vec Coords) float64 {
	magnitude := math.Sqrt((vec.x * vec.x) + (vec.y * vec.y) + (vec.z * vec.z))
	return magnitude
}

func MultiplyMatrixVector(matrix [3][3]float64, vec Coords) Coords {
	new_coords := Coords{
		x: (vec.x * matrix[0][0]) + (vec.y * matrix[0][1]) + (vec.z * matrix[0][2]),
		y: (vec.x * matrix[1][0]) + (vec.y * matrix[1][1]) + (vec.z * matrix[1][2]),
		z: (vec.x * matrix[2][0]) + (vec.y * matrix[2][1]) + (vec.z * matrix[2][2]),
	}
	return new_coords
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

func NormalizeColor(color Color) Color {
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
	return Color{color.red, color.green, color.blue}
}

func Convolution(matrix [3][3]uint8, gradient [3][3]int) int {
	var result int
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			result += int(matrix[i][j]) * gradient[i][j] // get sum of products of each matrix[i][j]
		}
	}

	if result < 0 {
		result = 0
	} else if result > 50 {
		result = 255
	}

	return result
}

/*
	File Operations
*/

func SavePNG(file_location string, canvas *image.RGBA) {
	file, err := os.Create(file_location)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	png.Encode(file, canvas)
}

func SaveGrayPNG(file_location string, canvas *image.Gray) {
	file, err := os.Create(file_location)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	png.Encode(file, canvas)
}
