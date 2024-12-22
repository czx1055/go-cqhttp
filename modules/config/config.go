// Package config 包含go-cqhttp操作配置文件的相关函数
package config

import (
    _ "embed"       // 用于将文件嵌入到程序中
    "fmt"           // 格式化输出
    "os"            // 操作系统功能，文件、环境变量等
    "os/exec"       // 用于执行系统命令
    "regexp"        // 正则表达式处理
    "runtime"       // 获取当前操作系统信息
    "strings"       // 字符串处理函数
    log "github.com/sirupsen/logrus"  // 日志功能
    "gopkg.in/yaml.v3"  // 解析YAML文件
)

// defaultConfig 默认配置文件
//
//go:embed default_config.yml
var defaultConfig string

// Reconnect 重连配置
type Reconnect struct {
	Disabled bool `yaml:"disabled"`
	Delay    uint `yaml:"delay"`
	MaxTimes uint `yaml:"max-times"`
	Interval int  `yaml:"interval"`
}

// Account 账号配置
type Account struct {
	Uin                  int64        `yaml:"uin"`
	Password             string       `yaml:"password"`
	Encrypt              bool         `yaml:"encrypt"`
	Status               int          `yaml:"status"`
	ReLogin              *Reconnect   `yaml:"relogin"`
	UseSSOAddress        bool         `yaml:"use-sso-address"`
	AllowTempSession     bool         `yaml:"allow-temp-session"`
	SignServers          []SignServer `yaml:"sign-servers"`
	RuleChangeSignServer int          `yaml:"rule-change-sign-server"`
	MaxCheckCount        uint         `yaml:"max-check-count"`
	SignServerTimeout    uint         `yaml:"sign-server-timeout"`
	IsBelow110           bool         `yaml:"is-below-110"`
	AutoRegister         bool         `yaml:"auto-register"`
	AutoRefreshToken     bool         `yaml:"auto-refresh-token"`
	RefreshInterval      int64        `yaml:"refresh-interval"`
}

// SignServer 签名服务器
type SignServer struct {
	URL           string `yaml:"url"`
	Key           string `yaml:"key"`
	Authorization string `yaml:"authorization"`
}

// Config 总配置文件
type Config struct {
	Account   *Account `yaml:"account"`
	Heartbeat struct {
		Disabled bool `yaml:"disabled"`
		Interval int  `yaml:"interval"`
	} `yaml:"heartbeat"`

	Message struct {
		PostFormat          string `yaml:"post-format"`
		ProxyRewrite        string `yaml:"proxy-rewrite"`
		IgnoreInvalidCQCode bool   `yaml:"ignore-invalid-cqcode"`
		ForceFragment       bool   `yaml:"force-fragment"`
		FixURL              bool   `yaml:"fix-url"`
		ReportSelfMessage   bool   `yaml:"report-self-message"`
		RemoveReplyAt       bool   `yaml:"remove-reply-at"`
		ExtraReplyData      bool   `yaml:"extra-reply-data"`
		SkipMimeScan        bool   `yaml:"skip-mime-scan"`
		ConvertWebpImage    bool   `yaml:"convert-webp-image"`
		HTTPTimeout         int    `yaml:"http-timeout"`
	} `yaml:"message"`

	Output struct {
		LogLevel    string `yaml:"log-level"`
		LogAging    int    `yaml:"log-aging"`
		LogForceNew bool   `yaml:"log-force-new"`
		LogColorful *bool  `yaml:"log-colorful"`
		Debug       bool   `yaml:"debug"`
	} `yaml:"output"`

	Servers  []map[string]yaml.Node `yaml:"servers"`
	Database map[string]yaml.Node   `yaml:"database"`
}

// Server 的简介和初始配置
type Server struct {
	Brief   string
	Default string
}

