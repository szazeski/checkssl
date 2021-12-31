#!/usr/bin/env bash
# from https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04

package="github.com/szazeski/checkssl"
package_split=(${package//\// })
package_name=${package_split[-1]}

platforms=("windows/amd64" "windows/386" "windows/arm" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/386" "linux/arm64" "linux/arm")

#  go tool dist list | column -c 75 | column -t
#aix/ppc64        freebsd/amd64   linux/mipsle   openbsd/386
#android/386      freebsd/arm     linux/ppc64    openbsd/amd64
#android/amd64    illumos/amd64   linux/ppc64le  openbsd/arm
#android/arm      js/wasm         linux/s390x    openbsd/arm64
#android/arm64    linux/386       nacl/386       plan9/386
#darwin/386       linux/amd64     nacl/amd64p32  plan9/amd64
#darwin/amd64     linux/arm       nacl/arm       plan9/arm
#darwin/arm       linux/arm64     netbsd/386     solaris/amd64
#darwin/arm64     linux/mips      netbsd/amd64   windows/386
#dragonfly/amd64  linux/mips64    netbsd/arm     windows/amd64
#freebsd/386      linux/mips64le  netbsd/arm64   windows/arm

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name $package
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done