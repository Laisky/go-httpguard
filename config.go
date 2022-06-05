package httpguard

var Config = &ConfigType{}

type ConfigType struct {
	Runtime   runtimeConfig         `mapstructure:"-"`
	Listen    string                `mapstructure:"listen"`
	Backend   string                `mapstructure:"backend"`
	JWTSecret string                `mapstructure:"jwt_secret"`
	Users     []configUser          `mapstructure:"users"`
	UsersMap  map[string]configUser `mapstructure:"-"`
}

func (c *ConfigType) Init() *ConfigType {
	c.UsersMap = make(map[string]configUser)
	for _, user := range c.Users {
		c.UsersMap[user.Username] = user
	}

	return c
}

type runtimeConfig struct {
	Debug      bool   `mapstructure:"-"`
	ConfigFile string `mapstructure:"-"`
}

type configUser struct {
	Username  string              `mapstructure:"username"`
	BasicAuth configUserBasicAuth `mapstructure:"basic_auth"`
	JWT       configUserJWT       `mapstructure:"jwt"`
	S3        configUserS3        `mapstructure:"s3"`
	Perms     configUserPerm      `mapstructure:"permission"`
}

type configUserPerm struct {
	// Read Get, Head, Options, etc.
	Read []string `mapstructure:"read"`
	// Write Post, Put, Patch, Delete, etc.
	Write []string `mapstructure:"write"`
}

type configUserBasicAuth struct {
	Enable   bool   `mapstructure:"enable"`
	Password string `mapstructure:"password"`
}

type configUserJWT struct {
	Enable bool `mapstructure:"enable"`
}

// configUserS3
//
// use username as appkey
type configUserS3 struct {
	Enable    bool   `mapstructure:"enable"`
	AppSecret string `mapstructure:"app_secret"`
}
