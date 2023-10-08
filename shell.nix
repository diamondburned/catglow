{}:

let
	pkgs = import ./nix/pkgs.nix {
		overlays = [
			(self: super: {
				go = super.go_1_21;
			})
		];
	};
in

with pkgs.lib;
with builtins;

let
	tinygo = pkgs.callPackage ./nix/tinygo.nix {};
	# tinygo = pkgs.tinygo;

	# Tinygo target for gopls to use.
	tinygoPaths = [ "xiao" "esp32" ];
	tinygoTargets = {
		"xiao"  = "xiao-rp2040";
		"esp32" = "esp32-coreboard-v2";
	};

	tinygoHook =
		with pkgs.lib;
		with builtins;
		''
			declare -A tinygoTargets
			${builtins.concatStringsSep " "
				(mapAttrsToList
					(name: target: "tinygoTargets[${escapeShellArg name}]=${escapeShellArg target}")
					(tinygoTargets)
				)
			}
			tinygoPaths=(
				${builtins.concatStringsSep " "
					(map (escapeShellArg) tinygoPaths)
				}
			)

			isTinygo() {
				root=${escapeShellArg (toString ./.)}
				path="''${PWD#"$root/"*}"

				for p in $tinygoPaths; do
					if [[ $path == $p* ]]; then
						export tinygoPath=$p
						export tinygoTarget=''${tinygoTargets["$p"]}
						return 0
					fi
				done

				return 1
			}

			hookTinygoEnv() {
				vars=$(tinygo info -json -target $tinygoTarget)

				export GOROOT=$(jq -r '.goroot' <<< "$vars")
				export GOARCH=$(jq -r '.goarch' <<< "$vars")
				export GOOS=$(jq -r '.goos' <<< "$vars")
				export GOFLAGS="-tags=$(jq -r '.build_tags | join(",")' <<< "$vars")"
				export GOWORK=
			}
		'';

	withTinygoHook = name: bin:
		pkgs.writeShellScriptBin name ''
			${tinygoHook}
			if isTinygo; then
				echo "Detected Tinygo, loading for target $tinygoTarget" >&2
				hookTinygoEnv
			fi
			exec ${bin} "$@"
		'';

	# go = withTinygoHook "go" "${pkgs.go}/bin/go";
	gopls = withTinygoHook "gopls" "${pkgs.gopls}/bin/gopls";
	goimports = withTinygoHook "goimports" "${pkgs.gotools}/bin/goimports";

	staticcheck = pkgs.writeShellScriptBin "staticcheck" ''
		${tinygoHook}
		if isTinygo; then
			echo "Not running staticcheck for Tinygo" >&2
			exit 0
		fi
		exec ${pkgs.go-tools}/bin/staticcheck "$@"
	'';
in

pkgs.mkShell {
	buildInputs = with pkgs; [
		niv
		go
		gopls
		gotools
		go-tools # staticcheck
		goimports
		staticcheck
		esptool
	] ++ [ tinygo ];

	CGO_ENABLED = "1";
}
