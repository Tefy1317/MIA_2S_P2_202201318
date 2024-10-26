package funciones

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func Mkdisk(size int, fit string, unit string, path string) string {
	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando mkdisk...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("Tamaño: %d\n", size)
	output += fmt.Sprintf("Fit: %s\n", fit)
	output += fmt.Sprintf("Unidad: %s\n", unit)
	output += fmt.Sprintf("Path: %s\n", path)

	if fit != "bf" && fit != "wf" && fit != "ff" {
		output += fmt.Sprintf("Error: Fit debe ser FF, BF o WF\n")
		return output
	}

	if size <= 0 {
		output += fmt.Sprintf("Error: Size debe ser mayor a 0\n")
		return output
	}

	if unit != "k" && unit != "m" {
		output += fmt.Sprintf("Error: Las unidades válidas son K o M\n")
		return output
	}

	err := CreateFile(path)
	if err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output
	}

	if unit == "k" {
		size = size * 1024
	} else {
		size = size * 1024 * 1024
	}

	file, err := OpenFile(path)
	if err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output
	}

	blockSize := 1024 * 1024
	zeroBlock := make([]byte, blockSize)

	remainingSize := size

	for remainingSize > 0 {
		if remainingSize < blockSize {
			zeroBlock = make([]byte, remainingSize)
		}
		_, err := file.Write(zeroBlock)
		if err != nil {
			output += fmt.Sprintf("Error escribiendo ceros: %v\n", err)
			return output
		}
		remainingSize -= blockSize
	}

	var nuevoMRB MRB
	nuevoMRB.MbrSize = int32(size)
	nuevoMRB.Signature = rand.Int31()
	copy(nuevoMRB.Fit[:], fit)

	tiempo := time.Now()
	fecha := tiempo.Format("2006-01-02")
	copy(nuevoMRB.CreationDate[:], fecha)

	if err := WriteObject(file, nuevoMRB, 0); err != nil {
		output += fmt.Sprintf("Error al escribir el MBR: %v\n", err)
		return output
	}

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error al leer el MBR: %v\n", err)
		return output
	}

	output += PrintMBR(TempMBR)

	defer file.Close()

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("mkdisk finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")
	return output
}

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) string {
	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando fdisk...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("Tamaño: %d\n", size)
	output += fmt.Sprintf("Path: %s\n", path)
	output += fmt.Sprintf("Nombre: %s\n", name)
	output += fmt.Sprintf("Unidad: %s\n", unit)
	output += fmt.Sprintf("Tipo: %s\n", type_)
	output += fmt.Sprintf("Fit: %s\n", fit)

	if fit != "bf" && fit != "ff" && fit != "wf" {
		fmt.Println("Error: Fit debe ser 'bf', 'ff', o 'wf'")
		output += fmt.Sprintf("Error: Fit debe ser 'bf', 'ff', o 'wf'\n")
		return output
	}

	if size <= 0 {
		fmt.Println("Error: el tamaño debe ser mayor a 0")
		output += fmt.Sprintf("Error: el tamaño debe ser mayor a 0\n")
		return output
	}

	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Println("Error: la unidad debe ser 'b', 'k', or 'm'")
		output += fmt.Sprintf("Error: la unidad debe ser 'b', 'k', or 'm'\n")
		return output
	}

	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	file, err := OpenFile(path)
	if err != nil {
		fmt.Println("Error: no se puede abrir archivo en path:", path)
		output += fmt.Sprintf("Error: no se puede abrir archivo en path: %s\n", path)
		return output
	}

	var TempMBR MRB

	if err := ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se puede leer MBR del archivo")
		output += fmt.Sprintf("Error: No se puede leer MBR del archivo\n")
		return output
	}

	if NombreParticionExiste(TempMBR, name, file) {
		output += fmt.Sprintf("Error: Ya existe una partición con el nombre '%s'.\n", name)
		return output
	}

	var particionesPrimarias, particionesExtendidas, particionesTotales int
	var espacioUsado int32 = 0

	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Size != 0 {
			particionesTotales++
			espacioUsado += TempMBR.Particion[i].Size

			if TempMBR.Particion[i].Type[0] == 'p' {
				particionesPrimarias++
			} else if TempMBR.Particion[i].Type[0] == 'e' {
				particionesExtendidas++
			}
		}
	}

	if type_ == "e" && particionesExtendidas > 0 {
		fmt.Println("Error: Solo se permite una partición extendida por disco.")
		output += fmt.Sprintf("Error: Solo se permite una partición extendida por disco.\n")
		return output
	}

	if type_ == "l" && particionesExtendidas == 0 {
		fmt.Println("Error: No se puede crear una partición lógica sin una partición extendida.")
		output += fmt.Sprintf("Error: No se puede crear una partición lógica sin una partición extendida.\n")
		return output
	}

	if type_ != "l" && particionesTotales >= 4 {
		output += fmt.Sprintf("Error: No se pueden crear más de 4 particiones primarias o extendidas en total.\n")
		return output
	}

	if type_ != "l" && espacioUsado+int32(size) > TempMBR.MbrSize {
		output += fmt.Sprintf("Error: No hay suficiente espacio en el disco para crear esta partición.\n")
		return output
	}

	var posicionInicial int32 = int32(binary.Size(TempMBR))
	if particionesTotales > 0 {
		posicionInicial = TempMBR.Particion[particionesTotales-1].Start + TempMBR.Particion[particionesTotales-1].Size
	}

	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				TempMBR.Particion[i].Size = int32(size)
				TempMBR.Particion[i].Start = posicionInicial
				copy(TempMBR.Particion[i].Name[:], name)
				copy(TempMBR.Particion[i].Fit[:], fit)
				copy(TempMBR.Particion[i].Status[:], "0")
				copy(TempMBR.Particion[i].Type[:], type_)
				TempMBR.Particion[i].Correlative = int32(particionesTotales + 1)

				if type_ == "e" {
					ebrStart := posicionInicial
					ebr := EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					WriteObject(file, ebr, int64(ebrStart))
				}

				break
			}
		}
	}

	if type_ == "l" {
		for i := 0; i < 4; i++ {
			if TempMBR.Particion[i].Type[0] == 'e' {
				ebrSig := TempMBR.Particion[i].Start
				var ebr EBR

				for {
					ReadObject(file, &ebr, int64(ebrSig))
					if ebr.PartNext == -1 {
						break
					}
					ebrSig = ebr.PartNext
				}

				espacioDisponible := TempMBR.Particion[i].Start + TempMBR.Particion[i].Size - (ebr.PartStart + ebr.PartSize)
				espacioNecesario := int32(binary.Size(ebr)) + int32(size)

				if espacioNecesario > espacioDisponible {
					output += fmt.Sprintf("Error: No hay suficiente espacio dentro de la partición extendida para la nueva partición lógica.\n")
					output += fmt.Sprintf("Espacio disponible: %d\n", espacioDisponible)
					output += fmt.Sprintf("Espacio necesario: %d\n", espacioNecesario)
					return output
				}

				newebrSig := ebr.PartStart + ebr.PartSize
				logicalPartitionStart := newebrSig + int32(binary.Size(ebr))

				ebr.PartNext = newebrSig
				WriteObject(file, ebr, int64(ebrSig))

				newEBR := EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				WriteObject(file, newEBR, int64(newebrSig))

				output += fmt.Sprintf("   \n")
				output += fmt.Sprintf("-> Nuevo EBR creado:\n")
				output += PrintEBR(newEBR)
				output += fmt.Sprintf("   \n")
				break
			}
		}
	}

	if err := WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		output += fmt.Sprintf("Error: Could not write MBR to file\n")
		return output
	}

	var TempMBR2 MRB

	if err := ReadObject(file, &TempMBR2, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file after writing")
		output += fmt.Sprintf("Error: Could not read MBR from file after writing\n")
		return output
	}

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-> MBR Actualizado:\n")
	output += fmt.Sprintf("   \n")
	output += PrintMBR(TempMBR2)

	defer file.Close()
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("fdisk finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")
	return output
}

