#!/usr/bin/env zsh

set -e
set -u
set -x

go build -o strager-portal .
sudo launchctl unload /Library/LaunchDaemons/net.strager.portal.plist 2>/dev/null || :
sudo cp ./net.strager.portal.plist /Library/LaunchDaemons/net.strager.portal.plist
sudo launchctl load /Library/LaunchDaemons/net.strager.portal.plist

sudo launchctl list net.strager.portal
