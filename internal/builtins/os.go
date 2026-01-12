package builtins

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"

	"github.com/dop251/goja"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// OSModule 操作系统模块
type OSModule struct {
	vm *goja.Runtime
}

// NewOSModule 创建操作系统模块
func NewOSModule(vm *goja.Runtime) *OSModule {
	return &OSModule{vm: vm}
}

// GetModule 获取操作系统模块对象
func (o *OSModule) GetModule() *goja.Object {
	obj := o.vm.NewObject()

	obj.Set("hostname", o.hostname)
	obj.Set("homedir", o.homedir)
	obj.Set("tmpdir", o.tmpdir)
	obj.Set("arch", o.arch)
	obj.Set("platform", o.platform)
	obj.Set("uptime", o.uptime)
	obj.Set("totalmem", o.totalmem)
	obj.Set("freemem", o.freemem)
	obj.Set("cpus", o.cpus)
	obj.Set("networkInterfaces", o.networkInterfaces)
	obj.Set("userInfo", o.userInfo)
	obj.Set("type", o.osType)
	obj.Set("release", o.release)

	return obj
}

// hostname 获取主机名
func (o *OSModule) hostname(call goja.FunctionCall) goja.Value {
	name, err := os.Hostname()
	if err != nil {
		return o.vm.ToValue("")
	}
	return o.vm.ToValue(name)
}

// homedir 获取用户主目录
func (o *OSModule) homedir(call goja.FunctionCall) goja.Value {
	dir, err := os.UserHomeDir()
	if err != nil {
		return o.vm.ToValue("")
	}
	return o.vm.ToValue(dir)
}

// tmpdir 获取临时目录
func (o *OSModule) tmpdir(call goja.FunctionCall) goja.Value {
	return o.vm.ToValue(os.TempDir())
}

// arch 获取硬件架构
func (o *OSModule) arch(call goja.FunctionCall) goja.Value {
	// 将 Go 的 GOARCH 转换为 Node.js 风格
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "ia32"
	}
	return o.vm.ToValue(arch)
}

// platform 获取操作系统平台
func (o *OSModule) platform(call goja.FunctionCall) goja.Value {
	// 将 Go 的 GOOS 转换为 Node.js 风格
	platform := runtime.GOOS
	switch platform {
	case "windows":
		platform = "win32"
	}
	return o.vm.ToValue(platform)
}

// uptime 系统运行时间
func (o *OSModule) uptime(call goja.FunctionCall) goja.Value {
	u, err := host.Uptime()
	if err != nil {
		return o.vm.ToValue(0)
	}
	return o.vm.ToValue(u)
}

// totalmem 总内存
func (o *OSModule) totalmem(call goja.FunctionCall) goja.Value {
	v, err := mem.VirtualMemory()
	if err != nil {
		return o.vm.ToValue(0)
	}
	return o.vm.ToValue(v.Total)
}

// freemem 空闲内存
func (o *OSModule) freemem(call goja.FunctionCall) goja.Value {
	v, err := mem.VirtualMemory()
	if err != nil {
		return o.vm.ToValue(0)
	}
	return o.vm.ToValue(v.Free)
}

// cpus CPU 信息
func (o *OSModule) cpus(call goja.FunctionCall) goja.Value {
	infos, err := cpu.Info()
	if err != nil {
		return o.vm.NewArray()
	}

	result := o.vm.NewArray()
	for i, info := range infos {
		item := o.vm.NewObject()
		item.Set("model", info.ModelName)
		item.Set("speed", int(info.Mhz))
		times, _ := cpu.Times(false)
		if len(times) > 0 {
			tObj := o.vm.NewObject()
			tObj.Set("user", times[0].User)
			tObj.Set("nice", times[0].Nice)
			tObj.Set("sys", times[0].System)
			tObj.Set("idle", times[0].Idle)
			tObj.Set("irq", times[0].Irq)
			item.Set("times", tObj)
		}
		result.Set(fmt.Sprintf("%d", i), item)
	}
	return result
}

// networkInterfaces 网络接口信息
func (o *OSModule) networkInterfaces(call goja.FunctionCall) goja.Value {
	interfaces, err := net.Interfaces()
	if err != nil {
		return o.vm.NewObject()
	}

	result := o.vm.NewObject()
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		ifaceArr := o.vm.NewArray()
		for i, addr := range addrs {
			item := o.vm.NewObject()
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			item.Set("address", ipnet.IP.String())
			item.Set("netmask", net.IP(ipnet.Mask).String())
			if ipnet.IP.To4() != nil {
				item.Set("family", "IPv4")
			} else {
				item.Set("family", "IPv6")
			}
			item.Set("mac", iface.HardwareAddr.String())
			item.Set("internal", (iface.Flags&net.FlagLoopback) != 0)

			ifaceArr.Set(fmt.Sprintf("%d", i), item)
		}
		result.Set(iface.Name, ifaceArr)
	}
	return result
}

// userInfo 用户信息
func (o *OSModule) userInfo(call goja.FunctionCall) goja.Value {
	u, err := user.Current()
	if err != nil {
		return goja.Undefined()
	}

	result := o.vm.NewObject()
	result.Set("username", u.Username)
	result.Set("uid", u.Uid)
	result.Set("gid", u.Gid)
	result.Set("shell", "") // Go 标准库不支持获取 shell
	result.Set("homedir", u.HomeDir)

	return result
}

// osType 操作系统类型
func (o *OSModule) osType(call goja.FunctionCall) goja.Value {
	t := runtime.GOOS
	switch t {
	case "windows":
		t = "Windows_NT"
	case "linux":
		t = "Linux"
	case "darwin":
		t = "Darwin"
	}
	return o.vm.ToValue(t)
}

// release 发布版本
func (o *OSModule) release(call goja.FunctionCall) goja.Value {
	info, err := host.Info()
	if err != nil {
		return o.vm.ToValue("")
	}
	return o.vm.ToValue(info.KernelVersion)
}
