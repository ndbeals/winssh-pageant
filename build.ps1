param (
    [string]
    $BuildPath = ".\build",

    [string[]]
    $Architectures = @("amd64","386"),

    [switch]
    $Release = $false,

    [string]
    $ReleasePath = "winssh-pageant",

    [string]
    $ver = ""
)

# Cleanup
Remove-Item -LiteralPath $BuildPath -Force -Recurse -ErrorAction SilentlyContinue

# Build output directory
$outDir = New-Item -ItemType Directory -Force -Path $BuildPath
$releaseDir = New-Item -ItemType Directory -Force -Path ".\release"

$oldGOOS = $env:GOOS
$oldGOARCH = $env:GOARCH

$env:GOOS="windows"
$env:GOARCH=$null

$returnValue = 0

# Build release package
if ($Release)
{
    Copy-Item Readme.md $outDir
    Copy-Item LICENSE $outDir
    
    Remove-Item -LiteralPath $ReleasePath -ErrorAction SilentlyContinue

    Write-Output "
## Checksums
| Architecture | Checksum |
|---|---|" | Out-File -FilePath checksums.md -Encoding utf8 
}

# Build for each architecture
Foreach ($arch in $Architectures)
{
    $env:GOARCH=$arch
    
    if ($Release)
    {        
        go build -ldflags -H=windowsgui -trimpath -o $outDir\winssh-pageant.exe
        if ($LastExitCode -ne 0) { $returnValue = $LastExitCode }
        Compress-Archive -Path $outDir\* -DestinationPath $releaseDir\$ReleasePath-${ver}_$arch.zip -Force

        $hash = (Get-FileHash $outDir\winssh-pageant.exe).Hash 
        Write-Output "| $arch | $hash |" | Out-File -FilePath checksums.md -Encoding utf8 -Append
        
        Remove-Item -LiteralPath $outDir\winssh-pageant.exe
    } else {
        go build -ldflags -H=windowsgui -trimpath -o $outDir\winssh-pageant-$arch.exe
    }
}

# Restore env vars
$env:GOOS = $oldGOOS
$env:GOARCH = $oldGOARCH

# Cleanup
if ($Release)
{
    Write-Output "" | Out-File -FilePath checksums.md -Encoding utf8 -Append
    Remove-Item -LiteralPath $BuildPath -Force -Recurse -ErrorAction SilentlyContinue
}

exit $returnValue
