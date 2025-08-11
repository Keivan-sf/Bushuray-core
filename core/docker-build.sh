docker build -t bushuray-core-builder .
docker run --rm -v "$PWD:/out" bushuray-core-builder cp /src/bushuray-core /out
