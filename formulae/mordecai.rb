class Mordecai < Formula
  desc "CLI tool for Mordecai"
  homepage "https://github.com/codeyarduk/mordecai"
  url "file:///Users/david/CodeYard/mordecai/cli/mordecai/mordecai-cli-1.0.0.tar.gz"
  sha256 "9a026d8971caaff2109c03cbda62d1f41533136600e261a66093674c7d348ea2"

  depends_on "go" => :build

  def install
     ENV["HOMEBREW_NO_SANDBOX"] = "1"
     cd "cmd/mordecai" do
       system "go", "build", "-o", "mordecai"
       bin.install "mordecai"
     end
  end  
  test do
    assert_match "Mordecai CLI Usage", shell_output("#{bin}/mordecai --help")
  end
end
