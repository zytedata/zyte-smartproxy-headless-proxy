#!/usr/bin/env python
# vim: set et sw=4 ts=8:


from __future__ import print_function

from selenium import webdriver


ZYTE_SMARTPROXY_HEADLESS_PROXY = "proxy:3128"

profile = webdriver.DesiredCapabilities.CHROME.copy()
profile["proxy"] = {
    "httpProxy": ZYTE_SMARTPROXY_HEADLESS_PROXY,
    "ftpProxy": ZYTE_SMARTPROXY_HEADLESS_PROXY,
    "sslProxy": ZYTE_SMARTPROXY_HEADLESS_PROXY,
    "noProxy": None,
    "proxyType": "MANUAL",
    "class": "org.openqa.selenium.Proxy",
    "autodetect": False
}
profile["acceptSslCerts"] = True


driver = webdriver.Remote("http://localhost:4444/wd/hub", profile)
driver.get("http://books.toscrape.com/")
page_source = driver.page_source

if isinstance(page_source, bytes):
    page_source = page_source.decode("utf-8")

print(page_source)
