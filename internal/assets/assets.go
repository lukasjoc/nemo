package assets

import (
	"slices"
	"strings"

	"github.com/lukasjoc/nemo/internal"
)

type Asset struct {
	Name    string
	Sources [][]string
	Width   int
	Height  int
}

var cache = []Asset{}

func toTiles(sources []string) [][]string {
	t := [][]string{}
	for _, line := range sources {
		t = append(t, strings.Split(line, "\n"))
	}
	return t
}

func longestTile(a []string) int {
	n := 0
	for _, t := range a {
		if len(t) > n {
			n = len(t)
		}
	}
	return n
}

func newAsset(name string, sources []string) Asset {
	a := Asset{name, toTiles(sources), 0, 0}
	found := slices.ContainsFunc(cache, func(a Asset) bool {
		return a.Name == name
	})
	if found {
		panic("asset cannot be added because its already cached with the same name")
	}
	a.Width = longestTile(a.Sources[0])
	a.Height = len(a.Sources[0])
	cache = append(cache, a)
	return a
}

func Random() Asset { return internal.Choose(cache...) }

var Nemo = newAsset("nemo", []string{`
  __
\/ @\
/\__/
`, `
 __
/@ \/
\__/\
`,
})

var NemoJr = newAsset("nemo_jr", []string{`
  ___
\/ CC\
/\__~/
`, `
 ___
/CC \/
\~__/\
`,
})

var Runner = newAsset("runner", []string{`
>(#)@>
`, `
<@(#)<
`,
})

// INFO: AQ.. are taken from the asciiquarium program
var AQ0 = newAsset("aq0", []string{`
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
})

var AQ1 = newAsset("aq1", []string{`
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
})
