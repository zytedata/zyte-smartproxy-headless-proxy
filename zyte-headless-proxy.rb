# vim: set ft=ruby et sw=2 ts=4:


# This is Homebrew (https://brew.sh) formula to install Zyte Smart Proxy Manager
# Headless Proxy on your Mac. Installation is quite straghtforward:
#
# To install release version, please run:
#   $ brew install https://raw.githubusercontent.com/zytedata/zyte-headless-proxy/master/zyte-headless-proxy.rb
#
# If you wanr to install development version, please use HEAD:
#   $ bew install --HEAD https://raw.githubusercontent.com/zytedata/zyte-headless-proxy/master/zyte-headless-proxy.rb
#
# Also, there is one optional parameter: --with-upx
# With this parameter, headless-proxy will be compressed by UPX
# (https://upx.github.io). This will give you a smaller binary size but
# as a side-effect, installation time will be much longer than simple
# installation.


class SmartProxyManagerHeadlessProxy < Formula
  desc "Complimentary user proxy for headless browsers to work with Zyte Smart Proxy Manager"
  homepage "https://github.com/zytedata/zyte-headless-proxy"

  revision 0
  url "https://github.com/zytedata/zyte-headless-proxy.git", :using => :git, :tag => "1.2.1"
  head "https://github.com/zytedata/zyte-headless-proxy.git"

  depends_on "go" => :build
  depends_on "make" => :build

  option "with-upx", "Compress binary with upx"
  depends_on "upx" => [:build, :optional]

  def install
    system "make"
    system "upx", "--ultra-brute", "-qq", "./zyte-headless-proxy" if build.with? "upx"
    bin.install "zyte-headless-proxy"
  end

  test do
    system "#{bin}/zyte-headless-proxy", "--version"
  end
end