func NombreParticionExiste(mbr MRB, nombre string, file *os.File) bool {

	nombre = strings.TrimSpace(strings.ToLower(nombre))

	for i := 0; i < 4; i++ {
		existingName := strings.Trim(string(mbr.Particion[i].Name[:]), "\x00 ")
		existingName = strings.ToLower(existingName)
		fmt.Printf("Comparando partición: '%s' con '%s'\n", existingName, nombre)
		if existingName == nombre {
			return true
		}
	}

	for i := 0; i < 4; i++ {
		if mbr.Particion[i].Type[0] == 'e' {
			var ebr EBR
			ebrSig := mbr.Particion[i].Start

			for {
				ReadObject(file, &ebr, int64(ebrSig))
				ebrName := strings.Trim(string(ebr.PartName[:]), "\x00 ")
				ebrName = strings.ToLower(ebrName)
				fmt.Printf("Comparando partición lógica: '%s' con '%s'\n", ebrName, nombre)
				if ebrName == nombre {
					return true
				}
				if ebr.PartNext == -1 {
					break
				}
				ebrSig = ebr.PartNext
			}
		}
	}

	fmt.Println("No se encontró ningún nombre duplicado.")
	return false
}

type particionMount struct {
	Path     string
	Name     string
	ID       string
	Status   byte
	LoggedIn bool
}

