package cmd

type UserInfo struct {
	AuthMethod   string                  `json:"auth_method,omitempty"`
	Enable       bool                    `json:"enabled"`
	IsUserLocked bool                    `json:"is_user_locked"`
	LdapUser     string                  `json:"ldap_user,omitempty"`
	Link         []map[string]string     `json:"link"`
	Name         string                  `json:"name"`
	SuperUser    bool                    `json:"superuser"`
	Permissions  map[string][]Permission `json:"permissions,omitempty"`
}

type Permission struct {
	GrantedDomainRef string   `json:"granted_domain_ref"`
	GrantedElements  []string `json:"granted_elements"`
	// this is the role assigned to the admin
	RoleRef string `json:"role_ref"`
}
