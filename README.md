#Build

tinygo build -scheduler=none -target=wasi -gc=custom -tags='custommalloc nottinygc_finalizer proxy_wasm_version_0_2_100' -o ./build/plugin.wasm main.go

#apps
applications to test the route/ingress

#audio-process
the wasm to route the audio to ASR, receives the text and then route the result to LLM to fetch the intention

#keycloak
the jwt application

#higress
higress deployment
hgctl install --set profile=local-k8s --set console.o11yEnabled=true