var particionesMontadas = make(map[string][]particionMount)

func PrintMount() string {
	var output string

	output += "-> Particiones montadas:\n"
	output += " \n"

	if len(particionesMontadas) == 0 {
		output += "-> No hay particiones montadas.\n"
		return output
	}

	for diskID, partitions := range particionesMontadas {
		output += fmt.Sprintf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			output += fmt.Sprintf("Nombre partición: %s, ID: %s, Path: %s, Status: %c\n",
				partition.Name, partition.ID, partition.Path, partition.Status)
		}
	}
	output += "\n"
	return output
}

func GetMountedPartitions() map[string][]particionMount {
	return particionesMontadas
}

func MarkPartitionAsLoggedIn(id string) string {

	var output string

	for diskID, partitions := range particionesMontadas {
		for i, partition := range partitions {
			if partition.ID == id {
				particionesMontadas[diskID][i].LoggedIn = true
				fmt.Printf("Partición con ID %s marcada como logueada.\n", id)
				output += fmt.Sprintf("-> Partición con ID %s marcada como logueada.\n", id)
				return output
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
	output += fmt.Sprintf("-> No se encontró la partición con ID %s para marcarla como logueada.\n", id)
	return output
}

func MarkPartitionAsLoggedOut(id string) string {
	var output string
	for diskID, partitions := range particionesMontadas {
		for i, partition := range partitions {
			if partition.ID == id {
				particionesMontadas[diskID][i].LoggedIn = false
				fmt.Printf("Partición con ID %s marcada como deslogueada.\n", id)
				output += fmt.Sprintf("-> Partición con ID %s marcada como deslogueada.\n", id)
				return output
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como deslogueada.\n", id)
	output += fmt.Sprintf("-> No se encontró la partición con ID %s para marcarla como deslogueada.\n", id)
	return output
}

func CleanMountedPartitions() {
	particionesMontadas = make(map[string][]particionMount)
}

func Mount(path string, nombre string) string {

	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando mount...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("   \n")

	file, err := OpenFile(path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo en la ruta:", path)
		output += fmt.Sprintf("Error: No se pudo abrir el archivo en la ruta: %s\n", path)
		return output
	}
	defer file.Close()

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		output += fmt.Sprintf("Error: No se pudo leer el MBR desde el archivo\n")
		return output
	}

	fmt.Printf("Buscando partición con nombre: '%s'\n", nombre)
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("Buscando partición con nombre: '%s'\n", nombre)
	output += fmt.Sprintf("   \n")

	particionEncontrada := false
	var partition Particion
	var indiceP int

	nombreBytes := [16]byte{}
	copy(nombreBytes[:], []byte(nombre))

	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Type[0] == 'p' && bytes.Equal(TempMBR.Particion[i].Name[:], nombreBytes[:]) {
			partition = TempMBR.Particion[i]
			indiceP = i
			particionEncontrada = true
			break
		}
	}

	if !particionEncontrada {
		fmt.Println("Error: Partición no encontrada o no es una partición primaria")
		output += fmt.Sprintf("Error: Partición no encontrada o no es una partición primaria\n")
		return output
	}

	if partition.Status[0] == '1' {
		fmt.Println("Error: La partición ya está montada")
		output += fmt.Sprintf("Error: La partición ya está montada\n")
		return output
	}

	diskID := generarID(path)

	indicePMount := particionesMontadas[diskID]

	var letter byte

	if len(particionesMontadas) == 0 {
		letter = 'A'
	} else if len(indicePMount) == 0 {
		ultimoID := obtenerUltimoID()
		lastLetter := particionesMontadas[ultimoID][0].ID[len(particionesMontadas[ultimoID][0].ID)-1]
		letter = lastLetter + 1
	} else {
		letter = indicePMount[0].ID[len(indicePMount[0].ID)-1]
	}

	carnet := "202201318"
	digitos := carnet[len(carnet)-2:]
	numeroParticion := len(indicePMount) + 1
	partitionID := fmt.Sprintf("%s%d%s", digitos, numeroParticion, strings.ToLower(string(letter)))

	partition.Status[0] = '1'
	copy(partition.Id[:], []byte(strings.ToLower(partitionID)))
	TempMBR.Particion[indiceP] = partition
	particionesMontadas[diskID] = append(particionesMontadas[diskID], particionMount{
		Path:   path,
		Name:   nombre,
		ID:     strings.ToLower(partitionID),
		Status: '1',
	})

	if err := WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		output += fmt.Sprintf("Error: No se pudo sobrescribir el MBR en el archivo\n")
		return output
	}

	fmt.Printf("Partición montada con ID: %s\n", partitionID)
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-> Partición montada con ID: %s\n", partitionID)

	fmt.Println("")
	fmt.Println("MBR del disco actualizado:")

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-> MBR actualizado:\n")

	output += PrintMBR(TempMBR)
	fmt.Println("")
	output += fmt.Sprintf("   \n")

	output += PrintMount()

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("mount finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")

	return output
}

func obtenerUltimoID() string {
	var ultimoID string
	for diskID := range particionesMontadas {
		ultimoID = diskID
	}
	return ultimoID
}

func generarID(path string) string {
	return strings.ToLower(path)
}

func GenerateMBRReport(mbr MRB, ebrs []EBR, outputPath string, file *os.File) error {

	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	content := "digraph G {\n"
	content += "\tnode [shape=plaintext]\n"
	content += "\tsubgraph cluster_MBR {\n"
	content += "\t\tlabel=\"Reporte del MBR\"\n"
	content += "\t\tfontsize=20;\n"
	content += "\t\tMBR [label=<\n"
	content += "\t\t\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"lightgray\">\n"
	content += "\t\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"dodgerblue3\"><B>MBR</B></TD></TR>\n"
	content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightblue\"><B>Tamaño</B></TD><TD>%d bytes</TD></TR>\n", mbr.MbrSize)
	content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightblue\"><B>Fecha Creación</B></TD><TD>%s</TD></TR>\n", string(mbr.CreationDate[:]))
	content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightblue\"><B>Signature</B></TD><TD>%d</TD></TR>\n", mbr.Signature)

	for i := 0; i < 4; i++ {
		part := mbr.Particion[i]
		if part.Size > 0 {
			content += fmt.Sprintf("\t\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightgreen\"><B>Partición %d</B></TD></TR>\n", i+1)
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Status</B></TD><TD>%s</TD></TR>\n", string(part.Status[:]))
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Type</B></TD><TD>%s</TD></TR>\n", string(part.Type[:]))
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Fit</B></TD><TD>%s</TD></TR>\n", string(part.Fit[:]))
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Start</B></TD><TD>%d</TD></TR>\n", part.Start)
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Size</B></TD><TD>%d</TD></TR>\n", part.Size)
			content += fmt.Sprintf("\t\t\t\t<TR><TD BGCOLOR=\"lightcyan\"><B>Name</B></TD><TD>%s</TD></TR>\n", strings.TrimRight(string(part.Name[:]), "\x00"))

			if string(part.Type[:]) == "e" {
				content += "\t\t\t\t<TR><TD COLSPAN=\"2\" BGCOLOR=\"lightcoral\"><B>EBRs</B></TD></TR>\n"
				for _, ebr := range ebrs {
					if ebr.PartStart >= part.Start && ebr.PartStart < (part.Start+part.Size) {
						content += "\t\t\t\t<TR><TD COLSPAN=\"2\"><TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" BGCOLOR=\"lightgray\">\n"
						content += fmt.Sprintf("\t\t\t\t\t<TR><TD BGCOLOR=\"lightpink\"><B>EBR Start</B></TD><TD>%d</TD></TR>\n", ebr.PartStart)
						content += fmt.Sprintf("\t\t\t\t\t<TR><TD BGCOLOR=\"lightpink\"><B>EBR Size</B></TD><TD>%d</TD></TR>\n", ebr.PartSize)
						content += fmt.Sprintf("\t\t\t\t\t<TR><TD BGCOLOR=\"lightpink\"><B>EBR Next</B></TD><TD>%d</TD></TR>\n", ebr.PartNext)
						content += fmt.Sprintf("\t\t\t\t\t<TR><TD BGCOLOR=\"lightpink\"><B>EBR Name</B></TD><TD>%s</TD></TR>\n", strings.TrimRight(string(ebr.PartName[:]), "\x00"))
						content += "\t\t\t\t</TABLE></TD></TR>\n"
					}
				}
			}
		}
	}

	content += "\t\t\t</TABLE>\n"
	content += "\t\t>];\n"
	content += "\t}\n"
	content += "}\n"

	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte MBR generado exitosamente en:", dotFilePath)
	return nil
}

