package builtins

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// RaftModule Raft 模块
type RaftModule struct {
	vm      *goja.Runtime
	manager *Manager
}

// NewRaftModule 创建 Raft 模块
func NewRaftModule(vm *goja.Runtime, manager *Manager) *RaftModule {
	return &RaftModule{
		vm:      vm,
		manager: manager,
	}
}

// GetModule 获取模块对象 interface
func (m *RaftModule) GetModule() *goja.Object {
	obj := m.vm.NewObject()
	obj.Set("createNode", m.createNode)
	return obj
}

// createNode 创建 Raft 节点
func (m *RaftModule) createNode(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(m.vm.NewTypeError("createNode requires configuration object"))
	}

	configObj := call.Arguments[0].ToObject(m.vm)
	nodeID := configObj.Get("nodeID").String()
	addr := configObj.Get("advertiseAddr").String()
	dataDir := configObj.Get("dataDir").String()

	// 获取 JS 实现的 FSM
	jsFSMObj := configObj.Get("fsm")
	if jsFSMObj == nil || goja.IsNull(jsFSMObj) || goja.IsUndefined(jsFSMObj) {
		panic(m.vm.NewTypeError("fsm implementation is required"))
	}
	jsFSM := jsFSMObj.ToObject(m.vm)

	// 准备配置
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)
	config.SnapshotInterval = 20 * time.Second
	config.SnapshotThreshold = 1024

	// 创建地址
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 传输层
	transport, err := raft.NewTCPTransport(addr, tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 确保存储目录存在
	os.MkdirAll(dataDir, 0755)

	// 创建快照存储
	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 2, os.Stderr)
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 创建日志存储 (BoltDB)
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.db"))
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 创建稳定存储 (BoltDB)
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.db"))
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 创建 FSM 桥接器
	fsmBridge := &JSFSMBridge{
		vm:      m.vm,
		jsFSM:   jsFSM,
		manager: m.manager,
	}

	// 创建 Raft 系统
	r, err := raft.NewRaft(config, fsmBridge, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		panic(m.vm.NewGoError(err))
	}

	// 返回 Node 对象
	nodeObj := m.vm.NewObject()

	nodeObj.Set("bootstrap", func(call goja.FunctionCall) goja.Value {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		future := r.BootstrapCluster(configuration)
		return m.vm.ToValue(future.Error() == nil)
	})

	nodeObj.Set("join", func(call goja.FunctionCall) goja.Value {
		id := call.Arguments[0].String()
		addr := call.Arguments[1].String()
		future := r.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)

		promise, resolve, reject := m.vm.NewPromise()
		go func() {
			if err := future.Error(); err != nil {
				reject(err)
			} else {
				resolve(true)
			}
		}()
		return m.vm.ToValue(promise)
	})

	nodeObj.Set("apply", func(call goja.FunctionCall) goja.Value {
		cmd := call.Arguments[0]
		var data []byte

		if goja.IsUndefined(cmd) {
			return goja.Undefined()
		}

		// 根据类型处理数据
		if str, ok := cmd.Export().(string); ok {
			data = []byte(str)
		} else {
			// 尝试 JSON 序列化
			b, err := json.Marshal(cmd.Export())
			if err != nil {
				panic(m.vm.NewGoError(err))
			}
			data = b
		}

		timeout := 10 * time.Second
		if len(call.Arguments) > 1 {
			timeout = time.Duration(call.Arguments[1].ToInteger()) * time.Millisecond
		}

		promise, resolve, reject := m.vm.NewPromise()
		go func() {
			future := r.Apply(data, timeout)
			if err := future.Error(); err != nil {
				reject(err)
			} else {
				resolve(future.Response())
			}
		}()
		return m.vm.ToValue(promise)
	})

	nodeObj.Set("stats", func(call goja.FunctionCall) goja.Value {
		return m.vm.ToValue(r.Stats())
	})

	nodeObj.Set("shutdown", func(call goja.FunctionCall) goja.Value {
		future := r.Shutdown()
		return m.vm.ToValue(future.Error() == nil)
	})

	return nodeObj
}

// JSFSMBridge JavaScript FSM 桥接器
type JSFSMBridge struct {
	vm      *goja.Runtime
	jsFSM   *goja.Object
	manager *Manager
}

// Apply 应用日志
func (f *JSFSMBridge) Apply(l *raft.Log) interface{} {
	// 必须在 EventLoop 中同步执行 JS 代码
	// 注意：这里需要通过 Manager 访问到 EventLoop，或者我们假设 Manager 有相关能力
	// 我们在 manager.go 中没有直接暴露 EventLoop，但 Runner 中有
	// 临时方案：我们扩展 Manager 接口或者假设 manager 已经增强了

	// 在目前的架构中，manager 持有 vm，但没有持有 loop 的引用。
	// 我们需要一个回调机制。在 process 模块中我们添加了 SetNextTick。
	// 类似地，我们可以添加一个 RunScriptSync 钩子。

	// 为了演示，我们假设我们可以在 vm 上直接运行，因为 Apply 是串行的。
	// 但这在多线程环境下是不安全的（goja.Runtime 不是线程安全的）。
	// 必须使用 Runner.loop.RunOnLoopSync。

	// 修复方案：在 Manager 中添加 Loop 引用或者回调。
	// 这里我们假设 Manager 已经有了 RunOnLoopSync (稍后会添加)

	return f.manager.RunOnLoopSync(func(vm *goja.Runtime) interface{} {
		applyMethod, ok := goja.AssertFunction(f.jsFSM.Get("apply"))
		if !ok {
			return nil
		}

		// 传递日志数据
		// Convert byte slice to string/buffer
		res, err := applyMethod(f.jsFSM, vm.ToValue(string(l.Data)))
		if err != nil {
			fmt.Println("FSM Apply Error:", err)
			return nil
		}
		return res.Export()
	})
}

// Snapshot 创建快照
func (f *JSFSMBridge) Snapshot() (raft.FSMSnapshot, error) {
	// 同步调用 JS 获取快照数据
	res := f.manager.RunOnLoopSync(func(vm *goja.Runtime) interface{} {
		snapMethod, ok := goja.AssertFunction(f.jsFSM.Get("snapshot"))
		if !ok {
			return nil
		}

		val, err := snapMethod(f.jsFSM)
		if err != nil {
			return err
		}
		return val.Export()
	})

	if err, ok := res.(error); ok {
		return nil, err
	}

	return &JSSnapshot{data: res}, nil
}

// Restore 恢复快照
func (f *JSFSMBridge) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	res := f.manager.RunOnLoopSync(func(vm *goja.Runtime) interface{} {
		restoreMethod, ok := goja.AssertFunction(f.jsFSM.Get("restore"))
		if !ok {
			return nil
		}

		_, err := restoreMethod(f.jsFSM, vm.ToValue(string(data)))
		return err
	})

	if res != nil {
		return res.(error)
	}
	return nil
}

// JSSnapshot 快照包装器
type JSSnapshot struct {
	data interface{}
}

func (s *JSSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// 序列化数据
		b, err := json.Marshal(s.data)
		if err != nil {
			return err
		}
		if _, err := sink.Write(b); err != nil {
			return err
		}
		return sink.Close()
	}()
	if err != nil {
		sink.Cancel()
	}
	return err
}

func (s *JSSnapshot) Release() {}
