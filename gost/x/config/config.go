package config

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	v = viper.GetViper()
)

func init() {
	v.SetConfigName("gost")
	v.AddConfigPath("/etc/gost/")
	v.AddConfigPath("$HOME/.gost/")
	v.AddConfigPath(".")
}

var (
	global    = &Config{}
	globalMux sync.RWMutex
)

func Global() *Config {
	globalMux.RLock()
	defer globalMux.RUnlock()

	cfg := &Config{}
	*cfg = *global
	return cfg
}

func Set(c *Config) {
	globalMux.Lock()
	defer globalMux.Unlock()

	global = c
}

func OnUpdate(f func(c *Config) error) error {
	globalMux.Lock()
	defer globalMux.Unlock()

	return f(global)
}

type LogConfig struct {
	Output   string             `yaml:",omitempty" json:"output,omitempty"`
	Level    string             `yaml:",omitempty" json:"level,omitempty"`
	Format   string             `yaml:",omitempty" json:"format,omitempty"`
	Rotation *LogRotationConfig `yaml:",omitempty" json:"rotation,omitempty"`
}

type LogRotationConfig struct {
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `yaml:"maxSize,omitempty" json:"maxSize,omitempty"`
	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `yaml:"maxBackups,omitempty" json:"maxBackups,omitempty"`
	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time. The default is to use UTC
	// time.
	LocalTime bool `yaml:"localTime,omitempty" json:"localTime,omitempty"`
	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `yaml:"compress,omitempty" json:"compress,omitempty"`
}

type ProfilingConfig struct {
	Addr string `json:"addr"`
}

type APIConfig struct {
	Addr       string      `json:"addr"`
	PathPrefix string      `yaml:"pathPrefix,omitempty" json:"pathPrefix,omitempty"`
	AccessLog  bool        `yaml:"accesslog,omitempty" json:"accesslog,omitempty"`
	Auth       *AuthConfig `yaml:",omitempty" json:"auth,omitempty"`
	Auther     string      `yaml:",omitempty" json:"auther,omitempty"`
}

type MetricsConfig struct {
	Addr string `json:"addr"`
	Path string `yaml:",omitempty" json:"path,omitempty"`
}

type TLSConfig struct {
	CertFile   string `yaml:"certFile,omitempty" json:"certFile,omitempty"`
	KeyFile    string `yaml:"keyFile,omitempty" json:"keyFile,omitempty"`
	CAFile     string `yaml:"caFile,omitempty" json:"caFile,omitempty"`
	Secure     bool   `yaml:",omitempty" json:"secure,omitempty"`
	ServerName string `yaml:"serverName,omitempty" json:"serverName,omitempty"`

	// for auto-generated default certificate.
	Validity     time.Duration `yaml:",omitempty" json:"validity,omitempty"`
	CommonName   string        `yaml:"commonName,omitempty" json:"commonName,omitempty"`
	Organization string        `yaml:",omitempty" json:"organization,omitempty"`
}

type PluginConfig struct {
	Addr  string     `json:"addr"`
	TLS   *TLSConfig `yaml:",omitempty" json:"tls,omitempty"`
	Token string     `yaml:",omitempty" json:"token,omitempty"`
}

type AutherConfig struct {
	Name   string        `json:"name"`
	Auths  []*AuthConfig `yaml:",omitempty" json:"auths,omitempty"`
	Reload time.Duration `yaml:",omitempty" json:"reload,omitempty"`
	File   *FileLoader   `yaml:",omitempty" json:"file,omitempty"`
	Redis  *RedisLoader  `yaml:",omitempty" json:"redis,omitempty"`
	HTTP   *HTTPLoader   `yaml:"http,omitempty" json:"http,omitempty"`
	Plugin *PluginConfig `yaml:",omitempty" json:"plugin,omitempty"`
}

type AuthConfig struct {
	Username string `json:"username"`
	Password string `yaml:",omitempty" json:"password,omitempty"`
}

type SelectorConfig struct {
	Strategy    string        `json:"strategy"`
	MaxFails    int           `yaml:"maxFails" json:"maxFails"`
	FailTimeout time.Duration `yaml:"failTimeout" json:"failTimeout"`
}

