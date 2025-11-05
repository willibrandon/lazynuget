# Terminal Manual Testing Checklist

This checklist covers manual testing for terminal graceful degradation (User Story 3).
Automated tests cannot fully verify terminal behavior, so manual testing is required.

See: T072, FR-015, FR-016, research.md

## Prerequisites

- [ ] LazyNuGet binary built with `make build-dev` or `make build`
- [ ] Multiple terminal applications available for testing
- [ ] Ability to resize terminal windows
- [ ] Ability to set environment variables

## Terminal Emulators to Test

Test on at least 2-3 of the following terminals per platform:

### macOS
- [ ] iTerm2 (recommended - excellent ANSI support)
- [ ] Terminal.app (built-in)
- [ ] Alacritty (GPU-accelerated, cross-platform)
- [ ] Kitty (GPU-accelerated)

### Linux
- [ ] Alacritty (GPU-accelerated, cross-platform)
- [ ] GNOME Terminal
- [ ] Konsole (KDE)
- [ ] xterm (minimal, good for testing degradation)

### Windows
- [ ] Windows Terminal (recommended - modern, full ANSI support)
- [ ] cmd.exe (legacy, limited color support)
- [ ] PowerShell (good ANSI support)
- [ ] ConEmu / Cmder

## Test Cases

### TC1: Color Depth Detection

**Objective**: Verify color support is detected correctly for different terminal types.

**Steps**:
1. [X] Launch LazyNuGet with `--log-level debug`
2. [X] Check startup logs for "Terminal capabilities: ColorDepth=..."
3. [X] Verify color depth matches terminal's actual capabilities:
   - iTerm2 / Windows Terminal / Alacritty: Should detect `truecolor` (24-bit) or `extended256`
   - Terminal.app (older): May detect `basic16` or `extended256`
   - cmd.exe: May detect `basic16`
   - xterm (without COLORTERM): May detect `basic16`

**Expected**: Correct color depth detected, logged at DEBUG level

**Pass/Fail**: ✅ PASS (iTerm2 on macOS detected truecolor correctly)

### TC2: NO_COLOR Environment Variable

**Objective**: Verify NO_COLOR environment variable disables all colors.

**Steps**:
1. [ ] Set environment variable: `export NO_COLOR=1` (Unix) or `set NO_COLOR=1` (Windows)
2. [ ] Launch LazyNuGet with `--log-level debug`
3. [ ] Check startup logs for "ColorDepth=none"
4. [ ] Verify TUI renders without any colors (grayscale/monochrome)
5. [ ] Unset NO_COLOR and restart - verify colors are restored

**Expected**: ColorNone detected, TUI works without colors

**Pass/Fail**: _______________

### TC3: TERM=dumb Degradation

**Objective**: Verify graceful degradation with dumb terminal.

**Steps**:
1. [ ] Set environment variable: `export TERM=dumb` (Unix)
2. [ ] Launch LazyNuGet
3. [ ] Verify app enters non-interactive mode automatically
4. [ ] Check logs for "Run mode determined: non-interactive"
5. [ ] Restore TERM and verify interactive mode works

**Expected**: Non-interactive mode, ColorNone, no TUI rendering

**Pass/Fail**: _______________

### TC4: Unicode Support Detection

**Objective**: Verify Unicode detection works with different locales.

**Steps**:
1. [ ] With UTF-8 locale (`LANG=en_US.UTF-8`):
   - Launch LazyNuGet with `--log-level debug`
   - Verify "Unicode=true" in startup logs
2. [ ] With C locale (`LANG=C`):
   - Launch LazyNuget
   - Verify "Unicode=false" in startup logs
3. [ ] Restore original locale

**Expected**: Unicode detection matches locale settings

**Pass/Fail**: _______________

### TC5: Small Terminal Dimensions

**Objective**: Verify clamping and warning for small terminals.

**Steps**:
1. [ ] Resize terminal to very small dimensions (e.g., 30x8)
2. [ ] Launch LazyNuGet with `--log-level debug`
3. [ ] Check logs for dimension warning:
   - "Terminal dimensions...are below recommended minimum 40x10"
   - "Dimensions have been clamped to safe values"
