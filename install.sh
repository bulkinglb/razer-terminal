#!/bin/sh
# curl -fsSL https://raw.githubusercontent.com/bulkinglb/razer-terminal/master/install.sh | sh
# VERSION=v0.1.0 picks another release, INSTALL_DIR another folder.
set -eu

repo=bulkinglb/razer-terminal
dir=${INSTALL_DIR:-/usr/local/bin}

if [ "$(uname -s)" != Linux ]; then
	echo "razer-terminal needs OpenRazer, which only runs on Linux" >&2
	exit 1
fi
case $(uname -m) in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*)
	echo "no release for $(uname -m); try: go install github.com/$repo@latest" >&2
	exit 1
	;;
esac

if [ -n "${VERSION:-}" ]; then
	url=https://github.com/$repo/releases/download/$VERSION
else
	url=https://github.com/$repo/releases/latest/download
fi
name=razer-terminal-linux-$arch

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "downloading $name.tar.gz"
curl -fsSL -o "$tmp/$name.tar.gz" "$url/$name.tar.gz"
curl -fsSL -o "$tmp/checksums.txt" "$url/checksums.txt"
if ! (cd "$tmp" && grep " $name.tar.gz\$" checksums.txt | sha256sum -c -) >/dev/null 2>&1; then
	echo "the download doesn't match the release's checksum, not installing" >&2
	exit 1
fi
tar -xzf "$tmp/$name.tar.gz" -C "$tmp"

sudo=
if ! mkdir -p "$dir" 2>/dev/null || [ ! -w "$dir" ]; then
	sudo=sudo
fi
$sudo mkdir -p "$dir"
$sudo install -m 755 "$tmp/$name/razer-terminal" "$dir/razer-terminal"
echo "installed $("$dir/razer-terminal" -version) to $dir"

case ":$PATH:" in
*":$dir:"*) ;;
*) echo "note: $dir isn't on your PATH yet" ;;
esac
