package jsrtspclient

import (
	"syscall/js"
)

type JsRtspManager struct {
	jsRtspSessionMap     map[int]*JsRtspClientSession
	nextHandle           int
	done                 chan struct{}
	createSessionCb      js.Callback
	releaseRtspSessionCb js.Callback
	shutdownCb           js.Callback
}

func NewRtspManager() *JsRtspManager {
	return &JsRtspManager{
		jsRtspSessionMap: make(map[int]*JsRtspClientSession),
		nextHandle:       0,
		done:             make(chan struct{}),
	}
}

func (manager *JsRtspManager) findSession(handle int) *JsRtspClientSession {
	jsSession, ok := manager.jsRtspSessionMap[handle]
	if false == ok {
		return nil
	}
	return jsSession
}

func (manager *JsRtspManager) deleteAllSession() {
	for _, jsSession := range manager.jsRtspSessionMap {
		jsSession.ReleaseSession()
	}
}

func (manager *JsRtspManager) bindGlobalFunc() {
	manager.createSessionCb = js.NewCallback(func(args []js.Value) {
		handle := manager.CreateRtspSession(args[0])
		callback := args[1]
		callback.Invoke(handle)
	})
	js.Global().Set("CreateRtspSession", manager.createSessionCb)

	manager.releaseRtspSessionCb = js.NewCallback(func(args []js.Value) {
		manager.ReleaseRtspSession(args[0].Int())
		if len(args) > 1 {
			callback := args[1]
			callback.Invoke()
		}
	})
	js.Global().Set("ReleaseRtspSession", manager.releaseRtspSessionCb)

	manager.shutdownCb = js.NewCallback(func(args []js.Value) {
		manager.done <- struct{}{}
	})
	js.Global().Set("Shutdown", manager.shutdownCb)
}

func (manager *JsRtspManager) releaseGobalFunc() {
	manager.createSessionCb.Release()
	manager.releaseRtspSessionCb.Release()
	manager.shutdownCb.Release()
}

func (manager *JsRtspManager) Start() {
	manager.bindGlobalFunc()
	<-manager.done
	manager.deleteAllSession()
	manager.releaseGobalFunc()
}

func (manager *JsRtspManager) CreateRtspSession(jsSessionObj js.Value) int {
	jsSession := NewJsRtspClientSession(jsSessionObj)
	manager.jsRtspSessionMap[manager.nextHandle] = jsSession
	manager.nextHandle++
	return manager.nextHandle - 1
}

func (manager *JsRtspManager) ReleaseRtspSession(handle int) {
	jsSession := manager.findSession(handle)
	if nil == jsSession {
		return
	}
	jsSession.ReleaseSession()
	delete(manager.jsRtspSessionMap, handle)
}