type AdmissionConfig struct {
	Name string `json:"name"`
	// DEPRECATED by whitelist since beta.4
	Reverse   bool          `yaml:",omitempty" json:"reverse,omitempty"`
	Whitelist bool          `yaml:",omitempty" json:"whitelist,omitempty"`
	Matchers  []string      `yaml:",omitempty" json:"matchers,omitempty"`
	Reload    time.Duration `yaml:",omitempty" json:"reload,omitempty"`
	File      *FileLoader   `yaml:",omitempty" json:"file,omitempty"`
	Redis     *RedisLoader  `yaml:",omitempty" json:"redis,omitempty"`
	HTTP      *HTTPLoader   `yaml:"http,omitempty" json:"http,omitempty"`
	Plugin    *PluginConfig `yaml:",omitempty" json:"plugin,omitempty"`
}

type BypassConfig struct {
	Name string `json:"name"`
	// DEPRECATED by whitelist since beta.4
	Reverse   bool          `yaml:",omitempty" json:"reverse,omitempty"`
	Whitelist bool          `yaml:",omitempty" json:"whitelist,omitempty"`
	Matchers  []string      `yaml:",omitempty" json:"matchers,omitempty"`
	Reload    time.Duration `yaml:",omitempty" json:"reload,omitempty"`
	File      *FileLoader   `yaml:",omitempty" json:"file,omitempty"`
	Redis     *RedisLoader  `yaml:",omitempty" json:"redis,omitempty"`
	HTTP      *HTTPLoader   `yaml:"http,omitempty" json:"http,omitempty"`
	Plugin    *PluginConfig `yaml:",omitempty" json:"plugin,omitempty"`
}

type FileLoader struct {
	Path string `json:"path"`
}

type RedisLoader struct {
	Addr     string `json:"addr"`
	DB       int    `yaml:",omitempty" json:"db,omitempty"`
	Password string `yaml:",omitempty" json:"password,omitempty"`
	Key      string `yaml:",omitempty" json:"key,omitempty"`
	Type     string `yaml:",omitempty" json:"type,omitempty"`
}

type HTTPLoader struct {
	URL     string        `yaml:"url" json:"url"`
	Timeout time.Duration `yaml:",omitempty" json:"timeout,omitempty"`
}

type NameserverConfig struct {
	Addr     string        `json:"addr"`
	Chain    string        `yaml:",omitempty" json:"chain,omitempty"`
	Prefer   string        `yaml:",omitempty" json:"prefer,omitempty"`
	ClientIP string        `yaml:"clientIP,omitempty" json:"clientIP,omitempty"`
	Hostname string        `yaml:",omitempty" json:"hostname,omitempty"`
	TTL      time.Duration `yaml:",omitempty" json:"ttl,omitempty"`
	Timeout  time.Duration `yaml:",omitempty" json:"timeout,omitempty"`
}

type ResolverConfig struct {
	Name        string              `json:"name"`
	Nameservers []*NameserverConfig `yaml:",omitempty" json:"nameservers,omitempty"`
	Plugin      *PluginConfig       `yaml:",omitempty" json:"plugin,omitempty"`
}

type HostMappingConfig struct {
	IP       string   `json:"ip"`
	Hostname string   `json:"hostname"`
	Aliases  []string `yaml:",omitempty" json:"aliases,omitempty"`
}

type HostsConfig struct {
	Name     string               `json:"name"`
	Mappings []*HostMappingConfig `yaml:",omitempty" json:"mappings,omitempty"`
	Reload   time.Duration        `yaml:",omitempty" json:"reload,omitempty"`
	File     *FileLoader          `yaml:",omitempty" json:"file,omitempty"`
	Redis    *RedisLoader         `yaml:",omitempty" json:"redis,omitempty"`
	HTTP     *HTTPLoader          `yaml:"http,omitempty" json:"http,omitempty"`
	Plugin   *PluginConfig        `yaml:",omitempty" json:"plugin,omitempty"`
}

type IngressRuleConfig struct {
	Hostname string `json:"hostname"`
	Endpoint string `json:"endpoint"`
}

