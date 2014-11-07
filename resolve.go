package libkb

import (
	"fmt"
	"github.com/keybase/go-jsonw"
)

//==================================================================

func (c *UserCache) GetResolution(key string) (string, bool, error) {
	res, found := c.resolveCache[key]
	if found {
		return res.res, true, res.err
	} else {
		return "", false, nil
	}
}

func (c *UserCache) PutResolution(key string, val string, err error) {
	c.resolveCache[key] = ResolveResult{val, err}
}

//==================================================================

type ResolveResult struct {
	name string
	uid  UID
	body *jsonw.Wrapper
	err  error
}

func ResolveUsername(input string) (res ResolveResult) {
	G.Log.Debug("+ Resolving username %s", input)
	var au AssertionUrl
	if au, res.err = ParseAssertionUrl(input, false); res.err != nil {
		return
	}
	if res.output, res.err = _resolveUsername(au); res.err != nil {
		return
	}
	return
}

func ResolveUsernameKeyValuePair(key, value string) (string, error) {
	var output string
	G.Log.Debug("+ Resolve username (%s,%s)", key, value)

	au, err := ParseAssertionUrlKeyValue(key, value, false)
	if err == nil {
		output, err = _resolveUsername(au)
	}

	G.Log.Debug("- Resolve username (%s,%s) -> %s", key, value, output)
	return output, nil
}

func _resolveUsername(au AssertionUrl) (out string, err error) {
	// A standard keybase name, so it's already resolved
	if au.IsKeybase() {
		return au.GetValue(), nil
	}

	if out, found, err := G.UserCache.GetResolution(au.CacheKey()); found {
		return out, err
	}

	out, err = __resolveUsername(au)
	G.UserCache.PutResolution(au.CacheKey(), out, err)

	return out, err
}

func __resolveUsername(au AssertionUrl) (out string, err error) {

	key, val, err := au.ToLookup()
	if err != nil {
		return
	}

	res, err := G.API.Get(ApiArg{
		Endpoint:    "user/lookup",
		NeedSession: false,
		Args:        HttpArgsFromKeyValuePair(key, S{val}),
	})

	if err != nil {
		return
	}

	them, err := res.Body.AtKey("them").ToArray()
	if err != nil {
		return
	}
	l, err := them.Len()
	if err != nil {
		return
	}

	G.UserCache.CacheServerGetVector(them)

	if l == 0 {
		err = fmt.Errorf("No resolution found for %s", au.ToString())
	} else if l > 1 {
		err = fmt.Errorf("Identity '%s' is ambiguous", au.ToString())
	} else {
		out, err = them.AtIndex(0).AtKey("basics").AtKey("username").GetString()
	}

	return
}
