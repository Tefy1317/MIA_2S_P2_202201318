package funciones

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var lastDiskPath string

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

func separarComando(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])
		params := strings.Join(parts[1:], " ")

		params = strings.TrimSpace(params)
		if params == "" {
			params = "Parámetro vacío"
		}

		return command, params
	}
	return "", input
}

func Analyze(linea string) string {
	command, params := separarComando(linea)

	switch command {
	case "mkdisk":
		mensajes := fn_mkdisk(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "rmdisk":
		mensajes := fn_rmdisk(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "fdisk":
		mensajes := fn_fdisk(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "mount":
		mensajes := fn_mount(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "mkfs":
		mensajes := fn_mkfs(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "login":
		mensajes := fn_login(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	case "logout":
		mensajes, err := Logout()
		if err != nil {
			return fmt.Sprintf("--> Comando leído: %s\nError: %v", command, err)
		}
		return fmt.Sprintf("--> Comando leído: %s\n%s", command, mensajes)
	case "mkgrp":
		// Lógica específica para mkgrp
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "rmgrp":
		// Lógica específica para rmgrp
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "mkusr":
		// Lógica específica para mkusr
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "rmusr":
		// Lógica específica para rmusr
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "chgrp":
		// Lógica específica para chgrp
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "mkfile":
		// Lógica específica para mkfile
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "mkdir":
		// Lógica específica para mkdir
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s", command, params)
	case "rep":
		mensajes := GenerateReport(params)
		return fmt.Sprintf("--> Comando leído: %s - Parámetro: %s\n%s", command, params, mensajes)
	default:
		// Para cualquier comando no especificado
		return fmt.Sprintf("--> Comando no reconocido: %s - Parámetro: %s", command, params)
	}
}

func fn_mkdisk(params string) string {

	var output string
	fmt.Println("Ejecutando mkdisk...")

	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	fs.Parse(os.Args[1:])

	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			output += "Error: Parametro no reconocido ->  -" + flagName + "=" + flagValue + "\n"
			return output
		}
	}

	if *size <= 0 {
		output += "Error: el tamaño debe ser mayor a 0\n"
		return output
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		output += "Error: Fit debe ser FF, BF o WF\n"
		return output
	}

	if *unit != "k" && *unit != "m" {
		output += "Error: la unidad debe ser M o K\n"
		return output
	}

	if *path == "" {
		output += "Error: se necesita un path\n"
		return output
	}

	lastDiskPath = *path
	mensajes := Mkdisk(*size, *fit, *unit, *path)
	output += mensajes

	fmt.Println("mkdisk finalizado")

	return mensajes
}

func fn_rmdisk(params string) string {
	var output string

	fmt.Println("Ejecutando rmdisk...")
	output += "-----------------------------\n"
	output += "Ejecutando rmdisk...\n"
	output += "-----------------------------\n"

	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta del disco a eliminar")
	fs.Parse(os.Args[1:])

	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.TrimSpace(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		if flagName == "path" {
			flagValue = strings.ToLower(flagValue)
		}

		switch flagName {
		case "path":
			fs.Set(flagName, flagValue)
		default:
			output += "Error: Parámetro no reconocido -> -" + flagName + "=" + flagValue + "\n"
			return output
		}
	}

	if *path == "" {
		output += "Error: Falta el parámetro obligatorio path.\n"
		return output
	}

	if _, err := os.Stat(*path); os.IsNotExist(err) {
		output += fmt.Sprintf("Error: No existe el disco para ser eliminado.\n")
		return output
	}

	err := os.Remove(*path)
	if err != nil {
		output += fmt.Sprintf("Error al eliminar el disco \n")
		return output
	}

	output += fmt.Sprintf("El disco fue eliminado exitosamente.\n")

	output += "-----------------------------\n"
	output += "rmdisk finalizado :)\n"
	output += "-----------------------------\n"

	fmt.Println("rmdisk finalizado")
	return output
}

func fn_fdisk(params string) string {

	var output string

	fmt.Println("Ejecutando fdisk...")

	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "k", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "wf", "Ajuste")
	delete := fs.String("delete", "", "Eliminar partición")
	add := fs.String("add", "", "Agregar espacio a partición")

	fs.Parse(os.Args[1:])

	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type", "delete", "add":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Parámetro no reconocido -> -" + flagName + "=" + flagValue)
			output += "Error: Parámetro no reconocido -> -" + flagName + "=" + flagValue + "\n"
			return output
		}
	}

	if *size <= 0 {
		fmt.Println("Error: el tamaño debe ser mayor a 0")
		output += "Error: el tamaño debe ser mayor a 0"
		return output
	}

	if *path == "" {
		fmt.Println("Error: path obligatorio")
		output += "Error: path obligatorio"
		return output
	}

	if *name == "" {
		fmt.Println("Error: nombre obligatorio")
		output += "Error: nombre obligatorio"
		return output
	}

	if *fit == "" {
		*fit = "wf"
	}

	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit debe ser 'bf', 'ff', o 'wf'")
		output += "Error: Fit debe ser 'bf', 'ff', o 'wf'"
		return output
	}

	if *unit != "k" && *unit != "m" && *unit != "b" {
		fmt.Println("Error: Unidad debe ser 'k', 'b' o 'm'")
		output += "Error: Unidad debe ser 'k', 'b' o 'm'"
		return output
	}

	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: el tipo de particion debe ser 'p', 'e', o 'l'")
		output += "Error: el tipo de particion debe ser 'p', 'e', o 'l'"
		return output
	}

	mensajes := Fdisk(*size, *path, *name, *unit, *type_, *fit)

	if *delete != "" {
		fmt.Println("Eliminando partición...")
		output += DeleteParticion(*path, *name, *delete)
		return output
	}

	if *add != "" {
		fmt.Println("Agregando espacio a partición...")

		addInt, errConv := strconv.Atoi(*add)
		if errConv != nil {
			output += fmt.Sprintf("Error: el valor de 'add' no es un número válido: %v\n", errConv)
			return output
		}

		outputAdd, err := AddParticion(*path, *name, addInt, *unit)
		if err != nil {
			output += fmt.Sprintf("Error al agregar espacio a la partición: %v\n", err)
		} else {
			output += outputAdd
		}
	}

	output += mensajes

	fmt.Println("fdisk finalizado")
	return output
}

