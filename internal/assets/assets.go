package assets

import (
	"fmt"
	"strings"

	"github.com/lukasjoc/nemo/internal"
)

type Asset struct {
	Group   string
	Sources [][]string
	Width   int
	Height  int
}

var cache = map[string][]Asset{}

func toTiles(sources []string) [][]string {
	t := [][]string{}
	for _, line := range sources {
		t = append(t, strings.Split(line, "\n"))
	}
	return t
}

func longestTile(a []string) int {
	n := -1
	for _, t := range a {
		if len(t) > n {
			n = len(t)
		}
	}
	return n
}

func newAsset(group string, sources ...string) Asset {
	tiles := toTiles(sources)
	a := Asset{
		Group:   group,
		Sources: tiles,
		Width:   longestTile(tiles[0]),
		Height:  len(tiles[0]),
	}
	if _, ok := cache[a.Group]; !ok {
		cache[a.Group] = []Asset{}
	}
	cache[a.Group] = append(cache[a.Group], a)
	return a
}

func Random(group string) Asset {
	if _, ok := cache[group]; !ok {
		panic(fmt.Sprintf("group with name `%s` doesnt exist", group))
	}
	return internal.Choose(cache[group]...)
}

var _ = newAsset("fish", `
  __
\/ @\
/\__/
`, `
 __
/@ \/
\__/\
`,
)

var _ = newAsset("fish", `
  ___
\/ CC\
/\__~/
`, `
 ___
/CC \/
\~__/\
`,
)

var _ = newAsset("fish", `>(#)@>`, `<@(#)<`)

// INFO: these 2 are taken from the asciiquarium program
var _ = newAsset("fish", `
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
)

var _ = newAsset("fish", `
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
)

var _ = newAsset("bubble", `*`)
var _ = newAsset("bubble", `.`)
var _ = newAsset("bubble", `o`)
var _ = newAsset("bubble", `O`)
