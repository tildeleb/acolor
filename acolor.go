// Copyright © 2014 Lawrence E. Bakst. All rights reserved.

// NB: For the most part the code here was transliterated from Bruce Lindbloom's javascript code located at
// http://www.brucelindbloom.com
// That javascript code "colorconv.js" code has no copyright notice and no apparant license.
//
// Go's visibilty rules conlict with the names of color componets which are case sensitive.
// All color spaces have a prefix of "C" to make them public.
// All color componets are converted to upper case, which is painful, but necessary to make them public.
// e.g. Lab has becomes "CLab" and the members are "L", "A", and "B"
// for color spaces that use the same upper AND lower case letter we preceed the lower case letter with the same upper case letter
// e.g. xyY becomes "x", "Y", "Yy"
package acolor

import "math"
//import "fmt"

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
    Yy  float64
}

type ColorSpace struct {
    PRed    CxyY
    PGreen  CxyY
    PBlue   CxyY
    WP      CxyY
}

type mat3 [3][3]float64
/*
            xr = 0.64;
            yr = 0.33;
            xg = 0.30;
            yg = 0.60;
            xb = 0.15;
            yb = 0.06;
            
            RefWhiteRGB.X = 0.95047;
            RefWhiteRGB.Z = 1.08883;
            
            GammaRGB = -2.2;
            GammaRGBIndex = 3;
*/

// whitepoints
var D50Whitepoint = CXYZ{X: 0.96422054826086956, Y: 1.0, Z: 0.825208953327554} // derived
var D55Whitepoint = CXYZ{X: 0.95682, Y: 1.0, Z: 0.92149}
var D65Whitepoint = CXYZ{X: 0.950470558654283, Y: 1.0, Z: 1.08882873639588} // calculated from tables
var D75Whitepoint = CXYZ{X: 0.94972, Y: 1.0, Z: 1.22638}
var LabDefaultWhitepoint = D50Whitepoint

func compand(c, gamma float64) (ret float64) {
    var ternary = func(c bool, a, b float64) float64 {
        if c {
            return a
        } else {
            return b
        }
    }
    switch {
    case gamma > 0:
        ret = ternary(c >= 0.0, math.Pow(c, 1.0 / gamma), -math.Pow(-c, 1.0 / gamma))
    case gamma < 0:
        /* sRGB */
        sign := 1.0
        if c < 0.0 {
            sign = -1.0
            c = -c;
        }
        ret = ternary(c <= 0.0031308, (c * 12.92), (1.055 * math.Pow(c, 1.0 / 2.4) - 0.055)) * sign
    default:
        /* L* */
        sign := 1.0
        if c < 0.0 {
            sign = -1.0
            c = -c;
        }
        ret = ternary(c <= (216.0 / 24389.0), c * (24389.0 / 2700.0), (1.16 * math.Pow(c, 1.0 / 3.0) - 0.16)) * sign
    }
    return
}  

func inverseCompand(c float64) float64 {
      if c <= 0.04045 {
        return c / 12.92
      }
      return math.Pow((c + 0.055) / 1.055, 2.4)
}   

var E float64 = 216.0 / 24389.0 // 0.008856
var K float64 = 24389.0 / 27.0  // 903.296296
var KE float64 = 8.0          // 8.0

// Convert XYZ to Lab using a given whitepoint
// http://www.brucelindbloom.com/Eqn_XYZ_to_Lab.html
func (c CXYZ) ToLabwithWP(wp CXYZ) (r CLab) {
    var f = func(x float64) float64 {
        if x > E {
            return math.Pow(x, 1/3)
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

func (c CXYZ) ToLab() CLab {
    return c.ToLabwithWP(D65Whitepoint)
}

// Converts Lab to XYZ using a given whitepoint
// http://www.brucelindbloom.com/Eqn_Lab_to_XYZ.html
func (c CLab) ToXYZwithWP(wp CXYZ) (r CXYZ) {
    var ternary = func(c bool, a, b float64) float64 {
        if c {
            return a
        } else {
            return b
        }
    }
    //fmt.Printf("ToXYZwithWP: Lab=%v\n", c)
    fy := (c.L + 16) / 116
    fx := c.A / 500 + fy
    fz := fy - c.B / 200

    //fmt.Printf("fx=%f, fy=%f, fz=%f\n", fx, fy, fz)

    fx3 := math.Pow(fx, 3)
    fz3 := math.Pow(fz, 3)
    //fmt.Printf("fx3=%f, E=%f, fz3=%f, E=%f\n", fx3, E, fz3, E)

    x := ternary(fx3 > E, fx3, (116 * fx - 16) / K)
    z := ternary(fz3 > E, fz3, (116 * fz - 16) / K)
    y := ternary(c.L > KE, math.Pow((c.L + 16) / 116, 3), c.L / K)
    //fmt.Printf("x=%f, y=%f, z=%f\n", x, y, z)

    r.X = x * wp.X
    r.Y = y * wp.Y
    r.Z = z * wp.Z
    return
}

func (c CLab) ToXYZ() CXYZ {
    return c.ToXYZwithWP(D65Whitepoint)
}

// dot an XYZ color with the inverse
func (c CXYZ) dot(m mat3) (ret CRGB) {
    if true {
        ret.R = m[0][0]*c.X + m[0][1]*c.Y + m[0][2]*c.Z
        ret.G = m[1][0]*c.X + m[1][1]*c.Y + m[1][2]*c.Z
        ret.B = m[2][0]*c.X + m[2][1]*c.Y + m[2][2]*c.Z
    } else {
        ret.R = m[0][0]*c.X + m[1][0]*c.Y + m[2][0]*c.Z
        ret.G = m[0][1]*c.X + m[1][1]*c.Y + m[2][1]*c.Z
        ret.B = m[0][2]*c.X + m[1][2]*c.Y + m[2][2]*c.Z       
    }
    return
}


func (c CXYZ) ToRGB(minv [3][3]float64, clip bool) CRGB {
    var clipf = func(c float64) float64 {
        return math.Max(0, math.Min(1, c))
    }

    if c.X == 0 && c.Y == 0 && c.Z == 0 {
        return CRGB{0.0, 0.0, 0.0}
    }

    trgb := c.dot(minv)
    rgb := CRGB{compand(trgb.R, -1), compand(trgb.G, -1), compand(trgb.B, -1)}
    if clip {
        rgb.R, rgb.G, rgb.B = clipf(rgb.R), clipf(rgb.G), clipf(rgb.B)
    }
    return rgb
}

// Convert XYZ to sRGB
// http://www.brucelindbloom.com/Eqn_XYZ_to_RGB.html
func (c CXYZ) TosRGB(clip bool) CRGB {
    // sRGB inverse matrix
    minv := mat3{
        {3.2404542, -1.5371385, -0.4985314},
        {-0.9692660, 1.8760108, 0.0415560},
        {0.0556434, -0.2040259, 1.0572252},
    }
    return c.ToRGB(minv, clip)
}
