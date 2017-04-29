# Wabisabi

Wabisabi is a renderer of buddhabrot and its family members. It shares its name with a Japanese asthethic called [Wabi-sabi](https://en.wikipedia.org/wiki/Wabi-sabi). Referencing the impossibility of creating the real buddhabrot and learning to accept the beauty in reality and its flaws. 

_The name will probably be changed to it's lovely nickname wasabi anytime soon hahaha <3_

> To the left, an original buddhabrot and to the right an anti-abrot.

<img src=https://github.com/karlek/wabisabi/blob/master/img/original.jpg?raw=true width=49.9%>
<img src=https://github.com/karlek/wabisabi/blob/master/img/anti.jpg?raw=true width=49.9%>

> An image of the calculation path rendering technique.

<img src=https://github.com/karlek/wabisabi/blob/master/img/calc.jpg?raw=true width=49.9%>

## Install

```fish
$ go get github.com/karlek/wabisabi
```

## Run

```fish
# Be sure to limit the memory usage beforehand; wabisabi is greedy little devil.
$ ulimit -Sv 4000000 # Where the number is the memory in kB.
$ wabisabi
```

## Features

* Saving and loading of histograms to re-render with different exposures.
* Calculating the original, anti- and primitive- buddhabrot.
* Exploring the different planes of Zr, Zi, Cr and Ci.
* Different histogram equalization functions (think color scaling).
* Using the color palette of an image to color the brot.
* Change the co-efficient of the complex function i.e __a__\*z\*z+__a__\*c
* Zooming.
* Multiple CPU support. 
* Hand optimized assembly(!) for generating random complex points. Thank you [7i](https://github.com/7i)!
* Plot Calculation-Paths. Credits to Raka Jovanovic and Milan Tuba (ISSN: 1109-2750).

>It should be noted that speed in random number generating algorithms competes with the necessity of having a random distribution. If you know of a way to benchmark randomness as well as speed, please create an issue!

![Benchmark](https://github.com/karlek/wabisabi/blob/master/img/benchmark.png?raw=true)

## Future features
t
* Metropolis-hastings algorithm for faster zooming.
* Orbit trapping; would be amazing!

## Random area / notes to myself

### Complex functions

Many complex functions which can be iterated create interesting orbits. 

```go
z = |z*z| + c
complex(real(c), -imag(c))
complex(-math.Abs(real(c)), imag(c))
complex(math.Abs(real(c)), imag(c))
complex(imag(c)-real(c), real(c)*imag(c))
```

### Z<sub>0</sub>

```go
z := randomPoint(random)
z := complex(math.Sin(real(c)), math.Sin(imag(c)))
```

### Future

* Only allow a certain type of orbits. 
    - How to discern between different types?
        + Constant increment on certain axis indicates spirals?
        + Convex hull to check roundness?
        + Is iteration length related to orbit types?

* Super sampling
    - Not sure how this differs from rendering a larger buddhabrot and just downsizing it?
        + Probably is just skipping the render and resizing step and calculating the values in the histograms directly.

* Since the orbits reminds me of a circle; it could be possible to unravel the circle and convert them into sine-waves to create tones :D
    - Outer convex hull to get the radius and by extension the amplitude. 

* Test slices instead of fixed size arrays for runtime allocation of iterations and width/height.

* More than 3 histograms?
    - Doesn't this only make sense with color spaces with more than 3 values such as CMYK?

### Co-efficient

The coefficient on the __real__ axis has two properties:

* Why does the coefficient seem to be capped at 1.37~? 
* When larger than _1_ it twists into something looking like a set of armor.
    - This then eventually twists into itself at around 1.37~ where it becomes only two specks of dust.
    - It twists on two points towards the center.
    - Try with values like: _1.01_.
* When smaller than _1_ it works as a zoom. 
    - On which axis? Both real and imaginary? Or only real? Not sure.  
* When smaller than _0_ (-1.1 to 0) it spirals in on itself.
    - It rotates on one point towards itself.

The coefficient on the __imaginary__ axis has two properties:

* When slightly larger than _1_ it makes the buddhabrot more ... ephemeral? Try with values like: _1.001_.
* When smaller than _1_
* When smaller than _0_ the right side of the brot becomes corrupted. Really cool!
    - Try with values like: _-0.01_ and _-0.1_.
    - With values like _-.5_ it looks like a sinking ship.

Combining _both_ coefficient:

### Problems

By allowing coefficient and exploring different planes a slow down of 30% is observed. Ugly solution is to create special function for each possibility.

### Possible bug

```fish
# width = 3000, height = 4000
go install; wabisabi -zoom 1 -seed 1 -tries 0.1
go install; wabisabi -zoom 0.5 -seed 1 -tries 0.1
```

Also have switched the real and imaginary axis. The zoom value should be on the imaginary axis not the real axis.

Problem lies with img.Set(y, x) and that the histogram has the wrong size.
With the current implementation. We can't allow aspect ratios other than 1:1.

### Fun stuff

Interesting old bug:

```go
p.X = int((zoom*float64(width)/2.8)*(r+real(offset))) + width/2
p.Y = int((zoom*float64(height)/2.8)*(i+imag(offset))) + height/2
```

Fix
```go
p.X = int((zoom*float64(width)/2.8)*(r+real(offset)) + width/2)
p.Y = int((zoom*float64(height)/2.8)*(i+imag(offset)) + height/2)
```

Created crosses by rounding coordinates numbers.

Public domain
-------------
I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
