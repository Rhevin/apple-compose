class AppleCompose < Formula
  desc "Docker Compose-compatible orchestrator for Apple Containers"
  homepage "https://github.com/Rhevin/apple-compose"
  url "https://github.com/Rhevin/apple-compose/archive/refs/tags/v0.1.0.tar.gz"
  # Update sha256 after first release: `sha256 $(curl -sL <url> | shasum -a 256 | cut -d' ' -f1)`
  sha256 "REPLACE_WITH_SHA256_ON_RELEASE"
  license "MIT"
  head "https://github.com/Rhevin/apple-compose.git", branch: "main"

  depends_on "go" => :build
  depends_on :macos => :sequoia  # macOS 15+
  depends_on arch: :arm64        # Apple Silicon only

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X main.version=#{version}"), "."
  end

  test do
    # Verify binary runs and shows help
    assert_match "Docker Compose-compatible", shell_output("#{bin}/apple-compose --help")
  end
end