func fn_mount(params string) string {
	var output string
	fmt.Println("Ejecutando mount...")

	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Nombre son obligatorios")
		output += "Error: Path y Nombre son obligatorios"
		return output
	}

	lowercaseName := strings.ToLower(*name)
	mensajes := Mount(*path, lowercaseName)
	output += mensajes
	fmt.Println("mount finalizado")
	return output
}

func fn_mkfs(input string) string {
	var output string
	fmt.Println("Ejecutando mkfs...")
	fs := flag.NewFlagSet("mkfs", flag.ExitOnError)
	id := fs.String("id", "", "Id")
	type_ := fs.String("type", "full", "Tipo")
	fs_ := fs.String("fs", "2fs", "Sistema de archivos")

	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")
		fs.Set(flagName, flagValue)
	}

	*id = strings.ToLower(*id)
	*type_ = strings.ToLower(*type_)
	*fs_ = strings.ToLower(*fs_)

	if *id == "" {
		output += "Error: id es un parámetro obligatorio.\n"
		return output
	}

	if *fs_ != "2fs" && *fs_ != "3fs" {
		output += "Error: El sistema de archivos debe ser '2fs' (EXT2) o '3fs' (EXT3).\n"
		return output
	}

	mensajes := Mkfs(*id, *type_, *fs_)
	output += mensajes
	fmt.Println("mkfs finalizado")
	return output
}

func fn_login(input string) string {

	var output string
	fmt.Println("Ejecutando login...")

	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id de la partición")

	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.Trim(match[2], "\"")
		fs.Set(flagName, flagValue)
	}

	*id = strings.ToLower(*id)

	if *user == "" || *pass == "" || *id == "" {
		output += "Error: Los campos user, pass e id son obligatorios.\n"
		return output
	}

	mensajes, err := Login(*user, *pass, *id)
	if err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output
	}

	output += "Inicio de sesión exitoso"

	output += mensajes
	return output
}

