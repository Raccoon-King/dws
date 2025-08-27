$goBinPath = "$($env:USERPROFILE)\go\bin"

# Check if the path already exists to avoid duplicates
if (-not ($env:Path.Split(';') -contains $goBinPath)) {
    # Add to current session's PATH
    $env:Path += ";$goBinPath"
    Write-Host "Added '$goBinPath' to current session's PATH."

    # Add to system-wide PATH for persistence
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
    [System.Environment]::SetEnvironmentVariable("Path", "$currentPath;$goBinPath", "Machine")
    Write-Host "Added '$goBinPath' to system-wide PATH. Please restart your terminal for changes to take effect."
} else {
    Write-Host "'$goBinPath' is already in the system PATH."
}