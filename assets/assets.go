package assets

import (
	"strings"
)

type Asset = string

func Load(asset Asset) []string {
	tiles := strings.Split(asset, "\n")
	return tiles
}

var NormieR Asset = `
  __
\/ @\
/\__/
`

var NormieL Asset = `
 __
/@ \/
\__/\
`

var Other Asset = `
          :.
\;,   ,;\\\\\,_,
 \\\\\;;:::::::0|
 ///;;::::::::</
/;    \///////
`

var Invader Asset = `
     ;;
>///;;///@(@>
      ;;;
`

var Coolio Asset = `

|\___//^^^/\
|         00\
|/---\\\\|/~|
`

var Nilly Asset = `
 \\    /-\     /-\ 
 \ \---\--\,,,,\,,\,,-@@\ 
 / /---,--,,,,,,,,,,,//~/ 
 //    /,,,/     /,,/ 
`

var NemoJr Asset = `
   ___
|\/ $$\
|/\__-/
`

var Happydoo Asset = `
><;;\#/?_?>
`

var Runner Asset = `
>(_)@>
`

var Wave Asset = `
	~~~~~~~~~~~~~~~~~~~~~~~~ ~~~~~~~~~~~~~~~~~~
	~~~~~~~~~ ~~~~~~~~~~~~~~~~~~~~~~ ~~~~~~~~~~
	~~~ ~~~~~~~~~~~~~~~~~ ~~~~~~~~~~~~~~~~~~~~~
`
