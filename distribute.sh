#!/usr/bin/env bash

checkCommand() {
    which "$1" >/dev/null 2>&1
    if [ "$?" != "0" ]; then
        echo Please make sure the following command is available: "$1" >&2
        exit "$?"
    fi
}

checkCommand go
checkCommand tar
checkCommand zip

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# NOTE: The build process is somewhat similar to https://github.com/syncthing/syncthing, thanks for that!
platforms=(
    linux-amd64 windows-amd64 darwin-amd64 dragonfly-amd64 freebsd-amd64 netbsd-amd64 openbsd-amd64 solaris-amd64
    freebsd-386 linux-386 netbsd-386 openbsd-386 windows-386
    linux-arm linux-arm64 linux-ppc64 linux-ppc64le
)

cd "$DIR"
commit="$(git rev-parse --short HEAD 2>/dev/null)"
date="$(date +"%Y-%m-%d_%H:%M:%S")"

if [ "$commit" == "" ]; then
    commit="unknown"
fi

if [ "$1" == "" ]; then
    echo You didn\'t provide a version string as the first parameter, setting version to \"unknown\".
    version="unknown"
else
    version="$1"
fi

rm -Rf ./bin/
mkdir ./bin/ 2>/dev/null

for plat in "${platforms[@]}"; do
    echo Building "$plat" ...

    GOOS="${plat%-*}"
    GOARCH="${plat#*-}"

    if [ "$GOOS" != "windows" ]; then
        tmpFile="/tmp/srrdb/bin/srrdb"
    else
        tmpFile="/tmp/srrdb/bin/srrdb.exe"
    fi

    CGO_ENABLED=0 GOOS="${plat%-*}" GOARCH="${plat#*-}" go build -ldflags '-X main.VERSION='"$version"' -X main.BUILD_COMMIT='"$commit"' -X main.BUILD_DATE='"$date" \
    -o "$tmpFile" "$DIR"/*.go

    if [ "$?" != 0 ]; then
        echo Build failed! >&2
        exit "$?"
    fi

    if [ "$GOOS" != "windows" ]; then
        tarPath="$DIR"/bin/srrdb-"$plat".tar.gz
        echo Build succeeded, creating "$tarPath" ...
        tar -czf "$tarPath" -C "${tmpFile%/*}" srrdb
    else
        zipPath="$DIR"/bin/srrdb-"$plat".zip
        echo Build succeeded, creating "$zipPath" ...
        zip -j "$zipPath" "$tmpFile"
    fi

    if [ "$?" != 0 ]; then
        echo Failed to pack the binary! >&2
        exit "$?"
    fi
    echo Done!

    rm "$tmpFile"

    echo
done
