package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"sync"
)

type MatrixMap struct { // used by matrix channel to link X and Y coordinates to their matrix to be processed by the sobel
	x      int
	y      int
	matrix [3][3]uint8
}

func SobelKernel(matrix [3][3]uint8) float64 {
	Gradient_X := [3][3]int{ // used to convolute image and find X edge
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	Gradient_Y := [3][3]int{ // convolute and fins Y edge
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	sobel_x := Convolution(matrix, Gradient_X)
	sobel_y := Convolution(matrix, Gradient_Y)

	return math.Sqrt(float64((sobel_x * sobel_x) + (sobel_y * sobel_y))) // squre root to get combined results of two convolutions
}

func EdgeDetection(canvas *image.RGBA) {
	matrix_chan := make(chan MatrixMap, (canvas.Bounds().Max.Y-canvas.Bounds().Min.Y)*(canvas.Bounds().Max.X-canvas.Bounds().Min.X))
	var wg sync.WaitGroup
	wg.Add(2)

	gray_canvas := image.NewGray(canvas.Bounds())
	edge_canvas := image.NewGray(canvas.Bounds())
	draw.Draw(edge_canvas, edge_canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	for y := canvas.Bounds().Min.Y; y < canvas.Bounds().Max.Y; y++ {
		for x := canvas.Bounds().Min.X; x < canvas.Bounds().Max.X; x++ {
			r, g, b, _ := canvas.At(x, y).RGBA()
			gray := uint8((r + g + b) / 3 >> 8) // Convert RGB to grayscale value
			gray_canvas.SetGray(x, y, color.Gray{Y: gray})
		}
	}

	// Loop through the grayscale image to extract 3x3 matrices and apply Sobel operator
	go func() {
		defer wg.Done()

		for y := 1; y < gray_canvas.Bounds().Max.Y-1; y++ {
			for x := 1; x < gray_canvas.Bounds().Max.X-1; x++ {
				// Extract the 3x3 matrix of grayscale values
				var matrix [3][3]uint8
				for i := -1; i <= 1; i++ {
					for j := -1; j <= 1; j++ {
						pixel := gray_canvas.GrayAt(x+j, y+i)
						matrix[i+1][j+1] = pixel.Y
					}
				}

				matrix_chan <- MatrixMap{x: x, y: y, matrix: matrix}
			}
		}

		close(matrix_chan)
	}()

	go func() {
		defer wg.Done()

		for matrix_map := range matrix_chan {
			edge := SobelKernel(matrix_map.matrix)
			edge_canvas.SetGray(matrix_map.x, matrix_map.y, color.Gray{Y: 255 - uint8(edge)})
		}
	}()

	wg.Wait()

	// Create a new RGBA image with the same dimensions as the original RGBA image
	final_canvas := image.NewRGBA(canvas.Bounds())

	// Copy the original RGBA image to the overlayed image
	draw.Draw(final_canvas, final_canvas.Bounds(), canvas, image.Point{}, draw.Src)

	bounds := final_canvas.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			alpha := edge_canvas.GrayAt(x, y).Y
			colorRGBA := final_canvas.RGBAAt(x, y)
			newColor := color.RGBA{colorRGBA.R, colorRGBA.G, colorRGBA.B, alpha}
			final_canvas.SetRGBA(x, y, newColor)
		}
	}

	SaveGrayPNG("gray.png", gray_canvas)
	SaveGrayPNG("edges.png", edge_canvas)
	SavePNG("final.png", final_canvas)

	save_location := "out_edges.png"
	file_to_save := final_canvas
	file, err := os.Create(save_location)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	png.Encode(file, file_to_save)
}
