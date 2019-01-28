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

func min(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func signedDistance(p r3.Vector) float64 {
    //displacement := math.Sin(16.*p.X)*math.Sin(16.*p.Y)*math.Sin(16.*p.Z)*noiseAmplitude
    displacement := -fractalBrownianMotion(r3.Vector.Mul(p, 3.4)) * noiseAmplitude
    return r3.Vector.Norm(p) - (sphereRadius+displacement)
}

func sphereTrace(orig, dir r3.Vector, pos *r3.Vector) bool {
    if r3.Vector.Dot(orig, orig) - math.Pow(r3.Vector.Dot(orig, dir), 2) > math.Pow(sphereRadius, 2) {
        return false
    }

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
const noiseAmplitude = 1.

func hash(n float64) float64 {
    x := math.Sin(n)* 43758.5453;
    return x - math.Floor(x)
}

func lerp (v0, v1, t float64) float64 {
    return v0 + (v1-v0) * max(0., min(1., t))
}


func lerpColor (v0, v1 color.RGBA, t float64) color.RGBA {
    return color.RGBA{ v0.R + (v1.R-v0.R) * uint8(max(0., min(1.,t))),
                       v0.G + (v1.G-v0.G) * uint8(max(0., min(1.,t))),
                       v0.B + (v1.B-v0.B) * uint8(max(0., min(1.,t))),
                       255 }
}


func noise(x r3.Vector) float64 {
    p := r3.Vector{math.Floor(x.X), math.Floor(x.Y), math.Floor(x.Z)}
    f := r3.Vector{x.X-p.X, x.Y-p.Y, x.Z-p.Z}

    threes := r3.Vector{3., 3., 3.}
    f1 := r3.Vector.Mul(f, 2.)
    f2 := r3.Vector.Sub(threes, f1)
    f3 := r3.Vector.Dot(f2, f)
    f4 := r3.Vector.Mul(f, f3)

    n := r3.Vector.Dot(p, r3.Vector{1., 57., 113.})
    f = f4

    return lerp(lerp(
                    lerp(hash(n+0.), hash(n+1.), f.X),
                    lerp(hash(n+57.), hash(n+58.), f.X), f.Y),
                lerp(
                    lerp(hash(n+113.), hash(n+114.), f.X),
                    lerp(hash(n+170.), hash(n+171.), f.X), f.Y), f.Z)
}

func rotate(v r3.Vector) r3.Vector {
    p1 := r3.Vector.Dot(r3.Vector{0.00, 0.80, 0.60}, v)
    p2 := r3.Vector.Dot(r3.Vector{-0.80, 0.36, -0.48}, v)
    p3 := r3.Vector.Dot(r3.Vector{-0.60, -0.48, 0.64}, v)
    return r3.Vector{p1, p2, p3}
}

func fractalBrownianMotion(x r3.Vector) float64 {
    p := rotate(x)
    f := 0.
    f += 0.5000*noise(p);
    p = r3.Vector.Mul(p, 2.32);

    f += 0.2500*noise(p);
    p = r3.Vector.Mul(p, 3.03);

    f += 0.1250*noise(p);
    p = r3.Vector.Mul(p, 2.61);
    f += 0.0625*noise(p);
    return f/0.9375;
}

func paletteFire(d float64) color.RGBA {

    yellow := color.RGBA{uint8(255), uint8(250), uint8(0), uint8(255)} // "note that the color is "hot", i.e. has components >255"
    orange := color.RGBA{uint8(255), uint8(150), uint8(0), uint8(255)}
    red := color.RGBA{uint8(255), uint8(0), uint8(0), uint8(255)}
    darkGray := color.RGBA{uint8(50), uint8(50), uint8(50), uint8(255)}
    gray := color.RGBA{uint8(100), uint8(100), uint8(100), uint8(255)}

    x := max(0., min(1., d))
    if x<0.25 {
        return lerpColor(gray, darkGray, x*4.0)
    } else if x<0.5 {
        return lerpColor(darkGray, red, x*4.0 - 1.0)
    } else if x<0.75 {
        return lerpColor(red, orange, x*4.0 - 2.0)
    }
    return lerpColor(orange, yellow, x*4.0 - 3.0)
}

func main() {
    const width = 640.
    const height = 480.
    const fov = math.Pi/3.

    backgroundColor := color.RGBA{51, 178, 204, 255}
    whiteColor := color.RGBA{255, 255,255, 255}
    _ = whiteColor

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
            if sphereTrace(r3.Vector{0, 0, 3}, r3.Vector.Normalize(r3.Vector{dirX, dirY, dirZ}), &hit) {    // Camera is placed on {0, 0, 3}
                noiseLevel := (sphereRadius - r3.Vector.Norm(hit))/noiseAmplitude
                lightSrc := r3.Vector{10, 10, 10}
                lightDir := r3.Vector.Normalize(r3.Vector.Sub(lightSrc, hit))
                lightIntensity := max(0.4, r3.Vector.Dot(lightDir,distanceFieldNormal(hit)))
                //img.Set(int(i), int(j), multiplyColorIntensity(whiteColor, lightIntensity))
                currColor := paletteFire((-0.2 + noiseLevel)*2)
                img.Set(int(i), int(j), multiplyColorIntensity(currColor, lightIntensity))
            } else {
                img.Set(int(i), int(j), backgroundColor)
            }

        }
    }

	png.Encode(f, img)

}