type IngressConfig struct {
	Name   string               `json:"name"`
	Rules  []*IngressRuleConfig `yaml:",omitempty" json:"rules,omitempty"`
	Reload time.Duration        `yaml:",omitempty" json:"reload,omitempty"`
	File   *FileLoader          `yaml:",omitempty" json:"file,omitempty"`
	Redis  *RedisLoader         `yaml:",omitempty" json:"redis,omitempty"`
	HTTP   *HTTPLoader          `yaml:"http,omitempty" json:"http,omitempty"`
	Plugin *PluginConfig        `yaml:",omitempty" json:"plugin,omitempty"`
}

type RecorderConfig struct {
	Name   string         `json:"name"`
	File   *FileRecorder  `yaml:",omitempty" json:"file,omitempty"`
	Redis  *RedisRecorder `yaml:",omitempty" json:"redis,omitempty"`
	Plugin *PluginConfig  `yaml:",omitempty" json:"plugin,omitempty"`
}

type FileRecorder struct {
	Path string `json:"path"`
	Sep  string `yaml:",omitempty" json:"sep,omitempty"`
}

type RedisRecorder struct {
	Addr     string `json:"addr"`
	DB       int    `yaml:",omitempty" json:"db,omitempty"`
	Password string `yaml:",omitempty" json:"password,omitempty"`
	Key      string `yaml:",omitempty" json:"key,omitempty"`
	Type     string `yaml:",omitempty" json:"type,omitempty"`
}

type RecorderObject struct {
	Name   string `json:"name"`
	Record string `json:"record"`
}

type LimiterConfig struct {
	Name   string        `json:"name"`
	Limits []string      `yaml:",omitempty" json:"limits,omitempty"`
	Reload time.Duration `yaml:",omitempty" json:"reload,omitempty"`
	File   *FileLoader   `yaml:",omitempty" json:"file,omitempty"`
	Redis  *RedisLoader  `yaml:",omitempty" json:"redis,omitempty"`
	HTTP   *HTTPLoader   `yaml:"http,omitempty" json:"http,omitempty"`
}

type ListenerConfig struct {
	Type       string            `json:"type"`
	Chain      string            `yaml:",omitempty" json:"chain,omitempty"`
	ChainGroup *ChainGroupConfig `yaml:"chainGroup,omitempty" json:"chainGroup,omitempty"`
	Auther     string            `yaml:",omitempty" json:"auther,omitempty"`
	Authers    []string          `yaml:",omitempty" json:"authers,omitempty"`
	Auth       *AuthConfig       `yaml:",omitempty" json:"auth,omitempty"`
	TLS        *TLSConfig        `yaml:",omitempty" json:"tls,omitempty"`
	Metadata   map[string]any    `yaml:",omitempty" json:"metadata,omitempty"`
}

type HandlerConfig struct {
	Type       string            `json:"type"`
	Retries    int               `yaml:",omitempty" json:"retries,omitempty"`
	Chain      string            `yaml:",omitempty" json:"chain,omitempty"`
	ChainGroup *ChainGroupConfig `yaml:"chainGroup,omitempty" json:"chainGroup,omitempty"`
	Auther     string            `yaml:",omitempty" json:"auther,omitempty"`
	Authers    []string          `yaml:",omitempty" json:"authers,omitempty"`
	Auth       *AuthConfig       `yaml:",omitempty" json:"auth,omitempty"`
	TLS        *TLSConfig        `yaml:",omitempty" json:"tls,omitempty"`
	Ingress    string            `yaml:",omitempty" json:"ingress,omitempty"`
	Metadata   map[string]any    `yaml:",omitempty" json:"metadata,omitempty"`
}

type ForwarderConfig struct {
	Name     string               `yaml:",omitempty" json:"name,omitempty"`
	Selector *SelectorConfig      `yaml:",omitempty" json:"selector,omitempty"`
	Nodes    []*ForwardNodeConfig `json:"nodes"`
	// DEPRECATED by nodes since beta.4
	Targets []string `yaml:",omitempty" json:"targets,omitempty"`
}

type ForwardNodeConfig struct {
	Name     string          `yaml:",omitempty" json:"name,omitempty"`
	Addr     string          `yaml:",omitempty" json:"addr,omitempty"`
	Host     string          `yaml:",omitempty" json:"host,omitempty"`
	Protocol string          `yaml:",omitempty" json:"protocol,omitempty"`
	Bypass   string          `yaml:",omitempty" json:"bypass,omitempty"`
	Bypasses []string        `yaml:",omitempty" json:"bypasses,omitempty"`
	HTTP     *HTTPNodeConfig `yaml:",omitempty" json:"http,omitempty"`
	TLS      *TLSNodeConfig  `yaml:",omitempty" json:"tls,omitempty"`
	Auth     *AuthConfig     `yaml:",omitempty" json:"auth,omitempty"`
}