4. [ ] Verify TUI renders at clamped minimum (40x10)
5. [ ] Resize to normal size (e.g., 80x24) and verify warning disappears

**Expected**: Warning logged, dimensions clamped to 40x10 minimum

**Pass/Fail**: _______________

### TC6: Large Terminal Dimensions

**Objective**: Verify clamping for very large terminals.

**Steps**:
1. [ ] Maximize terminal to very large size (e.g., 200+ columns)
2. [ ] Launch LazyNuGet with `--log-level debug`
3. [ ] Verify dimensions are clamped to maximum (500x200)
4. [ ] Check that TUI renders correctly without overflow

**Expected**: Dimensions clamped to safe maximum, TUI renders correctly

**Pass/Fail**: _______________

### TC7: Terminal Resize Events (Unix)

**Objective**: Verify SIGWINCH signal handling on Unix systems.

**Platform**: macOS, Linux only

**Steps**:
1. [ ] Launch LazyNuGet in a terminal
2. [ ] Resize terminal window while app is running
3. [ ] Verify TUI redraws and adapts to new dimensions
4. [ ] Resize multiple times rapidly
5. [ ] Verify no crashes or rendering artifacts

**Expected**: TUI responds smoothly to resize events

**Pass/Fail**: _______________

### TC8: Terminal Resize Events (Windows)

**Objective**: Verify console event polling on Windows.

**Platform**: Windows only

**Steps**:
1. [ ] Launch LazyNuGet in Windows Terminal or PowerShell
2. [ ] Resize terminal window while app is running
3. [ ] Verify TUI redraws within ~500ms (polling interval)
4. [ ] Resize multiple times rapidly
5. [ ] Verify no crashes or rendering artifacts

**Expected**: TUI responds to resize events (may have slight delay due to polling)

**Pass/Fail**: _______________

### TC9: CI Environment Detection

**Objective**: Verify automatic non-interactive mode in CI.

**Steps**:
1. [ ] Set CI environment variable: `export CI=true` (Unix) or `set CI=1` (Windows)
2. [ ] Launch LazyNuGet
3. [ ] Verify app enters non-interactive mode automatically
4. [ ] Check logs for "Run mode determined: non-interactive"
5. [ ] Unset CI and verify interactive mode works

**Expected**: Non-interactive mode automatically enabled in CI

**Pass/Fail**: _______________

### TC10: TTY Detection

**Objective**: Verify TTY detection works correctly.

**Steps**:
1. [ ] Run interactively in terminal: `./lazynuget --log-level debug`
   - Verify "TTY=true" in logs (if terminal supports it)
2. [ ] Run with piped output: `./lazynuget --version 2>&1 | cat`
   - Should complete without trying to render TUI
3. [ ] Run with redirected output: `./lazynuget --version > output.txt 2>&1`
   - Should complete without TUI

**Expected**: TTY correctly detected, non-interactive when piped/redirected

**Pass/Fail**: _______________

## Platform-Specific Notes

### macOS
- iTerm2 has best ANSI support - use as reference
- Terminal.app may have slightly limited color support on older versions
- Unicode should work by default (UTF-8 locale)

### Linux
- Alacritty has excellent performance and full truecolor support
- xterm is useful for testing minimal capabilities
- Ensure locale is set correctly for Unicode (`LANG=en_US.UTF-8`)

### Windows
- Windows Terminal is recommended - full ANSI/truecolor support
- cmd.exe has limited color support (16 colors max)
- PowerShell has good ANSI support on Windows 10+
- Resize events use polling (500ms interval) instead of signals

## Testing Summary

**Date Tested**: _______________
**Tester**: _______________
**Platform**: _______________
**Terminal**: _______________

**Total Test Cases**: 10
**Passed**: _______________
**Failed**: _______________

**Issues Found**:
1. _______________________________________________
2. _______________________________________________
3. _______________________________________________

**Notes**:
_____________________________________________________
_____________________________________________________
_____________________________________________________
