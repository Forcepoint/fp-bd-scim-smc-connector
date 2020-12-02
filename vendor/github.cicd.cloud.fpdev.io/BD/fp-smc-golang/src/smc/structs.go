package smc

type UserCreation struct {
	Name                   string      `json:"name"`
	Enabled                bool        `json:"enabled"`
	AllowSudo              bool        `json:"allow_sudo"`
	ConsoleSuperuser       bool        `json:"console_superuser"`
	AllowedToLoginInShared bool        `json:"allowed_to_login_in_shared"`
	EngineTarget           []string    `json:"engine_target"`
	LocalAdmin             bool        `json:"local_admin"`
	Superuser              bool        `json:"superuser"`
	CanUseApi              bool        `json:"can_use_api"`
	Comment                interface{} `json:"comment"`
	AuthMethod             string      `json:"auth_method, omitempty"`
	LdapUser               string      `json:"ldap_user, omitempty"`
	//LdapGroup              string      `json:"ldap_group, omitempty"`
	Permissions map[string][]Permission `json:"permissions, omitempty"`
}

type LDAPUser struct {
	DaysLeft    int                 `json:"days_left, omitempty"`
	DisplayName string              `json:"display_name"`
	Key         int                 `json:"key, omitempty"`
	Link        []map[string]string `json:"link, omitempty"`
	Name        string              `json:"name"`
	ReadyOnly   bool                `json:"ready_only, omitempty"`
	System      bool                `json:"system, omitempty"`
	UniqueId    string              `json:"unique_id"`
}

type UserData struct {
	Name                   string      `json:"name"`
	Enabled                bool        `json:"enabled"`
	AllowSudo              bool        `json:"allow_sudo"`
	ConsoleSuperuser       bool        `json:"console_superuser"`
	AllowedToLoginInShared bool        `json:"allowed_to_login_in_shared"`
	EngineTarget           []string    `json:"engine_target"`
	LocalAdmin             bool        `json:"local_admin"`
	Superuser              bool        `json:"superuser"`
	CanUseApi              bool        `json:"can_use_api"`
	Comment                interface{} `json:"comment, omitempty"`
	AuthMethod             string      `json:"auth_method, omitempty"`
	//LdapGroup              string      `json:"ldap_group, omitempty"`
	LdapUser     string                  `json:"ldap_user, omitempty"`
	IsUserLocked bool                    `json:"is_user_locked"`
	Key          int                     `json:"key"`
	Permissions  map[string][]Permission `json:"permissions, omitempty"`
	ReadOnly     bool                    `json:"read_only"`
	System       bool                    `json:"system"`
	SystemKey    int                     `json:"system_key"`
}

type Permission struct {
	GrantedDomainRef string   `json:"granted_domain_ref"`
	GrantedElements  []string `json:"granted_elements"`
	// this is the role assigned to the admin
	RoleRef string `json:"role_ref"`
}

type ActiveDirectoryLDAPS struct {
	Address                   string   `json:"address"`
	BaseDn                    string   `json:"base_dn"`
	BindPassword              string   `json:"bind_password"`
	BindUserId                string   `json:"bind_user_id"`
	Name                      string   `json:"name"`
	Protocol                  string   `json:"protocol"`
	Port                      int      `json:"port"`
	Timeout                   int      `json:"timeout"`
	Retries                   int      `json:"retries"`
	AuthPort                  int      `json:"auth_port"`
	ClientCertBasedUserSearch string   `json:"client_cert_based_user_search"`
	GroupObjectClass          []string `json:"group_object_class"`
	PageSize                  int      `json:"page_size"`
	UserObjectClass           []string `json:"user_object_class"`
}

type ExternalLDAPUser struct {
	AuthMethod string   `json:"auth_method"`
	IsDefault  bool     `json:"isdefault"`
	LdapServer []string `json:"ldap_server"`
	Name       string   `json:"name"`
	ReadOnly   bool     `json:"read_only"`
	System     bool     `json:"system"`
}
