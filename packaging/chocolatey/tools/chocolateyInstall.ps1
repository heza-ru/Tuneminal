$ErrorActionPreference = 'Stop'

$packageName = 'tuneminal'
$url = 'https://github.com/tuneminal/tuneminal/releases/download/v1.0.0/tuneminal-windows-amd64.exe'
$checksum = 'PLACEHOLDER_CHECKSUM'
$checksumType = 'sha256'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$filePath = Join-Path $toolsDir "tuneminal.exe"

Get-ChocolateyWebFile -PackageName $packageName -FileFullPath $filePath -Url $url -Checksum $checksum -ChecksumType $checksumType

# Create shim
Install-BinFile -Name "tuneminal" -Path $filePath

