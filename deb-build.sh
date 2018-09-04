#!/bin/sh

# [version tag]-[commits since tag]-[commit hash][-dirty]
VERSION=$(git describe --tags --long --dirty)

if [ -z "${VERSION}" ]; then
    echo "Could not detect version tag"
    exit 1
fi

echo "Building version ${VERSION}"

PKG_ROOT=pkg_root

mkdir -p ${PKG_ROOT}/usr/local/bin
mkdir -p ${PKG_ROOT}/usr/local/etc/go-Musicbot

cp out/binaries/MusicBot-linux-amd64 \
    ${PKG_ROOT}/usr/local/bin/go-Musicbot

cp dist/config.json \
    ${PKG_ROOT}/usr/local/etc/go-Musicbot/config.json

fpm \
	-n go-musicbot \
	-C ${PKG_ROOT} \
	-s dir \
	-t deb \
	-v "${VERSION}" \
	--force \
	--deb-compression bzip2 \
	--license MIT \
	-m "Sven Wiltink" \
	--url "https://github.com/svenwiltink/go-musicbot" \
	--description "A musicbot for rocketchat or irc" \
	--deb-systemd dist/go-musicbot.service \
	--config-files /usr/local/etc/go-Musicbot \
    --after-install dist/after.sh \
    --after-upgrade dist/after.sh 

