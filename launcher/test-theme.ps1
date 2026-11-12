$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing
Add-Type @'
using System;
using System.Runtime.InteropServices;
public static class ThemeNative {
 [DllImport("user32.dll")] public static extern bool PrintWindow(IntPtr h, IntPtr dc, uint flags);
 [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
 [DllImport("user32.dll")] public static extern IntPtr SendMessage(IntPtr h, uint m, IntPtr w, IntPtr l);
 [DllImport("user32.dll")] public static extern IntPtr GetDlgItem(IntPtr h, int id);
 [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetWindowText(IntPtr h, System.Text.StringBuilder value, int size);
 [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr h);
 public struct RECT { public int Left,Top,Right,Bottom; }
}
'@
$folder = Join-Path $env:RUNNER_TEMP 'theme-test'
Expand-Archive (Join-Path $env:RUNNER_TEMP 'SurvivalReimagined-Windows-v0.5.1.zip') $folder
$process = Start-Process (Join-Path $folder 'SurvivalReimagined.exe') -PassThru
try {
 for ($i=0; $i -lt 20; $i++) {
  Start-Sleep -Milliseconds 500
  $process.Refresh()
  if ($process.HasExited) { throw 'Launcher exited before showing its window' }
  if ($process.MainWindowHandle -ne [IntPtr]::Zero) { break }
 }
 $h = $process.MainWindowHandle
 if ($h -eq [IntPtr]::Zero) { throw 'No launcher window found' }
 function Capture([string]$name) {
  $rect = New-Object ThemeNative+RECT
  [void][ThemeNative]::GetWindowRect($h,[ref]$rect)
  $bmp = New-Object Drawing.Bitmap ($rect.Right-$rect.Left),($rect.Bottom-$rect.Top)
  $graphics = [Drawing.Graphics]::FromImage($bmp)
  $dc=$graphics.GetHdc()
  try { if (![ThemeNative]::PrintWindow($h,$dc,2)) { throw 'PrintWindow failed' } }
  finally { $graphics.ReleaseHdc($dc) }
  $colors = New-Object 'System.Collections.Generic.HashSet[int]'
  for ($x=20;$x -lt $bmp.Width;$x+=7) { for ($y=40;$y -lt $bmp.Height;$y+=7) { [void]$colors.Add($bmp.GetPixel($x,$y).ToArgb()) } }
  if ($colors.Count -lt 12) { throw 'Screenshot appears empty' }
  $bmp.Save((Join-Path $env:RUNNER_TEMP $name),[Drawing.Imaging.ImageFormat]::Png)
  $graphics.Dispose(); $bmp.Dispose()
 }
 [void][ThemeNative]::SendMessage($h,0x111,[IntPtr]106,[IntPtr]::Zero)
 Start-Sleep -Milliseconds 500
 if (![ThemeNative]::IsWindowVisible([ThemeNative]::GetDlgItem($h,104))) { throw 'Start button hidden on play page' }
 Capture 'launcher-v0.5.1-spelen.png'
 [void][ThemeNative]::SendMessage($h,0x111,[IntPtr]107,[IntPtr]::Zero)
 Start-Sleep -Milliseconds 500
 if (![ThemeNative]::IsWindowVisible([ThemeNative]::GetDlgItem($h,101))) { throw 'Browse button hidden on settings page' }
 if ([ThemeNative]::IsWindowVisible([ThemeNative]::GetDlgItem($h,104))) { throw 'Start button overlaps settings page' }
 $nameControl = [ThemeNative]::GetDlgItem($h,108)
 if (![ThemeNative]::IsWindowVisible($nameControl)) { throw 'LAN name field is hidden' }
 $value = New-Object Text.StringBuilder 64
 [void][ThemeNative]::GetWindowText($nameControl,$value,64)
 if ($value.ToString() -ne 'Survivor') { throw 'Legacy configuration has no default LAN name' }
 Capture 'launcher-v0.5.1-instellingen.png'
 [void][ThemeNative]::SendMessage($h,0x111,[IntPtr]109,[IntPtr]::Zero)
 Start-Sleep -Milliseconds 500
 $display=[ThemeNative]::GetDlgItem($h,110)
 $resolution=[ThemeNative]::GetDlgItem($h,111)
 if (![ThemeNative]::IsWindowVisible($display)) { throw 'Video page not visible' }
 if ([ThemeNative]::IsWindowVisible($nameControl)) { throw 'Settings overlap video page' }
 if ([ThemeNative]::SendMessage($display,0x147,[IntPtr]::Zero,[IntPtr]::Zero).ToInt64() -ne 0) { throw 'Video override enabled by default' }
 [void][ThemeNative]::SendMessage($h,0x111,[IntPtr]114,[IntPtr]::Zero)
 if ([ThemeNative]::SendMessage($display,0x147,[IntPtr]::Zero,[IntPtr]::Zero).ToInt64() -ne 1) { throw 'Screenshot preset did not select windowed mode' }
 if ([ThemeNative]::SendMessage($resolution,0x147,[IntPtr]::Zero,[IntPtr]::Zero).ToInt64() -ne 1) { throw 'Screenshot preset did not select 1280x720' }
 Capture 'launcher-v0.5.1-video.png'
 [void][ThemeNative]::SendMessage($h,0x111,[IntPtr]106,[IntPtr]::Zero)
 if ([ThemeNative]::IsWindowVisible($display)) { throw 'Video controls overlap play page' }
 [void][ThemeNative]::SendMessage($h,0x10,[IntPtr]::Zero,[IntPtr]::Zero)
 if (!$process.WaitForExit(5000)) { throw 'Launcher did not close normally' }
 if ($process.ExitCode -ne 0) { throw 'Launcher returned an error' }
} finally { if (!$process.HasExited) { $process.Kill() } }
Write-Host 'Windows launch, page navigation, screenshots and clean shutdown passed.'
