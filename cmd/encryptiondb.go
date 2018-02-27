package cmd

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio/pkg/wildcard"
)

// encryption keys db
type encryptionKeysDB struct {
}

// Match function matches wild cards in 'pattern' for resource.
func resourceMatch(pattern, resource string) bool {
	if runtime.GOOS == "windows" {
		// For windows specifically make sure we are case insensitive.
		return wildcard.Match(strings.ToLower(pattern), strings.ToLower(resource))
	}
	return wildcard.Match(pattern, resource)
}

/*
// getSSEKeyMap returns the prefix to sse-c key mapping for a particular alias.
func getSSEKeyMap(alias, ssekeys string) (sseKeyMap map[string]string, err *probe.Error) {
	sseKeyMap = make(map[string]string)
	if ssekeys == "" {
		return sseKeyMap, nil
	}
	fmt.Println("inside getSSEKeyMap urlalias==", alias, "sseKeys ===>", ssekeys)
	alias = strings.TrimSuffix(alias, string(filepath.Separator))
	fields := strings.Split(ssekeys, ":")
	for _, field := range fields {
		pair := strings.Split(field, "=")
		prefix := strings.TrimPrefix(pair[0], alias)
		prefix = strings.TrimPrefix(prefix, string(filepath.Separator))
		fmt.Println("prefix ==", prefix, "p0=", pair[0], " p1=", pair[1], "alias+=", alias+string(filepath.Separator))
		if len(pair) != 2  {
			return nil, probe.NewError(errors.New("sse-c prefix should be of the form prefix=key"))
		}
		if strings.HasPrefix(pair[0],p)
		sseKeyMap[prefix] = pair[1]
	}
	return
}

func getSSEKey(urlPath string, sseKeyMap map[string]string) string {
	for prefix, sseKey := range sseKeyMap {
		_, e := filepath.Match(prefix, urlPath)
		if e == nil {
			return sseKey
		}
	}
	return ""
}
*/

type prefixSSEPair struct {
	prefix string
	sseKey string
}

func parseEncryptionKeys(ssekeys string) (encMap map[string][]prefixSSEPair, err *probe.Error) {
	fmt.Println("------- parse satrt.....")
	encMap = make(map[string][]prefixSSEPair)
	if ssekeys == "" {
		return
	}

	fields := strings.Split(ssekeys, ":")
	for _, field := range fields {
		pair := strings.Split(field, "=")
		if len(pair) != 2 {
			return nil, probe.NewError(errors.New("sse-c prefix should be of the form prefix=key"))
		}
		alias, path := url2Alias(pair[0])
		if len(pair[1]) != 32 {
			return nil, probe.NewError(errors.New("sse-c key should be "))

		}
		fmt.Println("extracted alias and path===>", alias, path)
		if _, ok := encMap[alias]; !ok {
			encMap[alias] = make([]prefixSSEPair, 1)
		}
		ps := prefixSSEPair{prefix: path, sseKey: pair[1]}
		encMap[alias] = append(encMap[alias], ps)
	}
	fmt.Println("------- parse end.....")

	return
}
