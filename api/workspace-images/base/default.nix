{ pkgs ? import <nixpkgs> {} }:

pkgs.buildEnv {
  name = "workspace-default-env";
  paths = with pkgs; [
    # Core utilities
    coreutils
    findutils
    gnused
    gnugrep
    gawk

    # Development tools
    git
    curl
    wget
    vim
    nano

    # Build tools
    gcc
    gnumake
    cmake

    # Common language runtimes (will be overridden by workspace-specific configs)
    # These are here as examples - users will typically define their own

  ];
}
