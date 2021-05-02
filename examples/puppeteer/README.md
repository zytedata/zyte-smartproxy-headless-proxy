# Puppeteer example

This repository contains a simple example of how to work with headless
Chrome using [Puppeteer](https://pptr.dev), official JS API for headless
Chrome.

To run this example, please follow [official
instructions](https://poetry.eustace.io/docs/#installation) on how to
install Poetry. After that, please prepare the environment with the
following command:

```console
$ python3 "$(command -v poetry)" install
```

Also, you need to have node.js and npm installed to be
able to use Puppeteer. Please [follow the official install
guide](https://www.npmjs.com/get-npm) to obtains both tools. After you
will have them, please install Puppeteer with the following command:

```console
$ npm install
```

Now let's run headless proxy:

```console
$ python3 "$(which poetry)" run docker-compose up
```

If this command will fail on absent `zyte-smartproxy-headless-proxy` image,
please build one (follow the instructions to
[zyte-smartproxy-headless-proxy](https://github.com/zytedata/zyte-smartproxy-headless-proxy)
project). If you need to provide your own API key to Zyte Smart Proxy Manager,
please check `docker-compose.yml` and update the corresponding line to
`headless-proxy` service.

Now everything is up and running. To execute the example, please do the
following:

```console
$ npm run example
```

or

```console
$ node index.js
```
