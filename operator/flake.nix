{
  description = "kloudlite operator development setup";

  inputs = {
    nixpkgs = {
      url = "nixpkgs/nixos-unstable";
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShells.default = pkgs.mkShell {
          hardeningDisable = ["all"];
          buildInputs = with pkgs; [
            # cli tools
            curl
            jq
            yq
            go-task

            # source version control
            git
            pre-commit

            # programming tools
            go
            kubebuilder
            mongosh

            # kubernetes specific tools
            # k9s
            kubectl
            kubernetes-helm
            velero
            natscli

            # grpc tools
            protobuf
            protoc-gen-go
            protoc-gen-go-grpc

            # build tools
            gnumake
            podman
            upx
          ];

          shellHook = ''
            KUBEBUILDER_ASSETS="$PWD/bin/k8s/1.31.0-linux-amd64"
            export PATH="$PWD/bin:$PATH"
          '';
        };
      }
    );
}