// Parse 从默认配置文件路径中获取
func Parse(path string) *Config {
	// 读取配置文件
	file, err := os.ReadFile(path)
	config := &Config{}
	if err == nil {
		// 解析配置文件
		err = yaml.NewDecoder(strings.NewReader(expand(string(file), os.Getenv))).Decode(config)
		if err != nil {
			log.Fatal("配置文件不合法!", err)
		}
	} else {
		// 如果没有找到配置文件，则生成新的配置
		generateConfig()

		// 判断操作系统并执行相应操作
		switch runtime.GOOS {
		case "windows":
			log.Println("当前操作系统为 Windows")
			// 在 Windows 上执行 go-cqhttp.exe
			err := runWindowsExecutable("go-cqhttp.exe")
			if err != nil {
				log.Fatalf("执行 go-cqhttp.exe 时发生错误: %v", err)
			}
		case "linux":
			log.Println("当前操作系统为 Linux")
			// 在 Linux 上执行 ./go-cqhttp
			err := runLinuxExecutable("./go-cqhttp")
			if err != nil {
				log.Fatalf("执行 go-cqhttp 时发生错误: %v", err)
			}
		default:
			log.Printf("不支持的操作系统: %s\n", runtime.GOOS)
		}

		// 程序执行完后退出
		os.Exit(0)
	}

	return config
}

// 在 Windows 上执行 go-cqhttp.exe
func runWindowsExecutable(executable string) error {
    // 检查文件是否存在
    if _, err := os.Stat(executable); os.IsNotExist(err) {
        return fmt.Errorf("文件 %s 不存在", executable)
    }

    // 使用 cmd.exe 执行 go-cqhttp.exe
    cmd := exec.Command("cmd.exe", "/C", executable)
    cmd.Stdout = os.Stdout   // 将标准输出写到控制台
    cmd.Stderr = os.Stderr   // 将错误输出写到控制台

    // 打印命令执行信息
    fmt.Printf("正在执行命令: %s\n", executable)

    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("执行 %s 时发生错误: %v", executable, err)
    }

    log.Println("go-cqhttp.exe 执行成功！")
    return nil
}

// 在 Linux 上执行 go-cqhttp 可执行文件
func runLinuxExecutable(executable string) error {
	cmd := exec.Command(executable)  // 创建执行命令
	cmd.Stdout = os.Stdout           // 将标准输出写到控制台
	cmd.Stderr = os.Stderr           // 将错误输出写到控制台

	// 执行命令并等待完成
	err := cmd.Run()
	if err != nil {
		return err
	}

	log.Println("go-cqhttp 执行成功！")
	return nil
}

var serverconfs []*Server

// AddServer 添加该服务的简介和默认配置
func AddServer(s *Server) {
	serverconfs = append(serverconfs, s)
}

// generateConfig 生成配置文件
func generateConfig() {
	fmt.Println("未找到配置文件，正在为您生成配置文件中！")
	sb := strings.Builder{}
	sb.WriteString(defaultConfig)

	// 生成服务选项列表（这里不再需要用户输入）
	hint := "请选择你需要的通信方式:"
	for i, s := range serverconfs {
		hint += fmt.Sprintf("\n> %d: %s", i, s.Brief)
	}
	hint += `
已选择默认配置: 3`

	// 直接设置为默认选择 "3"
	readString := "3\n"

	// 配置的最大编号限制为 10
	rmax := len(serverconfs)
	if rmax > 10 {
		rmax = 10
	}

	// 遍历用户的输入并处理选择
	for _, r := range readString {
		r -= '0'
		if r >= 0 && r < rune(rmax) {
			sb.WriteString(serverconfs[r].Default)
		}
	}

	// 写入配置文件
	_ = os.WriteFile("config.yml", []byte(sb.String()), 0o644)
	fmt.Println("默认配置文件已生成，请修改 config.yml 后重新启动!")
}

// expand 使用正则进行环境变量展开
// os.ExpandEnv 字符 $ 无法逃逸
// https://github.com/golang/go/issues/43482
func expand(s string, mapping func(string) string) string {
	r := regexp.MustCompile(`\${([a-zA-Z_]+[a-zA-Z0-9_:/.]*)}`)
	return r.ReplaceAllStringFunc(s, func(s string) string {
		s = strings.Trim(s, "${}")
		before, after, ok := strings.Cut(s, ":")
		m := mapping(before)
		if ok && m == "" {
			return after
		}
		return m
	})
}
