# DMXAPI 配置工具一键安装脚本 (Windows PowerShell)
# 用法: iwr -useb https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo        = 'dmxapi/opencode_dmxapi'
$ReleasesApi = "https://cnb.cool/$Repo/-/releases"
$BinPrefix   = 'opencode-dmxapi'

function Write-Info($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-Warn($msg) { Write-Host "!! $msg"  -ForegroundColor Yellow }
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

    $tags = @($releases | ForEach-Object { $_.tag_name } | Where-Object { $_ })
    $stable = @($tags | Where-Object { $_ -notmatch '-' })
    if ($stable.Count -gt 0) {
        $parsed = @()
        foreach ($t in $stable) {
            $vstr = $t.TrimStart('v')
            $v = $null
            if ([Version]::TryParse($vstr, [ref]$v)) {
                $parsed += [PSCustomObject]@{ Tag = $t; Version = $v }
            }
        }
        if ($parsed.Count -gt 0) {
            return ($parsed | Sort-Object Version -Descending | Select-Object -First 1).Tag
        }
        return ($stable | Sort-Object -Descending)[0]
    }
    Write-Warn "未找到稳定版本，使用最新预发布版本"
    $parsedPre = @()
    foreach ($t in $tags) {
        $vstr = $t.TrimStart('v')
        # 去掉 "-beta.1" / "+sha" 等后缀，只保留前导数字段
        $vstr = ($vstr -split '[-+]')[0]
        $v = $null
        if ([Version]::TryParse($vstr, [ref]$v)) {
            $parsedPre += [PSCustomObject]@{ Tag = $t; Version = $v }
        }
    }
    if ($parsedPre.Count -gt 0) {
        return ($parsedPre | Sort-Object Version -Descending | Select-Object -First 1).Tag
    }
    return ($tags | Sort-Object -Descending)[0]
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
    Write-Info "临时目录: $tmpDir（脚本退出时自动删除）"
    Write-Info "下载 $asset ..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $binPath -UseBasicParsing
    } catch {
        Write-Err "下载失败: $url"
        Write-Err "请确认 release 资产已发布，或访问 $ReleasesApi 手动下载"
        exit 1
    }

    # 可选 SHA256 校验：老 release 可能不带 .sha256，下载失败则跳过
    $shaUrl  = "$url.sha256"
    $shaPath = "$binPath.sha256"
    $shaOk = $false
    try {
        Invoke-WebRequest -Uri $shaUrl -OutFile $shaPath -UseBasicParsing -ErrorAction Stop
        $shaOk = $true
    } catch {
        Write-Warn '未找到 SHA256 校验文件，跳过完整性校验（向后兼容旧 release）'
    }
    if ($shaOk) {
        $expected = (Get-Content $shaPath -Raw).Trim().Split()[0]
        $actual   = (Get-FileHash -Path $binPath -Algorithm SHA256).Hash.ToLower()
        if ($expected.ToLower() -ne $actual) {
            Write-Err "SHA256 校验失败！期望 $expected，实际 $actual"
            exit 1
        }
        Write-Info 'SHA256 校验通过'
    }

    Write-Info '启动配置...'
    Write-Host ''
    # 使用 Start-Process -NoNewWindow 让子进程接管当前控制台 stdin/stdout，
    # 避免 iex pipeline 占用 stdin 导致 huh 交互输入立刻 EOF
    $proc = Start-Process -FilePath $binPath -NoNewWindow -Wait -PassThru
    if ($proc.ExitCode -ne 0) {
        exit $proc.ExitCode
    }
} finally {
    Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
}
