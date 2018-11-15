-- This is a simple LUA script which does almost the same as render.html
-- endpoint. Please check Splash docs to get the details:
--   * https://splash.readthedocs.io/en/stable/api.html
--   * https://splash.readthedocs.io/en/stable/scripting-tutorial.html
--   * https://splash.readthedocs.io/en/stable/scripting-overview.html
--   * https://splash.readthedocs.io/en/stable/scripting-ref.html

function main(splash, args)
  if args.proxy_host ~= nil and args.proxy_port ~= nil then
    splash:on_request(function(request)
      request:set_proxy{
        host = args.proxy_host,
        port = args.proxy_port,
      }
    end)
  end

  splash:set_result_content_type("text/html; charset=utf-8")
  assert(splash:go(args.url))
  return splash:html()
end
