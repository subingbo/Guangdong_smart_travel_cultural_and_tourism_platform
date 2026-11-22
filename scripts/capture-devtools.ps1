# Capture the WeChat DevTools main window to a PNG using PrintWindow, which
# renders the window surface directly so it works even when the window is
# occluded or the desktop session is locked. Falls back to screen capture.
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File scripts/capture-devtools.ps1 -OutFile docs\screenshots\01-home.png
param(
  [Parameter(Mandatory = $true)][string]$OutFile,
  [string]$ProcessName = "wechatdevtools",
  [switch]$ScreenCopy
)

Add-Type -AssemblyName System.Drawing
Add-Type @"
using System;
using System.Drawing;
using System.Runtime.InteropServices;
public class Win32Shot {
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
  [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr hWnd, IntPtr hdc, uint flags);
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
  [DllImport("dwmapi.dll")] public static extern int DwmGetWindowAttribute(IntPtr hwnd, int attr, out RECT rect, int size);
  public struct RECT { public int Left, Top, Right, Bottom; }
  public const int DWMWA_EXTENDED_FRAME_BOUNDS = 9;
  public const uint PW_RENDERFULLCONTENT = 2;
}
"@

$proc = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue |
  Where-Object {
    $_.MainWindowHandle -ne 0 -and
    ($_.MainWindowTitle -like "*Travel_Guangdong*" -or $_.MainWindowTitle -like "*Devtools*" -or $_.MainWindowTitle -like "*DevTools*")
  } |
  Select-Object -First 1
if (-not $proc) {
  $proc = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue |
    Where-Object { $_.MainWindowHandle -ne 0 } |
    Select-Object -First 1
}
if (-not $proc) { throw "no visible $ProcessName window found" }

$hwnd = $proc.MainWindowHandle
[Win32Shot]::ShowWindow($hwnd, 9) | Out-Null   # SW_RESTORE
Start-Sleep -Milliseconds 500

# Prefer the DWM frame bounds (excludes invisible resize borders).
$rect = New-Object Win32Shot+RECT
$dwm = [Win32Shot]::DwmGetWindowAttribute($hwnd, [Win32Shot]::DWMWA_EXTENDED_FRAME_BOUNDS, [ref]$rect, 16)
if ($dwm -ne 0 -or ($rect.Right - $rect.Left) -le 0) {
  [Win32Shot]::GetWindowRect($hwnd, [ref]$rect) | Out-Null
}
$width = $rect.Right - $rect.Left
$height = $rect.Bottom - $rect.Top
if ($width -le 0 -or $height -le 0) { throw "invalid window rect ${width}x${height}" }

$bmp = New-Object System.Drawing.Bitmap $width, $height
$graphics = [System.Drawing.Graphics]::FromImage($bmp)
$hdc = $graphics.GetHdc()
if ($ScreenCopy) {
  $ok = $false
} else {
  $ok = [Win32Shot]::PrintWindow($hwnd, $hdc, [Win32Shot]::PW_RENDERFULLCONTENT)
}
$graphics.ReleaseHdc($hdc)

if (-not $ok) {
  # Fallback: composite from the screen (only useful when the window is visible).
  $graphics.CopyFromScreen($rect.Left, $rect.Top, 0, 0, $bmp.Size)
} elseif (-not $ScreenCopy) {
  # PrintWindow may leave the alpha channel transparent; flatten to white.
  $flat = New-Object System.Drawing.Bitmap $width, $height
  $fg = [System.Drawing.Graphics]::FromImage($flat)
  $fg.Clear([System.Drawing.Color]::White)
  $fg.DrawImage($bmp, 0, 0, $width, $height)
  $fg.Dispose()
  $bmp.Dispose()
  $bmp = $flat
}

$full = [System.IO.Path]::GetFullPath($OutFile)
$dir = [System.IO.Path]::GetDirectoryName($full)
if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
$bmp.Save($full, [System.Drawing.Imaging.ImageFormat]::Png)
$graphics.Dispose()
$bmp.Dispose()
Write-Output ("saved " + $full + " (" + $width + "x" + $height + ", printwindow=" + $ok + ")")
