{ pkgs, ... }:

{
  packages = [ pkgs.git pkgs.mosquitto pkgs.timescaledb ];

  enterShell = ''
    echo "Welcome to your homelab"
  '';

  languages.go.enable = true;
}
