package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	Version string = "4.0.0"
)

type CommonConfig struct {
	ListenIp            string
	ListenPort          int
	ListenUsername      string
	ListenPassword      string
	ListenVisible       bool
	ListenDebuginfo     bool
	GaeEnable           bool
	GaeAppids           []string
	GaePassword         string
	GaePath             string
	GaeMode             string
	GaeIpv6             bool
	GaeWindow           int
	GaeKeepalive        bool
	GaeCachesock        bool
	GaeHeadfirst        bool
	GaeObfuscate        bool
	GaeValidate         bool
	GaeTransport        bool
	GaeOptions          []string
	GaeRegions          []string
	GaeSslversion       string
	GaePagespeed        bool
	PacEnable           bool
	PacIp               string
	PacPort             int
	PacFile             string
	PacGfwlist          string
	PacAdblock          string
	PacAdmode           int
	PacExpired          int
	PhpEnable           bool
	PhpListen           string
	PhpPassword         string
	PhpCrlf             bool
	PhpValidate         bool
	PhpKeepalive        bool
	PhpFetchserver      []string
	PhpHosts            []string
	VpsEnable           bool
	VpsListen           string
	VpsFetchserver      []string
	ProxyEnable         bool
	ProxyAutodetect     bool
	ProxyHost           string
	ProxyPort           int
	ProxyUsername       string
	ProxyPasswrod       string
	AutorangeHosts      []string
	AutorangeEndswith   []string
	AutorangeNoendswith []string
	AutorangeMaxsize    int
	AutorangeWaitsize   int
	AutorangeBufsize    int
	AutorangeThreads    int
	FetchmaxLocal       int
	FetchmaxServer      int
	DnsEnable           bool
	DnsListen           string
	DnsServers          []string
	UseragentEnable     bool
	UseragentString     string
	LoveEnable          bool
	LoveTip             []string

	IplistMap         map[string][]string
	IplistFixed       []string
	HostMap           map[string]string
	UrlRewriteMap     map[string]string
	CrlfSites         []string
	NocrlfSites       []string
	ForcehttpsSites   []string
	NoforcehttpsSites map[string]struct{}
	FakehttpsSites    []string
	NofakehttpsSites  map[string]struct{}
	WithGAESites      []string
	WithPHPSites      []string
	WithVPSSites      []string
}

type GoConfig map[string]map[string]string

func ReadConfig(filename string) (GoConfig, error) {
	config := GoConfig{}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.Replace(string(b), "\r\n", "\n", -1), "\n")
	section := "default"
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ';' {
			continue
		}
		if line[0] == '[' && line[len(line)-1] == ']' {
			section = line[1 : len(line)-1]
			config[section] = make(map[string]string, 0)
		} else if strings.Contains(line, " =") {
			items := strings.SplitN(line, " =", 2)
			if _, ok := config[section]; !ok {
				config[section] = make(map[string]string, 0)
			}
			config[section][strings.TrimSpace(items[0])] = strings.TrimSpace(items[1])
		}
	}
	return config, nil
}

func (c GoConfig) GetOptions(section string) []string {
	options := make([]string, 0)
	if _, ok := c[section]; !ok {
		panic(fmt.Errorf("section %#v not exists", section))
	}
	for option, _ := range c[section] {
		options = append(options, option)
	}
	return options
}

func (c GoConfig) GetString(section string, option string) string {
	if _, ok := c[section]; !ok {
		panic(fmt.Errorf("section %#v not exists", section))
	}
	if _, ok := c[section][option]; !ok {
		panic(fmt.Errorf("section %#v option %#v not exists", section, option))
	}
	return c[section][option]
}

func (c GoConfig) GetStrings(section string, option string) []string {
	return regexp.MustCompile(`[|,]`).Split(c.GetString(section, option), -1)
}

func (c GoConfig) GetInt(section string, option string) int {
	value, err := strconv.Atoi(c.GetString(section, option))
	if err != nil {
		panic(err)
	}
	return value
}

func (c GoConfig) GetBool(section string, option string) bool {
	switch strings.ToLower(c.GetString(section, option)) {
	case "1":
		return true
	case "true":
		return true
	case "t":
		return true
	default:
		return false
	}
}

