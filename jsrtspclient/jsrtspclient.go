package jsrtspclient

import (
	"encoding/base64"
	"rtspclient/rtspclient"
	"syscall/js"
)

type JsRtspClientSession struct {
	Session            *rtspclient.RtspClientSession
	jsHandlerObj       js.Value // rtsp callback obj
	jsRtpHandler       js.Value
	jsRtspEventHandler js.Value
	jsSessionObj       js.Value    // js rtsp session, bind rtspsession interface
	playCb             js.Callback // js call go function ref
	closeCb            js.Callback
	bindHandlerObjCb   js.Callback
	bindHandlerFuncCb  js.Callback
}

func NewJsRtspClientSession(jsSessionObj js.Value) *JsRtspClientSession {
	jsSession := &JsRtspClientSession{}
	jsSession.Session = rtspclient.NewRtspClientSession(jsSession.RtpEventHandler, jsSession.RtspEventHandler)
	jsSession.jsSessionObj = jsSessionObj
	jsSession.bindJsSession()
	return jsSession
}

func (jsSession *JsRtspClientSession) bindJsSession() {
	jsSession.playCb = js.NewCallback(func(args []js.Value) {
		URL := args[0]
		err := jsSession.Session.Play(URL.String())
		if len(args) > 1 {
			callback := args[1]
			callback.Invoke(nil == err)
		}
	})
	jsSession.jsSessionObj.Set("Play", jsSession.playCb)

	jsSession.closeCb = js.NewCallback(func(args []js.Value) {
		jsSession.Session.Close()
		if len(args) > 0 {
			callback := args[0]
			callback.Invoke(true)
		}
	})
	jsSession.jsSessionObj.Set("Close", jsSession.closeCb)

	jsSession.bindHandlerObjCb = js.NewCallback(func(args []js.Value) {
		jsSession.BindHandlerObj(args[0])
		if len(args) > 1 {
			callback := args[1]
			callback.Invoke()
		}
	})
	jsSession.jsSessionObj.Set("BindHandlerObj", jsSession.bindHandlerObjCb)

	jsSession.bindHandlerFuncCb = js.NewCallback(func(args []js.Value) {
		jsSession.BindHandlerFunc(args[0], args[1])
		if len(args) > 2 {
			callback := args[2]
			callback.Invoke()
		}
	})
	jsSession.jsSessionObj.Set("BindHandlerFunc", jsSession.bindHandlerFuncCb)
}

func (jsSession *JsRtspClientSession) ReleaseSession() {
	jsSession.playCb.Release()
	jsSession.closeCb.Release()
	jsSession.bindHandlerObjCb.Release()
	jsSession.bindHandlerFuncCb.Release()
}

func (jsSession *JsRtspClientSession) BindHandlerObj(jsHandlerObj js.Value) {
	jsSession.jsHandlerObj = jsHandlerObj
}

func (jsSession *JsRtspClientSession) BindHandlerFunc(jsRtpHandler js.Value, jsRtspEventHandler js.Value) {
	jsSession.jsRtpHandler = jsRtpHandler
	jsSession.jsRtspEventHandler = jsRtspEventHandler
}

func (jsSession *JsRtspClientSession) RtpEventHandler(data *rtspclient.RtspData) {
	args := make(map[string]interface{})
	args["ChannelNum"] = data.ChannelNum
	args["Data"] = base64.StdEncoding.EncodeToString(data.Data)
	if jsSession.jsHandlerObj.Type() == js.TypeObject {
		jsSession.jsHandlerObj.Call("RtpEventHandler", args)
	} else if jsSession.jsRtpHandler.Type() == js.TypeFunction {
		jsSession.jsRtpHandler.Invoke(args)
	} else {
		jsSession.jsSessionObj.Call("RtpEventHandler", args)
	}
}

func (jsSession *JsRtspClientSession) RtspEventHandler(event *rtspclient.RtspEvent) {
	args := make(map[string]interface{})
	args["EventType"] = event.EventType
	args["Data"] = base64.StdEncoding.EncodeToString(event.Data)
	if jsSession.jsHandlerObj.Type() == js.TypeObject {
		jsSession.jsHandlerObj.Call("RtspEventHandler", args)
	} else if jsSession.jsRtspEventHandler.Type() == js.TypeFunction {
		jsSession.jsRtspEventHandler.Invoke(args)
	} else {
		jsSession.jsSessionObj.Call("RtspEventHandler", args)
	}
}
