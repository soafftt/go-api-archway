package upstream

type UpstreamService struct {
	ServiceName   string `json:"service_name"`
	Authorization *struct {
		Algorithm string `json:"algorithm"`
		KeyData   string `json:"key_data"`
	} `json:"authorization:omitempty"`
	Resources     []*UpstreamResource `json:"resources"` // Note: field name matches API spec (resources)
	resourceIndex map[string]*UpstreamResource
}

// sudomain 을 기준으로 리소스를 조회합니다. 서브도메인이 없는 경우, 빈 문자열("")로 등록된 리소스를 조회합니다.
func (u *UpstreamService) LookupResourceDomain(subDomain string) (resource *UpstreamResource, isEmptyDomain bool) {
	resource, ok := u.resourceIndex[subDomain]
	if !ok {
		// 서브도메인이 없는 경우, 빈공백("")으로 등록된 리소스 조회
		resource, ok = u.resourceIndex[""]
		if !ok {
			return nil, false
		}
		return resource, true
	}
	return resource, false
}

// InitializeResourceIndex 는 서비스의 리소스들을 서브도메인 기준으로 빠르게 조회할 수 있도록 맵을 초기화합니다.
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
		u.resourceIndex[resource.SubDomain] = resource
	}
}
