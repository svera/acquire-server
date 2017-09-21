#!/bin/ash
wget -qO /usr/lib/sackson-server/acquire.so $(curl -s https://api.github.com/repos/svera/acquire-sackson-driver/releases/latest | grep "browser_download_url.*so" | cut -d '"' -f 4 )