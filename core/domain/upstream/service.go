package upstream

type UpstreamService struct {
	ServiceName   string `json:"service_name"`
	Authorization *struct {
		Algorithm string `json:"algorithm"`
		KeyData   string `json:"key_data"`
		UserKey   string `json:"user_key"`
	} `json:"authorization,omitempty"`
	Resources     []*UpstreamResource `json:"resources"` // Note: field name matches API spec (resources)
	resourceIndex map[string]*UpstreamResource
}

// LookupResourceDomain은 domain을 기준으로 리소스를 조회합니다.
func (u *UpstreamService) LookupResourceDomain(domain string) (resource *UpstreamResource, isEmptyDomain bool) {
	resource, ok := u.resourceIndex[domain]
	if !ok {
		// 일치하는 domain이 없으면 빈 문자열로 등록된 fallback 리소스를 조회합니다.
		resource, ok = u.resourceIndex[""]
		if !ok {
			return nil, false
		}
		return resource, true
	}
	return resource, false
}

// InitializeResourceIndex는 서비스의 리소스를 domain 기준으로 조회할 수 있도록 인덱스를 초기화합니다.
// JSON Unmarshal 후 반드시 호출되어야 합니다.
func (u *UpstreamService) InitializeResourceIndex() {
	if u.resourceIndex != nil {
		return
	}

	u.resourceIndex = make(map[string]*UpstreamResource, len(u.Resources))
	for _, resource := range u.Resources {
		if resource == nil {
			continue
		}
		resource.InitializeRouter()
		u.resourceIndex[resource.Domain] = resource
	}
}
