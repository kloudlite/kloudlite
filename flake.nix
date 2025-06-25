{
  description = "kloudlite workspace";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {
          inherit system;
          # config.allowUnfree = true;
        };
        archMap = {
          "x86_64" = "amd64";
          "aarch64" = "arm64";
        };

        # Parse arch and os from system string like "x86_64-linux"
        arch = builtins.getAttr (builtins.elemAt (builtins.split "-" system) 0) archMap;
        os = builtins.elemAt (builtins.split "-" system) 2;
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            (stdenv.mkDerivation rec {
              name = "go-template";
              pname = "go-template";
              src = fetchurl {
                url = "https://github.com/nxtcoder17/go-template/releases/download/v0.1.0/go-template-${os}-${arch}";
                sha256 = builtins.getAttr arch {
                  "amd64" = "sha256-I7194+7GKmgO5wHs2Fiqa6S5KVnrqQZbyI37Lmjw9jM=";
                  "arm64" = "sha256-LBLmixoyqu4ARqxE251omY68ts9BHRUv5rVBvn/xqWM=";
                };
              };
              unpackPhase = ":";
              installPhase = ''
                mkdir -p $out/bin
                cp $src $out/bin/$name
                chmod +x $out/bin/$name
              '';
            })

            # your packages here
            mkcert
          ];

          shellHook = ''
            alias k=kubectl
              mkdir -p .secrets
          '';
        };
      }
    );
}
