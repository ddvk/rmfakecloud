You can evaluate the program with:

```sh
docker run -it --rm -p 3000:3000 ddvk/rmfakecloud
```

To setup it for its exploitation, you'll use need to setup a volume (it will contain user configuration and synchronized documents):

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

A `env` file is needed where all of the environmental variables are defined.
Using the `environment:` option in the compose file is also valid and everything is in one file.

For the possible environment variables, please have a look in the [configuration](configuration.md) section.


## Rebuild the image

You can use the script `dockerbuild.sh` or there is a `make` rule:

```sh
make docker
```
