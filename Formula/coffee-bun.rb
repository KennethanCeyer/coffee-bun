class CoffeeBun < Formula
  desc "Automated macOS MacBook setup & environment maintenance tool"
  homepage "https://github.com/KennethanCeyer/coffee-bun"
  url "https://github.com/KennethanCeyer/coffee-bun/archive/refs/tags/v1.0.0.tar.gz"
  license "MIT"
  head "https://github.com/KennethanCeyer/coffee-bun.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w", output: bin/"bun"), "./cmd/bun"
    bin.install_symlink bin/"bun" => "coffee-bun"
  end

  test do
    assert_match "coffee-bun (bun) version", shell_output("#{bin}/bun version")
  end
end
