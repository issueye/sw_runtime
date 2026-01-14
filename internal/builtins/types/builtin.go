package types

import "github.com/dop251/goja"

type BuiltinModule interface {
	GetModule() *goja.Object
}

type NamespaceModule interface {
	GetModule() *goja.Object
	GetSubModule(name string) (BuiltinModule, bool)
}
