# Close DevTools startup dialogs (e.g. the version-update prompt) but keep the
# project window. ASCII-only (PowerShell 5.1 ANSI decoding).
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class Win32Close {
  [DllImport("user32.dll")] public static extern bool PostMessage(IntPtr hWnd, uint msg, IntPtr wParam, IntPtr lParam);
}
"@

$keepPatterns = @('Travel_Guangdong', 'WeChat Web Devtools')
for ($i = 0; $i -lt 3; $i++) {
  $procs = Get-Process -Name wechatdevtools -ErrorAction SilentlyContinue |
    Where-Object { $_.MainWindowHandle -ne 0 }
  if (-not $procs) { break }
  $seenProjectWindow = $false
  foreach ($p in $procs) {
    $title = $p.MainWindowTitle
    $keep = $false
    foreach ($pat in $keepPatterns) {
      if ($title -like ('*' + $pat + '*')) { $keep = $true }
    }
    if ($keep) {
      $seenProjectWindow = $true
      Write-Output ("kept project window: pid=" + $p.Id)
      continue
    }
    [Win32Close]::PostMessage($p.MainWindowHandle, 0x0010, [IntPtr]::Zero, [IntPtr]::Zero) | Out-Null
    Write-Output ("WM_CLOSE sent to dialog pid=" + $p.Id)
  }
  if ($seenProjectWindow) { break }
  Start-Sleep -Seconds 2
}

Start-Sleep -Seconds 2
Get-Process -Name wechatdevtools -ErrorAction SilentlyContinue |
  Where-Object { $_.MainWindowHandle -ne 0 } |
  ForEach-Object { Write-Output ("window now: [" + $_.MainWindowTitle + "] pid=" + $_.Id) }
