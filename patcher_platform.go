// patcher_platform.go
package unrevealed

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
)

func (p *Patcher) initPlatform() {
	switch runtime.GOOS {
	case "windows":
		p.platform = platformWin32
		if runtime.GOARCH == "amd64" && p.MajorVersion > legacyMaxVersion {
			p.platform = platformWin64
		}
		p.exeName = exeNameWindows
	case "darwin":
		p.platform = p.darwinPlatform()
		p.exeName = exeNameUnix
	default:
		p.platform = platformLinux64
		p.exeName = exeNameUnix
	}
}

func (p *Patcher) darwinPlatform() string {
	if p.MajorVersion <= legacyMaxVersion {
		return platformMac64
	}
	if runtime.GOARCH == "arm64" {
		return platformMacArm64
	}
	return platformMacX64
}

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return windowsDataDir(home)
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", dataDirName)
	default:
		return filepath.Join(home, ".local", "share", dataDirName)
	}
}

func windowsDataDir(home string) string {
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		return filepath.Join(appdata, dataDirName)
	}
	return filepath.Join(home, "AppData", "Roaming", dataDirName)
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
