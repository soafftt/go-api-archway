package upstream

type UpstreamService struct {
	Service string                  `json:"servie"`
	HostMap map[string]*UptreamHost `json:"host"`
}

func (u *UpstreamService) LookupHost(host string) *UptreamHost {
	hostmap, ok := u.HostMap[host]
	if !ok {
		return nil
	}

	return hostmap
}
