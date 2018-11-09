class RtspSession {
    constructor() {
        CreateRtspSession(this, (handle)=> {
            console.log("Creat Session Success")
            this.BindHandlerFunc((data)=>{console.log(data)}, (data)=>{console.log(data)})
            this.handle = handle
        })
    }
    
    RtpEventHandler(data) {
        console.log(data)
    }

    RtspEventHandler(event) {
        console.log(data)
    }
}