type HTTPNodeConfig struct {
	Host   string            `yaml:",omitempty" json:"host,omitempty"`
	Header map[string]string `yaml:",omitempty" json:"header,omitempty"`
}

type TLSNodeConfig struct {
	ServerName string `yaml:"serverName,omitempty" json:"serverName,omitempty"`
	Secure     bool   `yaml:",omitempty" json:"secure,omitempty"`
}

type DialerConfig struct {
	Type     string         `json:"type"`
	Auth     *AuthConfig    `yaml:",omitempty" json:"auth,omitempty"`
	TLS      *TLSConfig     `yaml:",omitempty" json:"tls,omitempty"`
	Metadata map[string]any `yaml:",omitempty" json:"metadata,omitempty"`
}

type ConnectorConfig struct {
	Type     string         `json:"type"`
	Auth     *AuthConfig    `yaml:",omitempty" json:"auth,omitempty"`
	TLS      *TLSConfig     `yaml:",omitempty" json:"tls,omitempty"`
	Metadata map[string]any `yaml:",omitempty" json:"metadata,omitempty"`
}

type SockOptsConfig struct {
	Mark int `yaml:",omitempty" json:"mark,omitempty"`
}

type ServiceConfig struct {
	Name string `json:"name"`
	Addr string `yaml:",omitempty" json:"addr,omitempty"`
	// DEPRECATED by metadata.interface since beta.5
	Interface string `yaml:",omitempty" json:"interface,omitempty"`
	// DEPRECATED by metadata.so_mark since beta.5
	SockOpts   *SockOptsConfig   `yaml:"sockopts,omitempty" json:"sockopts,omitempty"`
	Admission  string            `yaml:",omitempty" json:"admission,omitempty"`
	Admissions []string          `yaml:",omitempty" json:"admissions,omitempty"`
	Bypass     string            `yaml:",omitempty" json:"bypass,omitempty"`
	Bypasses   []string          `yaml:",omitempty" json:"bypasses,omitempty"`
	Resolver   string            `yaml:",omitempty" json:"resolver,omitempty"`
	Hosts      string            `yaml:",omitempty" json:"hosts,omitempty"`
	Limiter    string            `yaml:",omitempty" json:"limiter,omitempty"`
	CLimiter   string            `yaml:"climiter,omitempty" json:"climiter,omitempty"`
	RLimiter   string            `yaml:"rlimiter,omitempty" json:"rlimiter,omitempty"`
	Recorders  []*RecorderObject `yaml:",omitempty" json:"recorders,omitempty"`
	Handler    *HandlerConfig    `yaml:",omitempty" json:"handler,omitempty"`
	Listener   *ListenerConfig   `yaml:",omitempty" json:"listener,omitempty"`
	Forwarder  *ForwarderConfig  `yaml:",omitempty" json:"forwarder,omitempty"`
	Metadata   map[string]any    `yaml:",omitempty" json:"metadata,omitempty"`
}

type ChainConfig struct {
	Name string `json:"name"`
	// REMOVED since beta.6
	// Selector *SelectorConfig `yaml:",omitempty" json:"selector,omitempty"`
	Hops     []*HopConfig   `json:"hops"`
	Metadata map[string]any `yaml:",omitempty" json:"metadata,omitempty"`
}

type ChainGroupConfig struct {
	Chains   []string        `yaml:",omitempty" json:"chains,omitempty"`
	Selector *SelectorConfig `yaml:",omitempty" json:"selector,omitempty"`
}

