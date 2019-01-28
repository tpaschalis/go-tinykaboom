package main

import "github.com/golang/geo/r3"
import "math"
import "os"

import "image"
import "image/color"
import "image/png"

func max(a, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func signedDistance(p r3.Vector) float64 {
    displacement := math.Sin(16.*p.X)*math.Sin(16.*p.Y)*math.Sin(16.*p.Z)*noiseAmplitude
    return r3.Vector.Norm(p) - (sphereRadius+displacement)
}

func sphereTrace(orig, dir r3.Vector, pos *r3.Vector) bool {
    *pos = orig
    for i:=0; i<128; i++ {
        d := signedDistance(*pos)
        if d < 0 {
            return true
        }
        *pos = r3.Vector.Add(*pos,r3.Vector.Mul(dir, max(d*0.1, 0.1)))
    }
    return false
}

func distanceFieldNormal(pos r3.Vector) r3.Vector {
    const eps = 0.1
    d := signedDistance(pos)
    nx := signedDistance(r3.Vector.Add(pos, r3.Vector{eps, 0, 0})) - d
    ny := signedDistance(r3.Vector.Add(pos, r3.Vector{0, eps, 0})) - d
    nz := signedDistance(r3.Vector.Add(pos, r3.Vector{0, 0, eps})) - d
    return r3.Vector.Normalize(r3.Vector{nx, ny, nz})
}

func multiplyColorIntensity(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * f),
		uint8(float64(c.G) * f),
		uint8(float64(c.B) * f),
		255}
}

const sphereRadius = 1.5
const noiseAmplitude = 0.2

func main() {
    const width = 640.
    const height = 480.
    const fov = math.Pi/3.

    backgroundColor := color.RGBA{51, 178, 204, 255}
    whiteColor := color.RGBA{255, 255,255, 255}

	f, _ := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

    var dirX, dirY, dirZ float64
    for j:=0.; j<height; j++ {
        for i:=0.; i<width; i++ {
            dirX = (i+0.5) - width/2.
            dirY = -(j+0.5) + height/2.     // "this flips the image at the same time"
            dirZ = -height/(2.*math.Tan(fov/2.))

            var hit r3.Vector
            if sphereTrace(r3.Vector{0, 0, 3}, r3.Vector.Normalize(r3.Vector{dirX, dirY, dirZ}), &hit) {
                //img.Set(int(i), int(j), color.RGBA{255, 255, 255, 255})
                lightSrc := r3.Vector{10, 10, 10}
                lightDir := r3.Vector.Normalize(r3.Vector.Sub(lightSrc, hit))
                lightIntensity := max(0.4, r3.Vector.Dot(lightDir,distanceFieldNormal(hit)))
                img.Set(int(i), int(j), multiplyColorIntensity(whiteColor, lightIntensity))
            } else {
                img.Set(int(i), int(j), backgroundColor)
            }

        }
    }

	png.Encode(f, img)

}