func ConvertDotToPDF(dotFilePath, pdfFilePath string) error {
	cmd := exec.Command("dot", "-Tpdf", dotFilePath, "-o", pdfFilePath)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error al convertir .dot a PDF: %v", err)
	}

	fmt.Printf("Archivo PDF generado exitosamente: %s\n", pdfFilePath)
	return nil
}

func GenerateDiskReport(mbr MRB, ebrs []EBR, outputPath string) error {

	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	content := "digraph G {\n"
	content += "\trankdir=LR;\n"
	content += "\tnode [shape=plaintext];\n"
	content += "\tsubgraph cluster_0 {\n"
	content += "\t\tlabel=\"Reporte DISK: " + filepath.Base(outputPath) + "\";\n"
	content += "\t\tfontsize=20;\n"
	content += "\t\ttable [label=<\n"
	content += "\t\t\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n"
	content += "\t\t\t\t<TR>\n"

	tamDisco := mbr.MbrSize
	espacioOcupado := int32(0)

	content += fmt.Sprintf("\t\t\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"gainsboro\">MBR<br/>%d bytes</TD>\n", binary.Size(mbr))
	espacioOcupado += int32(binary.Size(mbr))

	for i := 0; i < 4; i++ {
		part := mbr.Particion[i]

		if part.Size > 0 {
			espacioOcupado += part.Size
			porcentajeParticion := (float64(part.Size) / float64(tamDisco)) * 100
			nombreParticion := strings.TrimRight(string(part.Name[:]), "\x00")

			if string(part.Type[:]) == "e" {

				content += fmt.Sprintf("\t\t\t\t<TD COLSPAN=\"%d\" BGCOLOR=\"lightblue\">Extendida<br/>%.2f%% del Disco</TD>\n", len(ebrs)*2+1, porcentajeParticion)
				content += "\t\t\t</TR>\n<TR>\n"

				currentEBRStart := part.Start
				for _, ebr := range ebrs {
					if ebr.PartStart >= part.Start && ebr.PartStart < (part.Start+part.Size) {
						ebrPercentage := (float64(ebr.PartSize) / float64(tamDisco)) * 100

						content += fmt.Sprintf("\t\t\t\t<TD BGCOLOR=\"pink\">EBR<br/>%.2f%% del Disco</TD>\n", (float64(binary.Size(ebr))/float64(tamDisco))*100)
						content += fmt.Sprintf("\t\t\t\t<TD BGCOLOR=\"darkolivegreen2\">Lógica<br/>%.2f%% del Disco</TD>\n", ebrPercentage)

						currentEBRStart = ebr.PartStart + ebr.PartSize
					}
				}

				espacioDisponibleEx := part.Start + part.Size - currentEBRStart
				if espacioDisponibleEx > 0 {
					porcentajeLibre := (float64(espacioDisponibleEx) / float64(tamDisco)) * 100
					content += fmt.Sprintf("\t\t\t\t<TD BGCOLOR=\"navajowhite1\">Libre<br/>%.2f%% de extendida</TD>\n", porcentajeLibre)
				}
			} else {
				content += fmt.Sprintf("\t\t\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"mediumpurple1\">%s<br/>%.2f%% del Disco</TD>\n", nombreParticion, porcentajeParticion)
			}
		}
	}

	espacioDisponible := tamDisco - espacioOcupado
	if espacioDisponible > 0 {
		porcentajeLibre := (float64(espacioDisponible) / float64(tamDisco)) * 100
		content += fmt.Sprintf("\t\t\t\t<TD ROWSPAN=\"2\" BGCOLOR=\"navajowhite1\">Libre<br/>%.2f%% del Disco</TD>\n", porcentajeLibre)
	}

	content += "\t\t\t\t</TR>\n"
	content += "\t\t\t</TABLE>\n"
	content += "\t\t>];\n"
	content += "\t}\n"
	content += "}\n"

	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte DISK generado exitosamente en:", dotFilePath)
	return nil
}

