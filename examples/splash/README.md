# Splash example

This repository contains a simple example of how to work with
[Splash](https://splash.readthedocs.io) and zyte-smartproxy-headless-proxy.

To run this example, please follow [official
instructions](https://poetry.eustace.io/docs/#installation) on how to
install Poetry. After that, please prepare the environment with the
following command:

```console
$ python3 "$(command -v poetry)" install
```

This will install docker-compose. After that, please run the following
command:

```console
$ python3 "$(command -v poetry)" run docker-compose up
```

This will start up both headless-proxy and Splash instance.

This example has 2 different scripts:

1. `simple-endpoint.sh` is an example of how to use `/render.html` endpoint
   with headless proxy
2. `lua-endpoint.sh` is an example of how to use `/execute` endpoint with
   custom Lua script (see `example.lua` file).
