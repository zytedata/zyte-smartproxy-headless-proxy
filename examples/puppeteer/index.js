const puppeteer = require("puppeteer");

(async () => {
  const browser = await puppeteer.launch({
    ignoreHTTPSErrors: true,
    args: ["--proxy-server=127.0.0.1:3128"]
  })
  const page = await browser.newPage()
  await page.goto("http://books.toscrape.com")
  console.log(await page.content())
  await browser.close()
})()
