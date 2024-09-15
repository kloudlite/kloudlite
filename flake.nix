{
  description = "kloudlite Infrastructure as Code development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };

        binaries = with pkgs; [
          bash
          # coreutils
          # gettext
          envsubst
          # busybox-sandbox-shell
          jq
          # lz4
          zstd
          kubectl
          terraform
        ];

        container = pkgs.buildEnv {
          name = "container2";
          paths = with pkgs; [
            bash
            envsubst
            jq
            zstd
            kubectl
            terraform
          ];
        };

        gcp-spot-node-terminator = pkgs.buildEnv {
          name = "gcp-spot-node-terminator";
          paths = with pkgs; [
            bash
            curl
            zx
            kubectl
          ];
        };
      in
      {
        devShells.default = pkgs.mkShell {
          # hardeningDisable = [ "all" ];
          buildInputs = with pkgs; [
            # cli tools
            bash
            go-task

            terraform
            packer
            pulumi
            pulumiPackages.pulumi-language-go

            # source version control
            git
            pre-commit

            # programming tools

            # build tools
            # podman
            upx

            nmap
            zx

            # # custom
            # packages.new-infra
            # new-infra
          ];

          shellHook = ''
            export TF_PLUGIN_CACHE_DIR="$PWD/.terraform.d/plugin-cache"
            export PATH="$PWD/cmd:$PATH"
            mkdir -p $TF_PLUGIN_CACHE_DIR
          '';
        };

        packages.container = pkgs.stdenv.mkDerivation {
          name = "container";
          src = container;
          installPhase = "cp -r $src $out/";
        };
      }
    );
}


