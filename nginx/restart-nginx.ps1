$ErrorActionPreference = "Stop"

$nginxExe = "D:\5\nginx-1.31.3\nginx.exe"
$projectPrefix = "C:\Users\86186\Desktop\new2\nginx\"

# Stop both the installer-default instance and the project instance.
& $nginxExe -p "D:\5\nginx-1.31.3\" -c "conf/nginx.conf" -s stop 2>$null
& $nginxExe -p $projectPrefix -c "nginx.conf" -s stop 2>$null
Start-Sleep -Seconds 2

& $nginxExe -p $projectPrefix -c "nginx.conf" -t
if ($LASTEXITCODE -ne 0) {
    throw "Nginx configuration test failed."
}

& $nginxExe -p $projectPrefix -c "nginx.conf"
