# Pyppeteer example

This repository contains a simple example of how to work with headless
Chrome using [Pypeteer](https://miyakogi.github.io/pyppeteer/),
unofficial Python port of [Puppeteer](https://pptr.dev/).

To run this example, please follow [official
instructions](https://poetry.eustace.io/docs/#installation) on how to
install Poetry. After that, please prepare the environment with the
following command:

```console
$ python3 "$(command -v poetry)" install
```

This will install docker-compose and pyppeteer. After that, please
install Chrome with following command:

```console
$ python3 "$(command -v poetry)" run pyppeteer-install
```

Now let's run headless proxy:

```console
$ python3 "$(command -v poetry)" run docker-compose up
```

If this command will fail on absent `crawlera-headless-proxy` image,
please build one (follow the instructions to
[crawlera-headless-proxy](https://github.com/scrapinghub/crawlera-headless-proxy)
project). If you need to provide your own API key to Crawlera, please
check `docker-compose.yml` and update the corresponding line to
`headless-proxy` service.

Now everything is up and running. To execute the example, please do the
following:

```console
$ python "$(command -v poetry)" run ./run-example.py
```
