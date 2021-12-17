## Docker run command
`docker run -it --rm -p 3000:3000 -e JWT_SECRET_KEY='something' ddvk/rmfakecloud`

(you can pass `-h` to see the available options)

## Docker compose file
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
For the possible environment variables, please have a look in the [environment](environment.md) section.
