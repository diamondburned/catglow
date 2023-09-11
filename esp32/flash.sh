#!/usr/bin/env bash

fatal() {
	echo "$@" >&2
	exit 1
}

main() {
	package="$1"
	package=$(realpath --relative-to esp32 "$package")

	cd esp32
	tinygo flash -target=esp32-coreboard-v2 -port=/dev/ttyUSB0 "${@:2}" "./$package"
}

main "$@"
