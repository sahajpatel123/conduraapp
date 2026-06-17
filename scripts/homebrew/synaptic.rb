cask "synaptic" do
  version "0.1.0"
  sha256 :no_check

  url "https://github.com/sahajpatel123/synapticapp/releases/latest/download/synaptic-gui-darwin-arm64.dmg",
      verified: "github.com/sahajpatel123/synapticapp/"
  name "Synaptic"
  desc "Free, local-first AI agent that lives on your computer and orchestrates every other AI tool"
  homepage "https://synaptic.app"

  livecheck do
    url :url
    strategy :github_latest
  end

  depends_on macos: ">= :ventura"

  app "Synaptic.app"

  zap trash: [
    "~/Library/Application Support/synaptic",
    "~/.synaptic",
    "~/Library/Preferences/com.synaptic.app.plist",
  ]
end
