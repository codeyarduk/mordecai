class Mordecai < Formula
  desc "CLI tool for Mordecai"
  homepage "https://github.com/yourusername/mordecai-cli"
  url "file:///Users/david/CodeYard/mordecai/cli/mordecai-cli-go/mordecai-cli-1.0.0.tar.gz"
  sha256 "27f7441208acb6204559d621bb19f33c022ed6d8637df1a6e132a854a6aeea85"

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
