package utils

import (
	"errors"
	"strings"
)

func GetURLDomain(url string) (string, error) {
	urlNoHttp := strings.Replace(url, "https://", "", 1)
	urlNoHttpSplit := strings.Split(urlNoHttp, "/")

	if len(urlNoHttpSplit) == 0 {
		return "", errors.New("cannot extract domain from url " + url)
	}

	return urlNoHttpSplit[0], nil
}
