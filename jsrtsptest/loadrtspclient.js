if (!WebAssembly.instantiateStreaming) { // polyfill
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();
        return await WebAssembly.instantiate(source, importObject);
    };
}

const go = new Go();

let mod, inst, memoryBytes;

try {    
    WebAssembly.instantiateStreaming(fetch("rtsplib.wasm", {cache: 'no-cache'}), go.importObject).then((result) => {
        mod = result.module;
        inst = result.instance;
        runGo()
    });   
} catch (error) {
    
}

async function runGo() {
    console.log("load over")
    await go.run(inst);
}