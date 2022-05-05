# how to build

1. install tinygo
```shell
export TINYGOROOT=/usr/local/tinygo
export PATH=$PATH:$TINYGOROOT/bin
```

2. build wasm
```shell
tinygo build -scheduler=none -target=wasi -o httpfilter.wasi .

sha256sum httpfilter.wasi
```