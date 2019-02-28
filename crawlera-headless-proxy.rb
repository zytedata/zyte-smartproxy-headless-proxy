# vim: set ft=ruby et sw=2 ts=4:


# This is Homebrew (https://brew.sh) formula to install Crawlera
# Headless Proxy on your Mac. Installation is quite straghtforward:
#
# To install release version, please run:
#   $ brew install https://raw.githubusercontent.com/scrapinghub/crawlera-headless-proxy/master/crawlera-headless-proxy.rb
#
# If you wanr to install development version, please use HEAD:
#   $ bew install --HEAD https://raw.githubusercontent.com/scrapinghub/crawlera-headless-proxy/master/crawlera-headless-proxy.rb
#
# Also, there is one optional parameter: --with-upx
# With this parameter, headless-proxy will be compressed by UPX
# (https://upx.github.io). This will give you a smaller binary size but
# as a side-effect, installation time will be much longer than simple
# installation.


class CrawleraHeadlessProxy < Formula
  desc "Complimentary user proxy for headless browsers to work with Crawlera"
  homepage "https://github.com/scrapinghub/crawlera-headless-proxy"

  revision 0
  url "https://github.com/scrapinghub/crawlera-headless-proxy.git", :using => :git, :tag => "1.1.1"
  head "https://github.com/scrapinghub/crawlera-headless-proxy.git"

  depends_on "go" => :build
  depends_on "make" => :build

  option "with-upx", "Compress binary with upx"
  depends_on "upx" => [:build, :optional]

  def install
    system "make"
    system "upx", "--ultra-brute", "-qq", "./crawlera-headless-proxy" if build.with? "upx"
    bin.install "crawlera-headless-proxy"
  end

  test do
    system "#{bin}/crawlera-headless-proxy", "--version"
  end
end
