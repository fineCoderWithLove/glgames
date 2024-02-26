package discovery

import "fmt"

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Weight  int    `json:"weight"`
	Version string `json:"version"`
	Ttl     int64  `json:"ttl"`
}

func (s Server) BuildRegisterKey() string {
	if len(s.Version) == 0 {
		// user
		return fmt.Sprintf("/%s/%s", s.Name, s.Addr)
	}
	//user/v1
	return fmt.Sprintf("/%s/%s/%s", s.Name, s.Version, s.Addr)
}
