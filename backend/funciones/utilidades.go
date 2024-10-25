package funciones

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func CreateFile(name string) error {
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Err CreateFile dir==", err)
		return err
	}

	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Err CreateFile create==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

func FillWithZeros(file *os.File, start int32, size int32) error {

	file.Seek(int64(start), 0)
	buffer := make([]byte, size)

	_, err := file.Write(buffer)
	if err != nil {
		fmt.Println("Error al llenar el espacio con ceros:", err)
		return err
	}

	fmt.Println("Espacio llenado con ceros desde el byte", start, "por", size, "bytes.")
	return nil
}

func VerifyZeros(file *os.File, start int32, size int32) {
	zeros := make([]byte, size)
	_, err := file.ReadAt(zeros, int64(start))
	if err != nil {
		fmt.Println("Error al leer la sección eliminada:", err)
		return
	}

	isZeroFilled := true
	for _, b := range zeros {
		if b != 0 {
			isZeroFilled = false
			break
		}
	}

	if isZeroFilled {
		fmt.Println("La partición eliminada está completamente llena de ceros.")
	} else {
		fmt.Println("Advertencia: La partición eliminada no está completamente llena de ceros.")
	}
}