//REPORTES

func GenerateReport(params string) string {
	var output string

	fmt.Println("Ejecutando rep...")

	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	id := fs.String("id", "", "ID de la partición")
	path := fs.String("path", "", "Ruta del archivo del disco")
	name := fs.String("name", "", "Nombre del reporte a generar")

	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path", "name", "id":
			fs.Set(flagName, flagValue)
		default:
			output += "Error: Parametro no reconocido ->  -" + flagName + "=" + flagValue + "\n"
			return output
		}
	}

	if *path == "" || *name == "" || *id == "" {
		output += "Error: Parámetros necesarios no proporcionados (path, name, id)\n"
		return output
	}

	particion, err := BuscarParticionMontada(*id)
	if err != nil {
		output += err.Error() + "\n"
		output += "Particion no montada con el ID: " + *id + "\n"

		return output
	} else {
		output += fmt.Sprintf("Partición encontrada: %s, en el disco: %s\n", particion.Name, particion.Path)
	}

	file, err := OpenFile(particion.Path)
	if err != nil {
		output += fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s\n", *path)
		return output
	}
	defer file.Close()

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += "Error: No se pudo leer el MBR desde el archivo\n"
		return output
	}

	var ebrs []EBR
	for i := 0; i < 4; i++ {
		if string(TempMBR.Particion[i].Type[:]) == "e" {
			ebrPos := TempMBR.Particion[i].Start
			for {
				var ebr EBR
				err := ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					output += fmt.Sprintf("Error al leer EBR: %v\n", err)
					break
				}
				ebrs = append(ebrs, ebr)

				if ebr.PartNext == -1 {
					break
				}
				ebrPos = ebr.PartNext
			}
		}
	}

	switch *name {
	case "mbr":
		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = GenerateMBRReport(TempMBR, ebrs, dotFilePath, file)
		if err != nil {
			output += fmt.Sprintf("Error al generar el reporte MBR: %v\n", err)
			return output
		}
		switch filepath.Ext(*path) {
		case ".pdf":
			err = ConvertDotToPDF(dotFilePath, *path)
			if err != nil {
				output += fmt.Sprintf("Error al convertir a PDF: %v\n", err)
				return output
			}
		case ".jpg":
			err = ConvertDotToJPG(dotFilePath, *path)
			if err != nil {
				output += fmt.Sprintf("Error al convertir a JPG: %v\n", err)
				return output
			}
		default:
			output += "Error: Formato de salida no soportado. Use .pdf o .jpg\n"
			return output
		}
	case "disk":
		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = GenerateDiskReport(TempMBR, ebrs, dotFilePath)
		if err != nil {
			output += fmt.Sprintf("Error al generar el reporte DISK: %v\n", err)
			return output
		}
		switch filepath.Ext(*path) {
		case ".pdf":
			err = ConvertDotToPDF(dotFilePath, *path)
			if err != nil {
				output += fmt.Sprintf("Error al convertir a PDF: %v\n", err)
				return output
			}
		case ".jpg":
			err = ConvertDotToJPG(dotFilePath, *path)
			if err != nil {
				output += fmt.Sprintf("Error al convertir a JPG: %v\n", err)
				return output
			}
		default:
			output += "Error: Formato de salida no soportado. Use .pdf o .jpg\n"
			return output
		}

	default:
		output += "Error: Tipo de reporte no reconocido.\n"
		return output
	}

	output += "Reporte generado exitosamente\n"
	return output
}

func BuscarParticionMontada(id string) (*particionMount, error) {
	for _, particiones := range particionesMontadas {
		for _, particion := range particiones {
			fmt.Println("Comparando ID: " + particion.ID + " con " + id)
			if strings.ToLower(particion.ID) == id {
				return &particion, nil
			}
		}
	}
	return nil, fmt.Errorf("Error: No se encontró una partición montada con el ID: %s", id)
}

func ConvertDotToJPG(dotFilePath, jpgFilePath string) error {
	cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", jpgFilePath)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error al convertir .dot a JPG: %v", err)
	}

	fmt.Printf("Archivo JPG generado exitosamente: %s\n", jpgFilePath)
	return nil
}
