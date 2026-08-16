package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const systemdTemplate = `[Unit]
Description=Sing Box Panel
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s run --listen %s --data-dir %s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`

const openrcTemplate = `#!/sbin/openrc-run

name="sing-panel"
description="Sing Box Panel"

command="%s"
command_args="run --listen %s --data-dir %s"
command_background=true
pidfile="/run/sing-panel.pid"
directory="%s"

output_log="/var/log/sing-panel.log"
error_log="/var/log/sing-panel.log"

depend() {
    need net
    after firewall
}
`

func detectInitSystem() string {
	// Check for systemd
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return "systemd"
	}
	// Check for OpenRC (Alpine)
	if _, err := os.Stat("/sbin/openrc-run"); err == nil {
		return "openrc"
	}
	return ""
}

func getBinaryPath() (string, error) {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", err
	}
	// Resolve symlinks
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func cmdInstall(listen, dataDir string) {
	initSystem := detectInitSystem()
	if initSystem == "" {
		fmt.Println("错误: 未检测到支持的初始化系统 (systemd/openrc)")
		fmt.Println("仅支持 Debian (systemd) 和 Alpine (openrc)")
		os.Exit(1)
	}

	binPath, err := getBinaryPath()
	if err != nil {
		fmt.Printf("错误: 获取二进制路径失败: %v\n", err)
		os.Exit(1)
	}

	switch initSystem {
	case "systemd":
		installSystemd(binPath, dataDir, listen)
	case "openrc":
		installOpenRC(binPath, dataDir, listen)
	}
}

func installSystemd(binPath, dataDir, listen string) {
	serviceFile := fmt.Sprintf(systemdTemplate, dataDir, binPath, listen, dataDir)
	servicePath := "/etc/systemd/system/sing-panel.service"

	fmt.Println("检测到 systemd，正在安装服务...")

	if err := os.WriteFile(servicePath, []byte(serviceFile), 0644); err != nil {
		fmt.Printf("错误: 写入服务文件失败: %v\n", err)
		os.Exit(1)
	}

	commands := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "sing-panel"},
		{"systemctl", "start", "sing-panel"},
	}

	for _, cmd := range commands {
		fmt.Printf("执行: %s\n", strings.Join(cmd, " "))
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			fmt.Printf("错误: %s\n%s\n", strings.Join(cmd, " "), string(out))
			os.Exit(1)
		}
	}

	fmt.Println("安装完成!")
	fmt.Println("  查看状态: systemctl status sing-panel")
	fmt.Println("  查看日志: journalctl -u sing-panel -f")
	fmt.Println("  停止服务: systemctl stop sing-panel")
	fmt.Println("  卸载服务: sing-panel uninstall")
}

func installOpenRC(binPath, dataDir, listen string) {
	serviceFile := fmt.Sprintf(openrcTemplate, binPath, listen, dataDir, dataDir)
	servicePath := "/etc/init.d/sing-panel"

	fmt.Println("检测到 OpenRC，正在安装服务...")

	if err := os.WriteFile(servicePath, []byte(serviceFile), 0755); err != nil {
		fmt.Printf("错误: 写入服务文件失败: %v\n", err)
		os.Exit(1)
	}

	commands := [][]string{
		{"rc-update", "add", "sing-panel", "default"},
		{"rc-service", "sing-panel", "start"},
	}

	for _, cmd := range commands {
		fmt.Printf("执行: %s\n", strings.Join(cmd, " "))
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			fmt.Printf("警告: %s\n%s\n", strings.Join(cmd, " "), string(out))
		}
	}

	fmt.Println("安装完成!")
	fmt.Println("  查看状态: rc-service sing-panel status")
	fmt.Println("  查看日志: tail -f /var/log/sing-panel.log")
	fmt.Println("  停止服务: rc-service sing-panel stop")
	fmt.Println("  卸载服务: sing-panel uninstall")
}

func cmdUninstall() {
	initSystem := detectInitSystem()
	if initSystem == "" {
		fmt.Println("错误: 未检测到支持的初始化系统")
		os.Exit(1)
	}

	switch initSystem {
	case "systemd":
		fmt.Println("正在卸载 systemd 服务...")
		commands := [][]string{
			{"systemctl", "stop", "sing-panel"},
			{"systemctl", "disable", "sing-panel"},
		}
		for _, cmd := range commands {
			exec.Command(cmd[0], cmd[1:]...).Run()
		}
		os.Remove("/etc/systemd/system/sing-panel.service")
		exec.Command("systemctl", "daemon-reload").Run()
		fmt.Println("systemd 服务已卸载")

	case "openrc":
		fmt.Println("正在卸载 OpenRC 服务...")
		exec.Command("rc-service", "sing-panel", "stop").Run()
		exec.Command("rc-update", "delete", "sing-panel", "default").Run()
		os.Remove("/etc/init.d/sing-panel")
		fmt.Println("OpenRC 服务已卸载")
	}
}
