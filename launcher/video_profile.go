package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"
)

// Decode without changing ANSI/UTF-8 bytes. UTF-16 is accepted only with a BOM.
func decodeProfile(data []byte) (string, func(string) []byte, error) {
	if bytes.HasPrefix(data, []byte{0xff, 0xfe}) {
		raw := data[2:]
		if len(raw)%2 != 0 {
			return "", nil, fmt.Errorf("Ongeldige UTF-16-configuratie")
		}
		units := make([]uint16, len(raw)/2)
		for i := range units {
			units[i] = binary.LittleEndian.Uint16(raw[i*2:])
		}
		// Avoid silently replacing invalid surrogate pairs in an existing profile.
		for i := 0; i < len(units); i++ {
			u := units[i]
			if u >= 0xd800 && u <= 0xdbff {
				if i+1 >= len(units) || units[i+1] < 0xdc00 || units[i+1] > 0xdfff {
					return "", nil, fmt.Errorf("Ongeldige UTF-16-configuratie")
				}
				i++
			} else if u >= 0xdc00 && u <= 0xdfff {
				return "", nil, fmt.Errorf("Ongeldige UTF-16-configuratie")
			}
		}
		return string(utf16.Decode(units)), func(s string) []byte {
			u := utf16.Encode([]rune(s))
			out := make([]byte, 2+len(u)*2)
			out[0] = 0xff
			out[1] = 0xfe
			for i, v := range u {
				binary.LittleEndian.PutUint16(out[2+i*2:], v)
			}
			return out
		}, nil
	}
	if bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
		return "", nil, fmt.Errorf("Big-endian configuratie wordt niet automatisch gewijzigd")
	}
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		return string(data[3:]), func(s string) []byte { return append([]byte{0xef, 0xbb, 0xbf}, []byte(s)...) }, nil
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return "", nil, fmt.Errorf("Onbekende configuratiecodering")
	}
	return string(data), func(s string) []byte { return []byte(s) }, nil
}

func patchVideoProfile(original []byte, cfg Config) ([]byte, error) {
	args, err := videoArgs(cfg)
	if err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return original, nil
	}
	text, encode, err := decodeProfile(original)
	if err != nil {
		return nil, err
	}
	if strings.Contains(text, "/*") {
		return nil, fmt.Errorf("Videoprofiel bevat blokcommentaar; deze aangepaste opmaak wordt niet automatisch gewijzigd")
	}
	values := map[string]string{}
	order := []string{}
	for i := 0; i < len(args); i += 3 {
		values[args[i+1]] = args[i+2]
		order = append(order, args[i+1])
	}
	seen := map[string]bool{}
	// Only replace a standalone video-setting value. Keep comments, binds, user
	// names and any custom command chains byte-for-byte unchanged.
	re := regexp.MustCompile(`(?im)^([\t ]*set(?:a|u)?[\t ]+)("?(?:r_fullscreen|r_mode|r_vsync|com_maxfps)"?)([\t ]+)("[^"\r\n]*"|[^\s;\r\n]+)([^\r\n]*)`)
	text = re.ReplaceAllStringFunc(text, func(line string) string {
		m := re.FindStringSubmatchIndex(line)
		key := strings.ToLower(strings.Trim(line[m[4]:m[5]], "\""))
		value, ok := values[key]
		if !ok || strings.Contains(line[m[10]:m[11]], ";") {
			return line
		}
		seen[key] = true
		return line[:m[8]] + "\"" + value + "\"" + line[m[9]:]
	})
	newline := "\n"
	if strings.Contains(text, "\r\n") || text == "" {
		newline = "\r\n"
	}
	for _, key := range order {
		if !seen[key] {
			if text != "" && !strings.HasSuffix(text, "\n") {
				text += newline
			}
			text += "seta " + key + " \"" + values[key] + "\"" + newline
		}
	}
	return encode(text), nil
}

func prepareVideoProfile(cfg Config, state string) (string, bool, error) {
	args, err := videoArgs(cfg)
	if err != nil {
		return "", false, err
	}
	if len(args) == 0 {
		return "", false, nil
	}
	candidates := []string{
		filepath.Join(cfg.PlutoDir, "storage", "iw5", "players", "config_mp.cfg"),
		filepath.Join(cfg.PlutoDir, "storage", "iw5", "players2", "config_mp.cfg"),
		filepath.Join(cfg.GameDir, "players2", "config_mp.cfg"),
		filepath.Join(cfg.GameDir, "players", "config_mp.cfg"),
	}
	for _, path := range candidates {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", false, err
		}
		if !info.Mode().IsRegular() || info.Size() > 8*1024*1024 {
			return "", false, fmt.Errorf("Onverwacht videoprofiel: %s", path)
		}
		if _, err := safeTarget(filepath.Dir(path), filepath.Base(path)); err != nil {
			return "", false, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", false, err
		}
		next, err := patchVideoProfile(data, cfg)
		if err != nil {
			return "", false, fmt.Errorf("Videoprofiel %s: %w", path, err)
		}
		if bytes.Equal(data, next) {
			return path, false, nil
		}
		if info.Mode().Perm()&0200 == 0 {
			return "", false, fmt.Errorf("Videoprofiel is alleen-lezen: %s", path)
		}
		hash := sha256.Sum256(data)
		backup, err := safeTarget(filepath.Join(state, "backups"), fmt.Sprintf("video-profile/%x.cfg", hash))
		if err != nil {
			return "", false, err
		}
		saved, err := os.ReadFile(backup)
		if err == nil {
			if !bytes.Equal(saved, data) {
				return "", false, fmt.Errorf("Bestaande videobackup komt niet overeen; niets gewijzigd")
			}
		} else if os.IsNotExist(err) {
			if err = atomicWrite(backup, data); err != nil {
				return "", false, err
			}
		} else {
			return "", false, err
		}
		if err = atomicWrite(path, next); err != nil {
			return "", false, err
		}
		return path, true, nil
	}
	return "", false, fmt.Errorf("Geen bestaande config_mp.cfg gevonden. Start MW3 eenmaal via Plutonium zodat het videoprofiel wordt aangemaakt, of kies Bestaande instelling in VIDEO. Gezocht in: %s", strings.Join(candidates, ", "))
}
