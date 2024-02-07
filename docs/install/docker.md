You can evaluate the program with:

```sh
docker run -it --rm -p 3000:3000 ddvk/rmfakecloud
```

To setup it for normal ussage, you'll use need to setup a volume to store user configuration and documents:

```sh
docker run -it --rm -p 3000:3000 -v ./data:/data -e JWT_SECRET_KEY='something' ddvk/rmfakecloud
```

Explore others configuration variables on [the dedicated page](configuration.md).


## docker-compose file

```yaml
version: "3"
services:
  rmfakecloud:
    image: ddvk/rmfakecloud
    container_name: rmfakecloud
    restart: unless-stopped
    env_file:
      - env
    volumes:
      - ./data:/data
```

In this example, an external file named `env` is provided that contains the environment variables. Any of the [ways to set environment variables](https://docs.docker.com/compose/environment-variables/set-environment-variables/) for docker compose will work.

For the possible environment variables, please have a look in the [configuration](configuration.md) section.


## Rebuild the image

You can use the script `dockerbuild.sh` or there is a `make` rule:

```sh
make docker
```
