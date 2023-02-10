{ pkgs, ... }:

{
  packages = [ pkgs.git ];

  enterShell = ''
    echo "Welcome to your homelab"
  '';

  languages.go.enable = true;
}
