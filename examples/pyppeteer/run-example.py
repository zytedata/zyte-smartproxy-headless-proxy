#!/usr/bin/env python3
# vim: set et sw=4 ts=8:


import asyncio
import pyppeteer


async def main():
    browser = await pyppeteer.launch(
        ignoreHTTPSErrors=True,
        args=["--proxy-server=127.0.0.1:3128"]
    )
    page = await browser.newPage()
    await page.goto("http://books.toscrape.com")
    print(await page.content())
    await browser.close()


asyncio.get_event_loop().run_until_complete(main())
