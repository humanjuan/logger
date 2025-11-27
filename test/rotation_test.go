package acacia_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	acacia "github.com/humanjuan/acacia/v2"
)

func readAndCleanDir(t *testing.T, dir string, baseName string) []string {
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Fallo al leer directorio de prueba %s: %v", dir, err)
	}

	var rotatedFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() != baseName {
			rotatedFiles = append(rotatedFiles, file.Name())
		}
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Advertencia: No se pudo limpiar el directorio temporal %s: %v", dir, err)
		}
	})
	return rotatedFiles
}

func TestIntensiveRotation(t *testing.T) {
	const logName = "intensive.log"
	const maxRot = 3
	const logLine = "Log Line to force rotation. ABCDEFGHIJKLMNOPQRSTUVWXYZ\n"

	dir := t.TempDir()
	log, err := acacia.Start(logName, dir, acacia.Level.DEBUG)
	if err != nil {
		t.Fatalf("Fallo Start: %v", err)
	}
	defer log.Close()

	t.Log("--- 1. Test de Rotación por Tamaño ---")

	// Configurar rotación por tamaño: 1 MB y 3 backups
	log.Rotation(1, maxRot)
	big := strings.Repeat("X", 2*1024*1024)
	for rotationCount := 0; rotationCount <= maxRot; rotationCount++ {
		t.Logf("Iniciando ciclo %d de rotación por tamaño.", rotationCount)
		log.Info(big)
		log.Sync()
		rotatedFiles := readAndCleanDir(t, dir, logName)
		if rotationCount == 0 {
			if len(rotatedFiles) == 0 {
				t.Errorf("FAIL: No se detectaron backups tras la primera rotación. Archivos: %v", rotatedFiles)
			}
		}
		if len(rotatedFiles) > maxRot+1 {
			t.Errorf("FAIL: Se excedió el número máximo de backups (%d): %v", maxRot+1, rotatedFiles)
		}
	}

	t.Log("--- 2. Test de Rotación Diaria ---")
	log.DailyRotation(true)

	log.Info("DailyLog: Forzando creación de archivo fechado de hoy.")
	log.Sync()
	today := time.Now().Format("2006-01-02")
	todayName := fmt.Sprintf("%s-%s", strings.TrimSuffix(logName, filepath.Ext(logName)), today)
	expectedToday := todayName + filepath.Ext(logName)
	if _, err := os.Stat(filepath.Join(dir, expectedToday)); os.IsNotExist(err) {
		t.Errorf("FAIL: No se encontró el archivo fechado de hoy: %s", expectedToday)
	}
}

/*
# run
go test -run TestIntensiveRotation -v
*/
