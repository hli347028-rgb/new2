$ErrorActionPreference = "Stop"

netsh advfirewall firewall delete rule name="AIX Nginx HTTP 8081" | Out-Null
netsh advfirewall firewall add rule name="AIX Nginx HTTP 8081" dir=in action=allow protocol=TCP localport=8081 profile=any remoteip=any

if ($LASTEXITCODE -ne 0) {
    throw "Failed to create firewall rule for TCP 8081."
}