func (cc *CommonConfig) WriteSummary(w io.Writer) error {
	fmt.Fprintf(w, "------------------------------------------------------\n")
	fmt.Fprintf(w, "GoAgent Version    : %s (golang/%s)\n", Version, runtime.Version())
	fmt.Fprintf(w, "Listen Address     : %s\n", net.JoinHostPort(cc.ListenIp, strconv.Itoa(cc.ListenPort)))
	if cc.ProxyEnable {
		fmt.Fprintf(w, "Local Proxy        : %s\n", net.JoinHostPort(cc.ProxyHost, strconv.Itoa(cc.ProxyPort)))
	}
	if cc.ProxyEnable {
		fmt.Fprintf(w, "Debug INFO         : %v\n", cc.ListenDebuginfo)
	}
	fmt.Fprintf(w, "GAE Mode           : %s\n", cc.GaeMode)
	fmt.Fprintf(w, "GAE IPv6           : %v\n", cc.GaeIpv6)
	fmt.Fprintf(w, "GAE APPID          : %s\n", strings.Join(cc.GaeAppids, "|"))
	if cc.GaeValidate {
		fmt.Fprintf(w, "GAE Validate       : %v\n", cc.GaeValidate)
	}
	if cc.GaeObfuscate {
		fmt.Fprintf(w, "GAE Obfuscate      : %v\n", cc.GaeObfuscate)
	}
	if cc.GaeObfuscate {
		fmt.Fprintf(w, "GAE Obfuscate      : %v\n", cc.GaeObfuscate)
	}
	if cc.PacEnable {
		fmt.Fprintf(w, "Pac Server         : http://%s/%s\n", net.JoinHostPort(cc.PacIp, strconv.Itoa(cc.PacPort)), cc.PacFile)
		if pacFile, err := filepath.Abs(cc.PacFile); err == nil {
			fmt.Fprintf(w, "Pac File           : file://%s\n", pacFile)
		}
	}
	if cc.PhpEnable {
		fmt.Fprintf(w, "PHP Listen         : %s\n", cc.PhpListen)
		fmt.Fprintf(w, "PHP FetchServer    : %s\n", strings.Join(cc.PhpFetchserver, "|"))
	}
	if cc.VpsEnable {
		fmt.Fprintf(w, "VPS Listen         : %s\n", cc.VpsListen)
		fmt.Fprintf(w, "VPS FetchServer    : %s\n", strings.Join(cc.VpsFetchserver, "|"))
	}
	if cc.VpsEnable {
		fmt.Fprintf(w, "VPS Listen         : %s\n", cc.VpsListen)
		fmt.Fprintf(w, "VPS FetchServer    : %s", strings.Join(cc.VpsFetchserver, "|"))
	}
	if cc.DnsEnable {
		fmt.Fprintf(w, "DNS Listen         : %s\n", cc.DnsListen)
		fmt.Fprintf(w, "DNS FetchServer    : %s\n", strings.Join(cc.DnsServers, "|"))
	}
	fmt.Fprintf(w, "------------------------------------------------------\n")
	return nil
}

