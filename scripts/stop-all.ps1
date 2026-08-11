$ErrorActionPreference = "Stop"

$projectPorts = 8081, 9000, 9100, 9200, 9300
$allowedNames = "nginx", "server", "node"

$listenerPids = Get-NetTCPConnection -State Listen |
    Where-Object { $_.LocalPort -in $projectPorts } |
    Select-Object -ExpandProperty OwningProcess -Unique

foreach ($listenerPid in $listenerPids) {
    $process = Get-Process -Id $listenerPid -ErrorAction SilentlyContinue
    if ($process -and $process.ProcessName -in $allowedNames) {
        Stop-Process -Id $listenerPid -Force
    }
}

# Nginx master can respawn a worker after only the listening worker is stopped.
Get-Process -Name "nginx" -ErrorAction SilentlyContinue | Stop-Process -Force
