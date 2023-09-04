{ lib, stdenv, fetchzip, go_1_20, runCommandLocal, makeWrapper }:

let
	version = "0.28.1";

	createURL = { GOOS, GOARCH, ... }:
		let
			base = "https://github.com/tinygo-org/tinygo/releases/download";
			name = "tinygo${version}.${GOOS}-${GOARCH}.tar.gz";
		in
			"${base}/v${version}/${name}";
in

runCommandLocal "tinygo-${version}" {
	inherit version;
	src = fetchTarball (createURL go_1_20);
	nativeBuildInputs = [ makeWrapper ];
} ''
	cp --no-preserve=mode,ownership -r $src $out
	chmod +x $out/bin/*
	wrapProgram $out/bin/tinygo \
		--prefix PATH : "${lib.makeBinPath [ go_1_20 ]}" \
		--set GOROOT ${go_1_20}/share/go \
		--set TINYGOROOT $out
''