func DeleteParticion(path string, name string, delete_ string) string {

	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando delete...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("   \n")

	output += fmt.Sprintf("Path: %s\n", path)
	output += fmt.Sprintf("Nombre: %s\n", name)
	output += fmt.Sprintf("Tipo de eliminación: %s\n", delete_)
	output += fmt.Sprintf("   \n")

	file, err := OpenFile(path)
	if err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		fmt.Println("Error: Could not open file at path:", path)
		return output
	}

	var TempMBR MRB

	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		fmt.Println("Error: Could not read MBR from file")
		return output
	}

	found := false
	for i := 0; i < 4; i++ {

		partitionName := strings.TrimRight(string(TempMBR.Particion[i].Name[:]), "\x00")
		if partitionName == name {
			found = true

			if TempMBR.Particion[i].Type[0] == 'e' {
				output += fmt.Sprintf("Eliminando particiones lógicas dentro de la partición extendida...\n")
				ebrPos := TempMBR.Particion[i].Start
				var ebr EBR
				for {
					err := ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						output += fmt.Sprintf("Error al leer EBR: %v\n", err)
						break
					}

					if ebr.PartStart == 0 && ebr.PartSize == 0 {
						output += fmt.Sprintf("EBR vacío encontrado, deteniendo la búsqueda.\n")
						break
					}

					output += fmt.Sprintf("   \n")
					output += fmt.Sprintf("EBR leído antes de eliminar:\n")
					output += PrintEBR(ebr)
					output += fmt.Sprintf("   \n")

					if delete_ == "fast" {
						ebr = EBR{}
						WriteObject(file, ebr, int64(ebrPos))
					} else if delete_ == "full" {
						FillWithZeros(file, ebr.PartStart, ebr.PartSize)
						ebr = EBR{}
						WriteObject(file, ebr, int64(ebrPos))
					}

					output += fmt.Sprintf("EBR después de eliminar:\n")
					output += PrintEBR(ebr)

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}

			if delete_ == "fast" {
				TempMBR.Particion[i] = Particion{}
				output += fmt.Sprintf("Partición eliminada en modo Fast.\n")
			} else if delete_ == "full" {
				start := TempMBR.Particion[i].Start
				size := TempMBR.Particion[i].Size
				TempMBR.Particion[i] = Particion{}
				FillWithZeros(file, start, size)
				output += fmt.Sprintf("Partición eliminada en modo Full.\n")
				VerifyZeros(file, start, size)
			}
			break
		}
	}

	if !found {

		output += fmt.Sprintf("Buscando en particiones lógicas dentro de las extendidas...\n")
		for i := 0; i < 4; i++ {
			if TempMBR.Particion[i].Type[0] == 'e' {
				ebrPos := TempMBR.Particion[i].Start
				var ebr EBR
				for {
					err := ReadObject(file, &ebr, int64(ebrPos))
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}

					output += fmt.Sprintf("   \n")
					output += fmt.Sprintf("EBR leído:\n")
					output += PrintEBR(ebr)

					logicalName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if logicalName == name {
						found = true

						if delete_ == "fast" {
							ebr = EBR{}
							WriteObject(file, ebr, int64(ebrPos))
							fmt.Println("Partición lógica eliminada en modo Fast.")
						} else if delete_ == "full" {
							FillWithZeros(file, ebr.PartStart, ebr.PartSize)
							ebr = EBR{}
							WriteObject(file, ebr, int64(ebrPos))
							VerifyZeros(file, ebr.PartStart, ebr.PartSize)
							fmt.Println("Partición lógica eliminada en modo Full.")
						}
						break
					}

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		output += fmt.Sprintf("Error: No se encontró la partición con el nombre: %s\n", name)
		return output
	}

	if err := WriteObject(file, TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output
	}

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("MBR actualizado después de la eliminación:\n")
	output += PrintMBR(TempMBR)

	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Type[0] == 'e' {
			fmt.Println("Imprimiendo EBRs actualizados en la partición extendida:")
			ebrPos := TempMBR.Particion[i].Start
			var ebr EBR
			for {
				err := ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					fmt.Println("Error al leer EBR:", err)
					break
				}

				if ebr.PartStart == 0 && ebr.PartSize == 0 {
					output += fmt.Sprintf("EBR vacío encontrado, deteniendo la búsqueda.\n")
					break
				}

				output += fmt.Sprintf("   \n")
				output += fmt.Sprintf("EBR leído después de actualización:\n")
				output += PrintEBR(ebr)
				output += fmt.Sprintf("   \n")

				if ebr.PartNext == -1 {
					break
				}
				ebrPos = ebr.PartNext
			}
		}
	}

	defer file.Close()

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("delete finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")

	return output
}

