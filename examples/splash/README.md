# Splash example

This repository contains a simple example of how to work with
[Splash](https://splash.readthedocs.io) and crawlera-headless-proxy.

To run this example, please follow [official
instructions](https://pipenv.readthedocs.io/en/latest/install/) on how
to install pipenv. After that, please prepare the environment with the
following command:

```console
$ pipenv sync
```

This will install docker-compose. After that, please run the following
command:

```console
$ pipenv run docker-compose up
```

This will start up both headless-proxy and Splash instance.

This example has 2 different scripts:

1. `simple-endpoint.sh` is an example of how to use `/render.html` endpoint
   with headless proxy
2. `lua-endpoint.sh` is an example of how to use `/execute` endpoint with
   custom Lua script (see `example.lua` file).
