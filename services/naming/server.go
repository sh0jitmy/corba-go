package naming

import (
	"fmt"
	"strings"
	"sync"
)

type NamingContextImpl struct {
	bindings map[string]string
	mu       sync.RWMutex
}

func NewNamingContextImpl() *NamingContextImpl {
	return &NamingContextImpl{
		bindings: make(map[string]string),
	}
}

func nameToString(n CosNaming_Name) string {
	var parts []string
	for _, c := range n {
		parts = append(parts, string(c.Id))
	}
	return strings.Join(parts, "/")
}

func (n *NamingContextImpl) Bind(name CosNaming_Name, obj string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	strName := nameToString(name)
	if _, exists := n.bindings[strName]; exists {
		return &CosNaming_AlreadyBound{}
	}
	n.bindings[strName] = obj
	fmt.Printf("NamingService: Bound '%s' to '%s'\n", strName, obj)
	return nil
}

func (n *NamingContextImpl) Rebind(name CosNaming_Name, obj string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	strName := nameToString(name)
	n.bindings[strName] = obj
	fmt.Printf("NamingService: Rebound '%s' to '%s'\n", strName, obj)
	return nil
}

func (n *NamingContextImpl) Bind_context(name CosNaming_Name, nc string) error {
	return n.Bind(name, nc)
}

func (n *NamingContextImpl) Rebind_context(name CosNaming_Name, nc string) error {
	return n.Rebind(name, nc)
}

func (n *NamingContextImpl) Resolve(name CosNaming_Name) (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	strName := nameToString(name)
	ior, ok := n.bindings[strName]
	if !ok {
		return "", &CosNaming_NotFound{Why: 0, Rest_of_name: name}
	}
	fmt.Printf("NamingService: Resolved '%s'\n", strName)
	return ior, nil
}

func (n *NamingContextImpl) Unbind(name CosNaming_Name) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	strName := nameToString(name)
	if _, ok := n.bindings[strName]; !ok {
		return &CosNaming_NotFound{Why: 0, Rest_of_name: name}
	}
	delete(n.bindings, strName)
	fmt.Printf("NamingService: Unbound '%s'\n", strName)
	return nil
}

func (n *NamingContextImpl) New_context() (string, error) {
	// Dummy implementation returning empty object
	return "", nil
}

func (n *NamingContextImpl) Bind_new_context(name CosNaming_Name) (string, error) {
	// Dummy implementation returning empty object
	return "", nil
}

func (n *NamingContextImpl) Destroy() error {
	return nil
}
