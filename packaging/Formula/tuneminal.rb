class Tuneminal < Formula
  desc "Command-line karaoke machine with live audio visualization"
  homepage "https://github.com/tuneminal/tuneminal"
  url "https://github.com/tuneminal/tuneminal/releases/download/v1.0.0/tuneminal-darwin-amd64.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "MIT"
  head "https://github.com/tuneminal/tuneminal.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-s -w", "-o", "tuneminal", "cmd/tuneminal/main.go"
    bin.install "tuneminal"
  end

  test do
    system "#{bin}/tuneminal", "--help"
  end
end

