class MordecaiCli < Formula
  desc "CLI tool for Mordecai"
  homepage "https://github.com/yourusername/mordecai-cli"
  url "file:///Users/david/CodeYard/mordecai/go-test/mordecai-app-v1.0.0.tar.gz"
  sha256 "4b22afe88c4688176b6da4fc2d2ef6318c190e625c0de4fe7bbaa8d98c90aeda"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"mordecai-app"
  end

  test do
    assert_match "Mordecai CLI Usage", shell_output("#{bin}/mordecai-app --help")
  end
end
