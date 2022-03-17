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

if($ver.Length -lt 1)
{
    $ver = git describe --tags --abbrev=0
}
$__ = $ver -match '[a-zA-Z]*(\d+)\.(\d+)'
$verMajor = $Matches.1
$verMinor = $Matches.2


# Build release package
if ($Release)
{
    Remove-Item -LiteralPath $ReleasePath -ErrorAction SilentlyContinue
    Remove-Item -Path checksums.md -ErrorAction SilentlyContinue

    Write-Output "
## Checksums
| Architecture | File Name | Checksum |
|---|---|---|" | Out-File -FilePath checksums.md -Encoding utf8 

    Push-Location resources
    goversioninfo -ver-major ${verMajor} -ver-minor ${verMinor} -product-version ${ver} -product-ver-major ${verMajor} -product-ver-minor ${verMinor} -platform-specific
    Pop-Location
}

function PrepareBuildDir ([string] $path, [string] $arch) {
    $buildDir = New-Item -ItemType Directory -Force -Path "$path/$arch"

    Copy-Item -Force .\README.md $buildDir
    Copy-Item -Force .\LICENSE $buildDir
    Copy-Item -Force "resources/resource*.syso" ./

    Copy-Item -Recurse -Force -Path "resources/templates" -Destination $buildDir
    Copy-Item -Force -Path "resources/wix.json" -Destination $buildDir
    Copy-Item -Force -Path "resources/icon/icon.ico" -Destination $buildDir

    return $buildDir
}

function CreateStandaloneZip ([string] $path, [string] $outDir, [string] $arch) {
    Compress-Archive -Force -Path $path\README.md,$path\LICENSE,$path\*.exe -DestinationPath $outDir\$ReleasePath-${ver}_$arch.zip
}

# Build Standalone for each architecture
Foreach ($arch in $Architectures)
{
    $env:GOARCH=$arch

    if ($Release)
    {
        # $buildFlags = "-ldflags -s -w -X main.version=$ver"
        $buildDir = PrepareBuildDir $outDir $arch
        $buildFlags = "-ldflags=""-w -s -H=windowsgui"" -trimpath"
        $binary = "winssh-pageant.exe"
    } else {
        $buildDir = $outDir
        $buildFlags = ""
        $binary = "winssh-pageant-$arch.exe"
    }

    Invoke-Expression ("go build ${buildFlags} -o $buildDir\$binary" )
    if ($LastExitCode -ne 0) { $returnValue = $LastExitCode }

    if ($Release)
    {
        CreateStandaloneZip $buildDir $releaseDir $arch
        
        Push-Location $buildDir
        $msiName = "winssh-pageant-${ver}_${arch}.msi"
        go-msi make --path $buildDir\wix.json --src $buildDir\templates --out $buildDir\tmp --version $ver --arch $arch --msi "${releaseDir}\${msiName}" --keep
        Pop-Location

        $checksum = (Get-FileHash -Algorithm SHA256 -Path "${buildDir}\${binary}").Hash
        Write-Output "| $arch | $binary | ``${checksum}`` |" | Out-File -FilePath checksums.md -Encoding utf8 -Append

        $checksum = (Get-FileHash -Algorithm SHA256 -Path "${releaseDir}\${msiName}").Hash
        Write-Output "| $arch | $msiName | ``$checksum`` |" | Out-File -FilePath checksums.md -Encoding utf8 -Append
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
    Remove-Item -Path .\resource*.syso -Force -ErrorAction SilentlyContinue
    Remove-Item -Path .\resources\resource*.syso -Force -ErrorAction SilentlyContinue
}

exit $returnValue
