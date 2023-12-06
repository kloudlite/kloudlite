{
  description = "kloudlite api dev environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages = {
          # mocki = pkgs.writeScriptBin "mocki" ''
          #   ./cmd/mocki/bin/mocki
          # '';
        };
        formatter = pkgs.nixpkgs-fmt;
        devShells.default = pkgs.mkShell {
          packages = [
            # self.packages.${system}.mocki
          ];
          hardeningDisable = [ "all" ];

          buildInputs = with pkgs; [
            # cli tools
            curl
            jq
            yq
            go-task

            # source version control
            git
            pre-commit
            (python312.withPackages(ps: with ps; [
              ggshield
            ]))


            # programming tools
            go_1_21
            operator-sdk
            mongosh
            natscli

            # kubernetes specific tools
            k9s
            kubectl
            kubernetes-helm

            # grpc tools
            protobuf
            protoc-gen-go
            protoc-gen-go-grpc

            # build tools
            podman
            upx
          ];

          shellHook = ''
            export PATH="$PWD/cmd/mocki/bin:$PATH" # mocki binary
            # exec fish # -- not needed if using direnv as it will automatically load current shell
          '';
        };
      }
    );
}
