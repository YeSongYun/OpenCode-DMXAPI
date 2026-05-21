# DMXAPI 配置工具一键安装脚本 (Windows PowerShell)
# 用法: iwr -useb https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo        = 'dmxapi/opencode_dmxapi'
$ReleasesApi = "https://cnb.cool/$Repo/-/releases"
$BinPrefix   = 'opencode-dmxapi'

function Write-Info($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-Err ($msg) { Write-Host "X $msg"   -ForegroundColor Red }

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { return 'amd64' }
        'ARM64' { return 'arm64' }
        'x86'   {
            if ([Environment]::Is64BitOperatingSystem) { return 'amd64' }
            Write-Err '不支持 32 位 Windows'; exit 1
        }
        default {
            Write-Err "不支持的架构: $env:PROCESSOR_ARCHITECTURE"; exit 1
        }
    }
}

function Get-LatestTag {
    try {
        $resp = Invoke-WebRequest -Uri $ReleasesApi -Headers @{ 'Accept' = 'application/json' } -UseBasicParsing
    } catch {
        Write-Err "无法访问 $ReleasesApi : $($_.Exception.Message)"
        exit 1
    }
    try {
        $releases = $resp.Content | ConvertFrom-Json
    } catch {
        Write-Err '无法解析 releases JSON'
        exit 1
    }
    if (-not $releases -or $releases.Count -eq 0) {
        Write-Err '未找到任何 release'
        exit 1
    }
    return $releases[0].tag_name
}

Write-Info '检测平台...'
$arch = Get-Arch
$platform = "windows-$arch"
Write-Info "平台: $platform"

Write-Info '获取最新版本...'
$tag = Get-LatestTag
Write-Info "版本: $tag"

$asset = "$BinPrefix-$tag-$platform.exe"
$url   = "https://cnb.cool/$Repo/-/releases/download/$tag/$asset"

$tmpDir = Join-Path $env:TEMP ("dmxapi-" + [Guid]::NewGuid().ToString('N').Substring(0,8))
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
$binPath = Join-Path $tmpDir $asset

try {
    Write-Info "下载 $asset ..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $binPath -UseBasicParsing
    } catch {
        Write-Err "下载失败: $url"
        Write-Err "请确认 release 资产已发布，或访问 $ReleasesApi 手动下载"
        exit 1
    }

    Write-Info '启动配置...'
    Write-Host ''
    & $binPath
} finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
