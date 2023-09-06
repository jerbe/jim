package utils

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/20 01:46
  @describe :
*/

// GetClientIP 获取用户的真实IP
func GetClientIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-Real-IP")
	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	ip = r.Header.Get("X-Forward-For")
	for _, i := range strings.Split(ip, ",") {
		if net.ParseIP(i) != nil {
			return i, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	if net.ParseIP(ip) != nil {
		return ip, nil
	}

	return "", errors.New("no valid ip found")
}
