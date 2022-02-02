export DOCKER_BUILDKIT=1
docker build -t rmfakecloud --no-cache --build-arg VERSION="$(git describe --tags)" .
