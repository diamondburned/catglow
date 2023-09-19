{ lib, stdenv, fetchzip, go, runCommandLocal, makeWrapper }:

let
	version = "0.29.0";

	createURL = { GOOS, GOARCH, ... }:
		let
			base = "https://github.com/tinygo-org/tinygo/releases/download";
			name = "tinygo${version}.${GOOS}-${GOARCH}.tar.gz";
		in
			"${base}/v${version}/${name}";
in

runCommandLocal "tinygo-${version}" {
	inherit version;
	src = fetchTarball (createURL go);
	nativeBuildInputs = [ makeWrapper ];
} ''
	cp --no-preserve=mode,ownership -r $src $out
	chmod +x $out/bin/*
''
