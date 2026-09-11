#!/bin/sh
set -eu

image=${RPF_KERNEL_IMAGE:-rpf-kernel-lab:local}
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)

docker build -t "$image" -f "$root/lab/kernel/Dockerfile" "$root"
docker run --rm --privileged -v "$root:/src" -w /src "$image" ./tools/build-sensor.sh
docker run --rm -v "$root:/src" -w /src "$image" go test -race ./...
docker run --rm -v "$root:/src" -w /src "$image" go vet ./...
docker run --rm -v "$root:/src" -w /src "$image" go build -o build/out/rpf-sensor ./cmd/rpf-sensor
docker run --rm -v "$root:/src" -w /src "$image" go build -o build/out/rpf ./cmd/rpf
docker run --rm -v "$root:/src" -w /src "$image" go build -o build/out/rpf-local-connect ./cmd/rpf-local-connect
docker run --rm -v "$root:/src" -w /src "$image" go build -o build/out/rpf-mock-server ./cmd/rpf-mock-server
docker run --rm --privileged -v "$root:/src" -w /src "$image" ./tools/test-sensor.sh
docker run --rm --privileged -v "$root:/src" -w /src "$image" ./tools/test-adversarial.sh
