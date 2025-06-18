package queue

import (
	"fmt"
	"regexp"
	"strings"
)

var APP_NAMESPACES = map[string]struct{}{"workflows": {}, "comfy": {}}

type AppID struct {
	Owner     string
	Alias     string
	Path      string
	Namespace string
}

func (a AppID) URLString() string {
	var b strings.Builder
	if a.Namespace != "" {
		b.WriteString(a.Namespace)
		b.WriteString("/")
	}
	b.WriteString(a.Owner)
	b.WriteString("/")
	b.WriteString(a.Alias)
	return b.String()
}

func ensureAppIDFormat(str string) (string, error) {
	parts := strings.Split(str, "/")
	if len(parts) > 1 {
		return str, nil
	}
	re := regexp.MustCompile(`^([0-9]+)-([a-zA-Z0-9-]+)$`)
	match := re.FindStringSubmatch(str)
	if len(match) == 3 {
		return fmt.Sprintf("%s/%s", match[1], match[2]), nil
	}
	return "", fmt.Errorf("invalid app id: %s. must be in the format <appOwner>/<appId>", str)
}

func AppIDFromEndpoint(str string) (*AppID, error) {
	normalizedID, err := ensureAppIDFormat(str)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(normalizedID, "/")
	ret := new(AppID)
	if _, ok := APP_NAMESPACES[parts[0]]; ok {
		ret.Namespace = parts[0]
		ret.Owner = parts[1]
		ret.Alias = parts[2]
		if l := len(parts); l > 3 {
			ret.Path = strings.Join(parts[3:l], "/")
		}
	} else {
		ret.Owner = parts[0]
		ret.Alias = parts[1]
		if l := len(parts); l > 2 {
			ret.Path = strings.Join(parts[2:l], "/")
		}
	}
	return ret, nil
}
