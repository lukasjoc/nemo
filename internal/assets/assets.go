package assets

import (
	"slices"
	"strings"

	"github.com/lukasjoc/nemo/internal"
)

type Asset struct {
	Name    string
	Sources [][]string
}

var cache = []Asset{}

func toTiles(sources []string) [][]string {
	t := [][]string{}
	for _, line := range sources {
		t = append(t, strings.Split(line, "\n"))
	}
	return t
}

func newAsset(name string, sources []string) Asset {
	a := Asset{name, toTiles(sources)}
	found := slices.ContainsFunc(cache, func(a Asset) bool {
		return a.Name == name
	})
	if found {
		panic("asset cannot be added because its already cached with the same name")
	}
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
