@echo off
REM Quick Cloudflare Tunnel to local nginx (Docker :80).
REM 530 = cloudflared is not running or URL is from an OLD run — always use the NEW https://....trycloudflare.com from this window.
REM For a stable URL use a named tunnel: cloudflared tunnel login -> create in Zero Trust -> cloudflared tunnel run

set "CF=%ProgramFiles(x86)%\cloudflared\cloudflared.exe"
if not exist "%CF%" set "CF=%ProgramFiles%\cloudflared\cloudflared.exe"
if not exist "%CF%" (
  echo cloudflared.exe not found. Install: winget install Cloudflare.cloudflared
  pause
  exit /b 1
)

echo Starting tunnel to http://127.0.0.1:80 (HTTP/2 to edge, IPv4^) ...
echo Keep THIS window open during the demo. Copy the trycloudflare URL from the lines below.
echo.
"%CF%" tunnel --url http://127.0.0.1:80 --edge-ip-version 4 --protocol http2
pause
