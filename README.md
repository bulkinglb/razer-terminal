# razer-terminal

A terminal UI for Razer devices on Linux: headset EQ, sidetone and game/chat balance, mouse DPI, polling rate, scroll wheel and lighting, all in one place. Built on [OpenRazer](https://openrazer.github.io).

Tested with a BlackShark V3 (2.4 GHz dongle) and a Basilisk V3.

## Setup

Mice and keyboards need OpenRazer and its daemon:

    sudo pacman -S openrazer-daemon openrazer-driver-dkms
    sudo gpasswd -a $USER plugdev
    systemctl --user enable --now openrazer-daemon

The BlackShark V3 isn't in OpenRazer yet. Build it from [PR #2794](https://github.com/openrazer/openrazer/pull/2794) instead of the packages above, and add this kernel option so the dongle accepts settings (kernel 7.2, 7.1.9 or 6.18.45 and newer):

    usbcore.quirks=1532:057a:r,1532:0579:r

## Install

    curl -fsSL https://raw.githubusercontent.com/bulkinglb/razer-terminal/master/install.sh | sh

This puts the latest release in `/usr/local/bin`; end it with `| INSTALL_DIR=~/.local/bin sh` to skip sudo. With Go: `go install github.com/bulkinglb/razer-terminal@latest`. From a checkout, `make install` puts it in `~/.local/bin`.

## Use

    razer-terminal

↑↓ select, ←→ change, shift+←→ fine steps, tab switches device, r rereads, q quits. Changes apply right away. A headset value shows UNKNOWN until the headset reports it or you set it. Setting a DPI stage to OFF removes it, and editing a stage switches the mouse to it.

MIT license.
