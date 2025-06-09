Build

tinygo build -scheduler=none -target=wasi -gc=custom -tags='custommalloc nottinygc_finalizer proxy_wasm_version_0_2_100' -o ./build/plugin.wasm main.go
