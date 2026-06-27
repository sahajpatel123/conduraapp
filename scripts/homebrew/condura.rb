cask "condura" do
  version "0.1.1"
  sha256 "9bce357d396799d7be04e116691fff5fec428cd441145e563df93796e487158f"

  url "https://github.com/sahajpatel123/conduraapp/releases/download/v0.1.1/condura-gui-darwin-arm64.dmg",
      verified: "github.com/sahajpatel123/conduraapp/"
  name "Condura"
  desc "Free, local-first AI agent that lives on your computer and orchestrates every other AI tool"
  homepage "https://condura.app"

  livecheck do
    url :url
    strategy :github_latest
  end

  depends_on macos: ">= :ventura"

  app "Condura.app"

  zap trash: [
    "~/Library/Application Support/condura",
    "~/.condura",
    "~/Library/Preferences/com.condura.app.plist",
  ]
end