type HopConfig struct {
	Name      string          `json:"name"`
	Interface string          `yaml:",omitempty" json:"interface,omitempty"`
	SockOpts  *SockOptsConfig `yaml:"sockopts,omitempty" json:"sockopts,omitempty"`
	Selector  *SelectorConfig `yaml:",omitempty" json:"selector,omitempty"`
	Bypass    string          `yaml:",omitempty" json:"bypass,omitempty"`
	Bypasses  []string        `yaml:",omitempty" json:"bypasses,omitempty"`
	Resolver  string          `yaml:",omitempty" json:"resolver,omitempty"`
	Hosts     string          `yaml:",omitempty" json:"hosts,omitempty"`
	Nodes     []*NodeConfig   `yaml:",omitempty" json:"nodes,omitempty"`
}

type NodeConfig struct {
	Name      string           `json:"name"`
	Addr      string           `yaml:",omitempty" json:"addr,omitempty"`
	Host      string           `yaml:",omitempty" json:"host,omitempty"`
	Protocol  string           `yaml:",omitempty" json:"protocol,omitempty"`
	Interface string           `yaml:",omitempty" json:"interface,omitempty"`
	SockOpts  *SockOptsConfig  `yaml:"sockopts,omitempty" json:"sockopts,omitempty"`
	Bypass    string           `yaml:",omitempty" json:"bypass,omitempty"`
	Bypasses  []string         `yaml:",omitempty" json:"bypasses,omitempty"`
	Resolver  string           `yaml:",omitempty" json:"resolver,omitempty"`
	Hosts     string           `yaml:",omitempty" json:"hosts,omitempty"`
	Connector *ConnectorConfig `yaml:",omitempty" json:"connector,omitempty"`
	Dialer    *DialerConfig    `yaml:",omitempty" json:"dialer,omitempty"`
	Metadata  map[string]any   `yaml:",omitempty" json:"metadata,omitempty"`
	HTTP      *HTTPNodeConfig  `yaml:",omitempty" json:"http,omitempty"`
	TLS       *TLSNodeConfig   `yaml:",omitempty" json:"tls,omitempty"`
	Auth      *AuthConfig      `yaml:",omitempty" json:"auth,omitempty"`
}

type Config struct {
	Services   []*ServiceConfig   `json:"services"`
	Chains     []*ChainConfig     `yaml:",omitempty" json:"chains,omitempty"`
	Hops       []*HopConfig       `yaml:",omitempty" json:"hops,omitempty"`
	Authers    []*AutherConfig    `yaml:",omitempty" json:"authers,omitempty"`
	Admissions []*AdmissionConfig `yaml:",omitempty" json:"admissions,omitempty"`
	Bypasses   []*BypassConfig    `yaml:",omitempty" json:"bypasses,omitempty"`
	Resolvers  []*ResolverConfig  `yaml:",omitempty" json:"resolvers,omitempty"`
	Hosts      []*HostsConfig     `yaml:",omitempty" json:"hosts,omitempty"`
	Ingresses  []*IngressConfig   `yaml:",omitempty" json:"ingresses,omitempty"`
	Recorders  []*RecorderConfig  `yaml:",omitempty" json:"recorders,omitempty"`
	Limiters   []*LimiterConfig   `yaml:",omitempty" json:"limiters,omitempty"`
	CLimiters  []*LimiterConfig   `yaml:"climiters,omitempty" json:"climiters,omitempty"`
	RLimiters  []*LimiterConfig   `yaml:"rlimiters,omitempty" json:"rlimiters,omitempty"`
	TLS        *TLSConfig         `yaml:",omitempty" json:"tls,omitempty"`
	Log        *LogConfig         `yaml:",omitempty" json:"log,omitempty"`
	Profiling  *ProfilingConfig   `yaml:",omitempty" json:"profiling,omitempty"`
	API        *APIConfig         `yaml:",omitempty" json:"api,omitempty"`
	Metrics    *MetricsConfig     `yaml:",omitempty" json:"metrics,omitempty"`
}

func (c *Config) Load() error {
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(c)
}

func (c *Config) Read(r io.Reader) error {
	if err := v.ReadConfig(r); err != nil {
		return err
	}

	return v.Unmarshal(c)
}

func (c *Config) ReadFile(file string) error {
	v.SetConfigFile(file)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return v.Unmarshal(c)
}

func (c *Config) Write(w io.Writer, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(c)
		return nil
	case "yaml":
		fallthrough
	default:
		enc := yaml.NewEncoder(w)
		defer enc.Close()
		enc.SetIndent(2)

		return enc.Encode(c)
	}
}
