package assets

import (
	"errors"
	"strings"
)

var Nemo = []string{`
  __
\/ @\
/\__/
`, `
 __
/@ \/
\__/\
`,
}

var NemoJr = []string{`
  ___
\/ CC\
/\__~/
`, `
 ___
/CC \/
\~__/\
`,
}

var Runner = []string{`
>(#)@>
`, `
<@(#)<
`,
}

// INFO: AQ.. are taken from the asciiquarium program
var AQ0 = []string{`
       \
     ...\..,
\  /'       \
 >=     (  ' >
/  \      / /
    /"'"'/''
`, `
      /
  ,../...
 /       '\  /
< '  )     =<
 \ \      /  \
  ''\'"'"\
`,
}

var AQ1 = []string{`
    \
\ /--\
>=  (o>
/ \__/
    /
`, `
  /
 /--\ /
<o)  =<
 \__/ \
  \
`,
}

type Tiles []string

func LoadTiles(a []string) ([]Tiles, error) {
	tiles := []Tiles{}
	if len(a) != 2 {
		return nil, errors.New("invalid asset layout")
	}
	for _, vers := range a {
		tiles = append(tiles, strings.Split(vers, "\n"))
	}
	return tiles, nil
}