func ReadConfigFile(filename string) (*CommonConfig, error) {
	c, err := ReadConfig(filename)
	if err != nil {
		return nil, err
	}

	cc := &CommonConfig{}
	cc.ListenIp = c.GetString("listen", "ip")
	cc.ListenPort = c.GetInt("listen", "port")
	cc.ListenUsername = c.GetString("listen", "username")
	cc.ListenPassword = c.GetString("listen", "password")
	cc.ListenVisible = c.GetBool("listen", "visible")
	cc.ListenDebuginfo = c.GetBool("listen", "debuginfo")

	cc.GaeEnable = c.GetBool("gae", "enable")
	cc.GaeAppids = c.GetStrings("gae", "appid")
	cc.GaePassword = c.GetString("gae", "password")
	cc.GaePath = c.GetString("gae", "path")
	cc.GaeMode = c.GetString("gae", "mode")
	cc.GaeIpv6 = c.GetBool("gae", "ipv6")
	cc.GaeSslversion = c.GetString("gae", "sslversion")
	cc.GaeWindow = c.GetInt("gae", "window")
	cc.GaeCachesock = c.GetBool("gae", "cachesock")
	cc.GaeHeadfirst = c.GetBool("gae", "headfirst")
	cc.GaeKeepalive = c.GetBool("gae", "keepalive")
	cc.GaeObfuscate = c.GetBool("gae", "obfuscate")
	cc.GaeValidate = c.GetBool("gae", "validate")
	cc.GaeTransport = c.GetBool("gae", "transport")
	cc.GaeOptions = c.GetStrings("gae", "options")
	cc.GaeRegions = c.GetStrings("gae", "regions")

	cc.PacEnable = c.GetBool("pac", "enable")
	cc.PacIp = c.GetString("pac", "ip")
	cc.PacPort = c.GetInt("pac", "port")
	cc.PacFile = c.GetString("pac", "file")
	cc.PacAdmode = c.GetInt("pac", "admode")
	cc.PacAdblock = c.GetString("pac", "adblock")
	cc.PacGfwlist = c.GetString("pac", "gfwlist")
	cc.PacExpired = c.GetInt("pac", "expired")

	cc.PhpEnable = c.GetBool("php", "enable")
	cc.PhpListen = c.GetString("php", "listen")
	cc.PhpPassword = c.GetString("php", "password")
	cc.PhpCrlf = c.GetBool("php", "crlf")
	cc.PhpValidate = c.GetBool("php", "validate")
	cc.PhpKeepalive = c.GetBool("php", "keepalive")
	cc.PhpFetchserver = c.GetStrings("php", "fetchserver")
	cc.PhpHosts = c.GetStrings("php", "hosts")

	cc.VpsEnable = c.GetBool("vps", "enable")
	cc.VpsListen = c.GetString("vps", "listen")
	cc.VpsFetchserver = c.GetStrings("vps", "fetchserver")

	cc.ProxyEnable = c.GetBool("proxy", "enable")
	cc.ProxyAutodetect = c.GetBool("proxy", "autodetect")
	cc.ProxyHost = c.GetString("proxy", "host")
	cc.ProxyPort = c.GetInt("proxy", "port")
	cc.ProxyUsername = c.GetString("proxy", "username")
	cc.ProxyPasswrod = c.GetString("proxy", "password")

	cc.AutorangeHosts = c.GetStrings("autorange", "hosts")
	cc.AutorangeEndswith = c.GetStrings("autorange", "endswith")
	cc.AutorangeNoendswith = c.GetStrings("autorange", "noendswith")
	cc.AutorangeThreads = c.GetInt("autorange", "threads")
	cc.AutorangeMaxsize = c.GetInt("autorange", "maxsize")
	cc.AutorangeWaitsize = c.GetInt("autorange", "waitsize")
	cc.AutorangeBufsize = c.GetInt("autorange", "bufsize")

	cc.DnsEnable = c.GetBool("dns", "enable")
	cc.DnsListen = c.GetString("dns", "listen")
	cc.DnsServers = c.GetStrings("dns", "servers")

	cc.UseragentEnable = c.GetBool("useragent", "enable")
	cc.UseragentString = c.GetString("useragent", "string")

	cc.FetchmaxLocal = c.GetInt("fetchmax", "local")
	cc.FetchmaxServer = c.GetInt("fetchmax", "server")

	cc.LoveEnable = c.GetBool("love", "enable")
	cc.LoveTip = c.GetStrings("love", "tip")

	cc.IplistMap = make(map[string][]string)
	cc.IplistFixed = make([]string, 0)
	cc.HostMap = make(map[string]string, 0)
	cc.UrlRewriteMap = make(map[string]string, 0)
	cc.CrlfSites = make([]string, 0)
	cc.NocrlfSites = make([]string, 0)
	cc.ForcehttpsSites = make([]string, 0)
	cc.NoforcehttpsSites = make(map[string]struct{}, 0)
	cc.FakehttpsSites = make([]string, 0)
	cc.NofakehttpsSites = make(map[string]struct{}, 0)
	cc.WithGAESites = make([]string, 0)
	cc.WithPHPSites = make([]string, 0)
	cc.WithVPSSites = make([]string, 0)
	for _, option := range c.GetOptions("iplist") {
		cc.IplistMap[option] = c.GetStrings("iplist", option)
	}
	for _, option := range c.GetOptions("profile") {
		pattern := option
		rules := c.GetStrings("profile", option)
		for {
			if len(rules) == 0 {
				break
			}
			rule := rules[0]
			rules = rules[1:]
			switch rule {
			case "crlf":
				cc.CrlfSites = append(cc.CrlfSites, pattern)
			case "nocrlf":
				cc.NocrlfSites = append(cc.NocrlfSites, pattern)
			case "forcehttps":
				cc.ForcehttpsSites = append(cc.ForcehttpsSites, pattern)
			case "noforcehttps":
				cc.NoforcehttpsSites[pattern] = struct{}{}
			case "fakehttps":
				cc.FakehttpsSites = append(cc.FakehttpsSites, pattern)
			case "nofakehttps":
				cc.NofakehttpsSites[pattern] = struct{}{}
			case "withgae":
				cc.WithGAESites = append(cc.WithGAESites, pattern)
			case "withphp":
				cc.WithPHPSites = append(cc.WithPHPSites, pattern)
			case "withvps":
				cc.WithVPSSites = append(cc.WithVPSSites, pattern)
			case "direct":
				cc.HostMap[pattern] = ""
			default:
				if _, ok := cc.IplistMap[rule]; ok {
					cc.HostMap[pattern] = rule
				} else if strings.Contains(pattern, "\\") {
					cc.UrlRewriteMap[pattern] = rule
				}
			}
		}
	}

	return cc, nil
}
