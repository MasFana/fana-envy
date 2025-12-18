$ErrorActionPreference = "Stop"
$Version = "1.0.0"
$BinaryName = "envy"
$SourceDir = "./cmd/fana-envy"
$DistDir = "./dist"
$ReleaseDir = "./release"

# Clean up
if (Test-Path $DistDir) { Remove-Item -Recurse -Force $DistDir }
if (Test-Path $ReleaseDir) { Remove-Item -Recurse -Force $ReleaseDir }
New-Item -ItemType Directory -Force -Path $ReleaseDir | Out-Null

$Platforms = @(
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" },
    @{ OS = "linux";   Arch = "amd64"; Ext = "" },
    @{ OS = "darwin";  Arch = "amd64"; Ext = "" },
    @{ OS = "darwin";  Arch = "arm64"; Ext = "" }
)

Write-Host "🚧 Starting Build Process v$Version..." -ForegroundColor Cyan

foreach ($P in $Platforms) {
    $OS = $P.OS
    $Arch = $P.Arch
    $Ext = $P.Ext
    $TargetName = "${BinaryName}_${OS}_${Arch}"
    $OutputDir = "$DistDir/$TargetName"
    
    Write-Host "   🔨 Building for $OS/$Arch..." -NoNewline
    
    $Env:GOOS = $OS
    $Env:GOARCH = $Arch
    
    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
    
    go build -trimpath -ldflags "-s -w" -o "$OutputDir/$BinaryName$Ext" $SourceDir
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host " [OK]" -ForegroundColor Green
        
        # Copy README and license if exists
        Copy-Item "README.md" -Destination $OutputDir
        if (Test-Path "LICENSE") { Copy-Item "LICENSE" -Destination $OutputDir }
        
        # Archive
        $ArchiveName = "$ReleaseDir/$TargetName.zip"
        if ($OS -eq "linux") { $ArchiveName = "$ReleaseDir/$TargetName.tar.gz" }
        
        Write-Host "      📦 Archiving to $ArchiveName..." -ForegroundColor DarkGray
        Compress-Archive -Path "$OutputDir/*" -DestinationPath $ArchiveName -Force
    } else {
        Write-Host " [FAILED]" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n✅ Build Complete! Artifacts are in $ReleaseDir" -ForegroundColor Green
