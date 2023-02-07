#!/bin/bash
bash BUILD/winbuild.sh
bash BUILD/linuxbuild.sh
bash BUILD/wasm.sh

bash BUILD/winbuild-bench.sh
bash BUILD/linuxbuild-bench.sh
bash BUILD/wasm-bench.sh