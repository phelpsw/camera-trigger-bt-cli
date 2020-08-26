
# Camera Trigger Bluetooth CLI

A CLI tool for interacting with camera trigger devices over bluetooth.

## Building on macOS
```
brew install golang
brew install dep

# setup GOPATH
mkdir -p $HOME/go/{bin,src}
# Set GOPATH environment variable
# Add export GOPATH="$HOME/go" to ~/.zshrc for example
mkdir -p $HOME/go/src/github.com/phelpsw/

cd $HOME/go/src/github.com/phelpsw/
git clone https://github.com/phelpsw/camera-trigger-bt-cli.git
cd camera-trigger-bt-cli

dep ensure
go build
```

### List Discoverable Devices
```
sudo ./camera-trigger-bt-cli list
```

### Commands Available
```
./camera-trigger-bt-cli --help
```

### Monitor Status
```
sudo ./camera-trigger-bt-cli -d camera-trigger-001 monitor
```