func AddParticion(path string, name string, add int, unit string) (string, error) {

	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando add...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("   \n")

	file, err := OpenFile(path)
	if err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output, err
	}
	defer file.Close()

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error: %v\n", err)
		return output, err
	}

	output += fmt.Sprintf("MBR antes de la modificación:\n")
	output += PrintMBR(TempMBR)

	var foundPartition *Particion
	var partitionType byte

	for i := 0; i < 4; i++ {
		partitionName := strings.TrimRight(string(TempMBR.Particion[i].Name[:]), "\x00")
		if partitionName == name {
			foundPartition = &TempMBR.Particion[i]
			partitionType = TempMBR.Particion[i].Type[0]
			break
		}
	}

	if foundPartition == nil {
		for i := 0; i < 4; i++ {
			if TempMBR.Particion[i].Type[0] == 'e' {
				ebrPos := TempMBR.Particion[i].Start
				var ebr EBR
				for {
					if err := ReadObject(file, &ebr, int64(ebrPos)); err != nil {
						output += fmt.Sprintf("Error al leer EBR: %v\n", err)
						fmt.Println("Error al leer EBR:", err)
						return output, err
					}

					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00")
					if ebrName == name {
						partitionType = 'l'
						foundPartition = &Particion{
							Start: ebr.PartStart,
							Size:  ebr.PartSize,
						}
						break
					}

					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}
				if foundPartition != nil {
					break
				}
			}
		}
	}

	if foundPartition == nil {
		fmt.Println("Error: No se encontró la partición con el nombre:", name)
		output += fmt.Sprintf("Error: No se encontró la partición con el nombre: %s\n", name)
		return output, nil
	}

	var addBytes int
	if unit == "k" {
		addBytes = add * 1024
	} else if unit == "m" {
		addBytes = add * 1024 * 1024
	} else {
		output += fmt.Sprintf("Error: Unidad desconocida, debe ser 'k' o 'm'\n")
		fmt.Println("Error: Unidad desconocida, debe ser 'k' o 'm'")
		return output, nil
	}

	var shouldModify = true

	if add > 0 {
		nextPartitionStart := foundPartition.Start + foundPartition.Size
		if partitionType == 'l' {
			for i := 0; i < 4; i++ {
				if TempMBR.Particion[i].Type[0] == 'e' {
					extendedPartitionEnd := TempMBR.Particion[i].Start + TempMBR.Particion[i].Size
					if nextPartitionStart+int32(addBytes) > extendedPartitionEnd {
						fmt.Println("Error: No hay suficiente espacio libre dentro de la partición extendida")
						shouldModify = false
					}
					break
				}
			}
		} else {
			if nextPartitionStart+int32(addBytes) > TempMBR.MbrSize {
				fmt.Println("Error: No hay suficiente espacio libre después de la partición")
				shouldModify = false
			}
		}
	} else {
		if foundPartition.Size+int32(addBytes) < 0 {
			fmt.Println("Error: No es posible reducir la partición por debajo de 0")
			shouldModify = false
		}
	}

	if shouldModify {
		foundPartition.Size += int32(addBytes)
	} else {
		output += fmt.Sprintf("No se realizaron modificaciones debido a un error.\n")
		fmt.Println("No se realizaron modificaciones debido a un error.")
		return output, nil
	}

	if partitionType == 'l' {
		ebrPos := foundPartition.Start
		var ebr EBR
		if err := ReadObject(file, &ebr, int64(ebrPos)); err != nil {
			output += fmt.Sprintf("Error al leer EBR: %v\n", err)
			fmt.Println("Error al leer EBR:", err)
			return output, err
		}

		ebr.PartSize = foundPartition.Size
		if err := WriteObject(file, ebr, int64(ebrPos)); err != nil {
			output += fmt.Sprintf("Error al escribir EBR: %v\n", err)
			fmt.Println("Error al escribir el EBR actualizado:", err)
			return output, err
		}

		output += fmt.Sprintf("  ")
		output += fmt.Sprintf("EBR actualizado:\n")
		output += PrintEBR(ebr)
	}

	if err := WriteObject(file, TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error al escribir MBR: %v\n", err)
		fmt.Println("Error al escribir el MBR actualizado:", err)
		return output, err
	}

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("MBR después de la modificación:\n")
	output += PrintMBR(TempMBR)
	output += fmt.Sprintf("   \n")

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("add finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")

	return output, nil
}
