package funciones

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type ActiveUser struct {
	Id       string
	User     string
	Password string
}

var activeUser ActiveUser

func Login(user string, pass string, id string) (string, error) {
	var output string

	output += "------------------------\n"
	output += "Ejecutando Login\n"
	output += "------------------------\n"
	output += fmt.Sprintf("User: %s\n", user)
	output += fmt.Sprintf("Pass: %s\n", pass)
	output += fmt.Sprintf("Id: %s\n", id)

	if activeUser.User != "" {
		output += "Ya existe un usuario logueado!\n"
		return output, fmt.Errorf("Ya existe un usuario logueado en esta partición")
	}

	mountedPartitions := GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var login bool = false

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id && partition.LoggedIn {
				output += "Ya existe un usuario logueado!\n"
				return output, fmt.Errorf("Ya existe un usuario logueado en esta partición")
			}
			if partition.ID == id {
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		output += "Error: No se encontró ninguna partición montada con el ID proporcionado\n"
		return output, fmt.Errorf("No se encontró ninguna partición montada con el ID proporcionado")
	}

	file, err := OpenFile(filepath)
	if err != nil {
		output += fmt.Sprintf("Error: No se pudo abrir el archivo: %v\n", err)
		return output, fmt.Errorf("Error al abrir el archivo")
	}
	defer file.Close()

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += fmt.Sprintf("Error: No se pudo leer el MBR: %v\n", err)
		return output, fmt.Errorf("Error al leer el MBR")
	}

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Size != 0 {
			if strings.Contains(string(TempMBR.Particion[i].Id[:]), id) {
				if TempMBR.Particion[i].Status[0] == '1' {
					index = i
				} else {
					return output, fmt.Errorf("La partición no está montada")
				}
				break
			}
		}
	}

	if index == -1 {
		output += "No se encontró la partición\n"
		return output, fmt.Errorf("No se encontró la partición con el ID proporcionado")
	}

	var tempSuperblock Superblock
	if err := ReadObject(file, &tempSuperblock, int64(TempMBR.Particion[index].Start)); err != nil {
		output += fmt.Sprintf("Error: No se pudo leer el Superblock: %v\n", err)
		return output, fmt.Errorf("Error al leer el Superblock")
	}

	indexInode := InitSearch("/users.txt", file, tempSuperblock)

	var crrInode Inode
	if err := ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Inode{})))); err != nil {
		output += fmt.Sprintf("Error: No se pudo leer el Inodo: %v\n", err)
		return output, fmt.Errorf("Error al leer el Inodo")
	}

	data := GetInodeFileData(crrInode, file, tempSuperblock)

	lines := strings.Split(data, "\n")

	for _, line := range lines {
		words := strings.Split(line, ",")

		if len(words) == 5 {
			if (strings.Contains(words[3], user)) && (strings.Contains(words[4], pass)) {
				login = true
				break
			}
		}
	}

	//output += fmt.Sprintf("Inode: %v\n", crrInode.I_block)

	if login {
		output += "Usuario logueado con éxito\n"
		activeUser.Id = id
		activeUser.User = user
		activeUser.Password = pass
		output += MarkPartitionAsLoggedIn(id)
		return "Inicio de sesión exitoso\n" + output, nil
	}

	output += "Credenciales incorrectas\n"
	return output, fmt.Errorf("Credenciales incorrectas")
}

func InitSearch(path string, file *os.File, tempSuperblock Superblock) int32 {
	fmt.Println("-------------Busqueda Inicial--------------")
	fmt.Println("path:", path)

	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 Inode
	if err := ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}

	fmt.Println("---------- Termina Busqueda Inicial ----------")

	return SarchInodeByPath(StepsPath, Inode0, file, tempSuperblock)
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

func SarchInodeByPath(StepsPath []string, InodeP Inode, file *os.File, tempSuperblock Superblock) int32 {
	fmt.Println("-----------------BUSQUEDA INODO POR PATH-----------------")

	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	fmt.Println("--> SearchedName:", SearchedName)

	for _, block := range InodeP.I_block {
		if block != -1 {
			if index < 13 {

				var crrFolderBlock Folderblock

				if err := ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Folderblock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {

					fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							fmt.Println("Folder found----------------")
							return folder.B_inodo
						} else {
							fmt.Println("NextInode----------------")
							var NextInode Inode
							if err := ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Inode{})))); err != nil {
								return -1
							}
							return SarchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("-----------------TERMINA BUSQUEDA INODO POR PATH-----------------")
	return 0
}

func GetInodeFileData(Inode Inode, file *os.File, tempSuperblock Superblock) string {
	fmt.Println("-----------------CONTENIDO DEL BLOQUE-----------------")
	index := int32(0)

	var content string

	for _, block := range Inode.I_block {
		if block != -1 {

			if index < 13 {
				var crrFileBlock Fileblock
				if err := ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Fileblock{})))); err != nil {
					return ""
				}

				content += string(crrFileBlock.B_content[:])

			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("-----------------TERMINA CONTENIDO DEL BLOQUE-----------------")
	return content
}

func Logout() (string, error) {
	var output string
	output += "------------------------\n"
	output += "Ejecutando Logout\n"
	output += "------------------------\n"

	if activeUser.User == "" {
		output += "No hay ninguna sesión activa para cerrar\n"
		return output, fmt.Errorf("No hay ninguna sesión activa")
	}
	output += MarkPartitionAsLoggedOut(activeUser.Id)
	output += fmt.Sprintf("Sesión cerrada exitosamente")
	output += fmt.Sprintf("Usuario: %s\n", activeUser.User)
	fmt.Println("Sesión cerrada exitosamente")
	activeUser = ActiveUser{}

	return output, nil
}
