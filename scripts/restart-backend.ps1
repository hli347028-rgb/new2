$ErrorActionPreference = "Stop"

$workspace = "C:\Users\86186\Desktop\new2"
$serverExe = Join-Path $workspace "bin\server.exe"
$expectedPath = (Resolve-Path -LiteralPath $serverExe).Path

$listenerPids = Get-NetTCPConnection -State Listen |
    Where-Object { $_.LocalPort -in 9000, 9100 } |
    Select-Object -ExpandProperty OwningProcess -Unique

foreach ($listenerPid in $listenerPids) {
    $process = Get-Process -Id $listenerPid -ErrorAction SilentlyContinue
    if ($process -and $process.Path -eq $expectedPath) {
        Stop-Process -Id $listenerPid -Force
    }
}

Start-Sleep -Seconds 2
Start-Process -FilePath $serverExe `
    -ArgumentList "-conf", "configs" `
    -WorkingDirectory $workspace `
    -RedirectStandardOutput (Join-Path $workspace "logs\backend.out.log") `
    -RedirectStandardError (Join-Path $workspace "logs\backend.err.log") `
    -WindowStyle Hidden
