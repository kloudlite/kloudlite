{
  description = "kloudlite api dev environment";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable"; };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.mocki = pkgs.writeScriptBin "mocki" ''
          #! /usr/bin/env bash
          $PROJECT_ROOT/cmd/mocki/bin/mocki "$@"
        '';
        packages.nats-manager = pkgs.writeScriptBin "nats-manager" ''
          $PROJECT_ROOT/cmd/nats-manager/bin/nats-manager --url "nats://nats.kloudlite.svc.cluster.local:4222" --stream "resource-sync" "$@"
        '';
        defaultPackage = self.packages.${system}.mocki;

        devShells.default = pkgs.mkShell {
          packages = [
            self.packages.${system}.mocki
            # self.packages.${system}.nats-manager
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
            (python312.withPackages (ps: with ps; [
              ggshield
            ]))

            # programming tools
            go_1_23
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
            upx

            rover
            # node
            nodePackages.pnpm
          ];

          shellHook = ''
            # export PATH="$PWD/cmd/mocki/bin:$PATH"
            export PROJECT_ROOT="$PWD"
          '';
        };
      }
    );
}
