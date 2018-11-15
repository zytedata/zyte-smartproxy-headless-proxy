# Examples

This main purpose of this directory is to provide a set of examples, how
to work with Crawlera and headless browsers using different technologies
like Puppeteer or Selenium.

All these examples assume following:

1. crawlera-headless-proxy runs with either docker or standalone binary.
2. crawlera-headless-proxy HTTP interface is accessible with `127.0.0.1:3128`
   host and port pairs.
3. crawlera-headless-proxy can access Crawlera with correct API key.

Please find an example of such config near this README (file
`example-config.toml`). `USER_API_KEY` is Crawlera API key.

To run the proxy with the given config, please execute it with Docker:

```console
$ docker run -i --rm -p 3128:3128 -v $(pwd)/example-config.toml:/config.toml:ro crawlera-headless-proxy
```

Or as a standalone binary:

```console
$ crawlera-headless-proxy -c $(pwd)/example-config.toml
```

Examples also provide docker-compose configuration files which also have
headless-proxy as a container. Please check them if you do not want to
run this tool separately.
