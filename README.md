# Wasabi

[![Maintainability](https://api.codeclimate.com/v1/badges/159615e96c52b724a4ae/maintainability)](https://codeclimate.com/github/karlek/wasabi/maintainability)

Wasabi is a renderer of buddhabrot and its family members. It used to share its name with a Japanese aesthetic called [Wabi-sabi](https://en.wikipedia.org/wiki/Wabi-sabi). Referencing the impossibility of creating the real buddhabrot and learning to accept the beauty in reality and its flaws. However, the affectionate nickname wasabi soon replaced it.

## Showcase

> To the left, an original buddhabrot and to the right an anti-buddhabrot.

<p align="center">
	<img src="https://public.karlek.io/original.jpg" width="45.0%">
	<img src="https://public.karlek.io/anti.jpg" width="45.0%">
</p>

> To the left an image of the linear calculation path rendering technique, and
> to the right a second degree bezier interpolation.

<p align="center">
	<img src="https://public.karlek.io/calc.jpg" width="45.0%">
	<img src="https://public.karlek.io/bezier.jpg" width="45.0%">
</p>

> It's also possible to plot the other capital planes of the complex space. I
> created a new method to visualize angles between points inside orbits

<p align="center">
	<img src="https://public.karlek.io/1-zrcr.jpg" width="45.0%">
	<img src="https://public.karlek.io/2-angles.jpg" width="45.0%">
</p>

> There are multiple ways to tweak the complex functions.

<p align="center"><img src="https://public.karlek.io/inwards/inwards.jpg" width="45.0%"></p>
	
> The project has rendered a few visually interesting bugs.

<p align="center">
	<img src="https://public.karlek.io/bug.jpg" width="45.0%">
	<img src="https://public.karlek.io/race-condition.jpg" width="45.0%">
</p>

> To the left an point orbit trap around origo, and to the right off-center.

<p align="center">
	<img src="https://public.karlek.io/orbit-trap.jpg" width="45.0%">
	<img src="https://public.karlek.io/point-trap.jpg" width="45.0%">
</p>

> Histogram merging, i.e plotting multiple renders on the same canvas.

<p align="center"><img src="https://public.karlek.io/histogram-merge.jpg" width="45.0%"></p>

> One of the more famous of my renders ;)

<p align="center"><img src="https://public.karlek.io/magma.png" width="45.0%"></p>

## Features

* Calculating the original, anti- and primitive- buddhabrot.
* Exploring the different planes of Zr, Zi, Cr and Ci.
* Modular design for easier exploration of the complex function space.
* Histogram equalization functions to control image exposure.
* Cache histograms for faster exposure tweaking.
* Parallel computing for all heavy calculations.
* Plot calculation-paths. Credits to Raka Jovanovic and Milan Tuba (ISSN: 1109-2750).
* Plot orbit angle distribution.
* Hand optimized assembly(!) for generating random complex points. Thank you [7i](https://github.com/7i)!

>It should be noted that speed in random number generating algorithms competes
>with the necessity of having a random distribution. If you know of a way to
>benchmark randomness as well as speed, please create an issue!

![Benchmark](https://public.karlek.io/benchmark.png)

## Install

```fish
$ go build github.com/karlek/wasabi/cmd/wasabi
```

## Run

```fish
# Be sure to limit the memory usage beforehand; wasabi is greedy little devil.
$ ulimit -Sv 4000000 # Where the number is the memory in kB.
$ wasabi blueprint.json
```

## Tips

For doing animations I recommend writing a simple shell script. I use `jq` to
iteratively update the blueprint and `fish` as my shell of preference. My
scripts usually looks like this:

```fish
# Animation of the real coefficient.
for i in (seq -1 0.1 1)
	jq ".realCoefficient = $i" < wimm.json > /tmp/a.json
	wasabi -out "$i" /tmp/a.json 
end
```

## Contributing

The easiest way to contribute is to find a new interesting complex function or
z/c-sampling strategy. Please make a pull request with a pretty image and the
`blueprint.json`.

Public domain
-------------
I hereby release this code into the [public domain](https://creativecommons.org/publicdomain/zero/1.0/).
