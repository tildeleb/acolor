// Copyright © 2014 Lawrence E. Bakst. All rights reserved.
package acolor

import "math"

type CLab struct {
    L   float64
    A   float64
    B   float64
}

type CXYZ struct {
    X   float64
    Y   float64
    Z   float64
}

type CRGB struct {
    R   float64
    G   float64
    B   float64   
}

type CxyY struct {
    X   float64
    Y   float64
    YY  float64
}

type ColorSpace struct {
    PRed    CxyY
    PGreen  CxyY
    PBlue   CxyY
    WP      CxyY
}

type mat3 [3][3]float64

func compand(c float64) float64 {
    if c <= 0.0031308 {
        return c * 12.92
    }
    return 1.055 * math.Pow(c, 1 / 2.4) - 0.055
}  


func inverseCompand(c float64) float64 {
      if c <= 0.04045 {
        return c / 12.92
      }
      return math.Pow((c + 0.055) / 1.055, 2.4)
}

// Convert XYZ to Lab using a given whitepoint
// http://www.brucelindbloom.com/Eqn_XYZ_to_Lab.html
func XYZToLabwithWP(c, wp CXYZ) (r CLab) {
    var f = func(x float64) float64 {
        if x > E {
            return math.Pow(x, 1/3.0)
        } else {
            return (K*x + 16) / 116
        }
    }

    // normalize to WP
    X := c.X/wp.X
    Y := c.Y/wp.Y
    Z := c.Z/wp.Z

    fx := f(X)
    fy := f(Y)
    fz := f(Z)
    r.L = 116 * fy - 16
    r.A = 500 * (fx - fy)
    r.B = 200 * (fy - fz)
    return
}

const E = 216.0 / 24389
const K = 24389.0 / 27
const KE = 216.0 / 27

// Converts Lab to XYZ using a given whitepoint
// http://www.brucelindbloom.com/Eqn_Lab_to_XYZ.html
func LabToXYZwithWP(c CLab, wp CXYZ) (r CXYZ) {
    var ternary = func(c bool, a, b float64) float64 {
        if c {
            return a
        } else {
            return b
        }
    }
    fy := (c.L + 16) / 116
    fz := fy - c.B / 200
    fx := c.A / 500 + fy

    fx3 := math.Pow(fx, 3)
    fz3 := math.Pow(fz, 3)

    x := ternary(fx3 > E, fx3, (116 * fx - 16) / K)
    z := ternary(fz3 > E, fz3, (116 * fz - 16) / K)
    y := ternary(c.L > KE, math.Pow((c.L + 16) / 116, 3), c.L / K)

    r.X = x * wp.X
    r.Y = y * wp.Y
    r.Z = z * wp.Z
    return
}

// dot an XYZ color with the inverse
func (c CXYZ) dot(m mat3) (ret CRGB) {
    ret.R = m[0][0]*c.X + m[0][1]*c.Y + m[0][2]*c.Z
    ret.G = m[1][0]*c.X + m[1][1]*c.Y + m[1][2]*c.Z
    ret.G = m[2][0]*c.X + m[2][1]*c.Y + m[2][2]*c.Z
    return
}

// Convert XYZ to RGB
// http://www.brucelindbloom.com/Eqn_XYZ_to_RGB.html
func (c CXYZ) XYZtosRGB(clip bool) CRGB {
    var cf = func(c float64) float64 {
        return math.Max(0, math.Min(1, c))
    }

    if c.X == 0 && c.Y == 0 && c.Z == 0 {
        return CRGB{0.0, 0.0, 0.0}
    }

    m := mat3{
        {3.2404542, -1.5371385, -0.4985314},
        {-0.9692660, 1.8760108, 0.0415560},
        {0.0556434, -0.2040259, 1.0572252},
    }

    trgb := c.dot(m) 
    rgb := CRGB{compand(trgb.R), compand(trgb.G), compand(trgb.B)}
    if clip {
        rgb.R, rgb.G, rgb.B = cf(rgb.R), cf(rgb.G), cf(rgb.B)
    }
    return rgb
}