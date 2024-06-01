package assets

import (
	"fmt"
	"strings"
)

type Asset = string

func Load(asset Asset) []string {
	tiles := strings.Split(asset, "\n")
	for _, tile := range tiles {
		tile = fmt.Sprintf(` %s `, tile)
	}
	return tiles
}

var Normie Asset = `
   __
 \/ @\
 /\__/
`

var Other Asset = `
          \:. 
 \;,   ,;\\\\\,, 
   \\\\\;;:::::::o 
   ///;;::::::::< 
  /;   \/////// 
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
