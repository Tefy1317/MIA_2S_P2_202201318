package funciones

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func Mkfs(id string, type_ string, fs_ string) string {

	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Ejecutando mkfs...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += "Id: " + id + "\n"
	output += "Type: " + type_ + "\n"
	output += "Fs: " + fs_ + "\n"

	var mountedPartition particionMount
	var partitionFound bool

	for _, partitions := range GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		output += "Particion no encontrada\n"

		return output
	}

	if mountedPartition.Status != '1' {
		output += "Particion no montada\n"
		return output
	}

	file, err := OpenFile(mountedPartition.Path)
	if err != nil {
		output += "Error al abrir el archivo\n"
		return output
	}

	var TempMBR MRB
	if err := ReadObject(file, &TempMBR, 0); err != nil {
		output += "Error al leer el MBR\n"
		return output
	}

	output += PrintMBR(TempMBR)
	output += "-------------\n"

	fmt.Println("-------------")

	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Particion[i].Size != 0 {
			if strings.Contains(string(TempMBR.Particion[i].Id[:]), id) {
				index = i
				break
			}
		}
	}

	if index != -1 {
		PrintPartition(TempMBR.Particion[index])
	} else {
		output += "Particion no encontrada\n"
		return output
	}

	numerador := int32(TempMBR.Particion[index].Size - int32(binary.Size(Superblock{})))
	denominador_base := int32(4 + int32(binary.Size(Inode{})) + 3*int32(binary.Size(Fileblock{})))
	var temp int32 = 0

	if fs_ == "3fs" {
		temp = int32(binary.Size(Journaling{}))
	} else if fs_ == "2fs" {
		temp = 0
	} else {
		fmt.Println("Error: Sólo están disponibles los sistemas de archivos 2FS y 3FS.")
		output += "Error: Sólo están disponibles los sistemas de archivos 2FS y 3FS.\n"
		return output
	}

	denominador := denominador_base + temp
	n := int32(numerador / denominador)

	fmt.Println("INODOS:", n)
	output += "INODOS: "
	output += fmt.Sprint(n)

	var newSuperblock Superblock
	if fs_ == "2fs" {
		newSuperblock.S_filesystem_type = 2
	} else if fs_ == "3fs" {
		newSuperblock.S_filesystem_type = 3
	}
	newSuperblock.S_inodes_count = n
	newSuperblock.S_blocks_count = 3 * n
	newSuperblock.S_free_blocks_count = 3*n - 2
	newSuperblock.S_free_inodes_count = n - 2
	copy(newSuperblock.S_mtime[:], "26/10/2024")
	copy(newSuperblock.S_umtime[:], "26/10/2024")
	newSuperblock.S_mnt_count = 1
	newSuperblock.S_magic = 0xEF53
	newSuperblock.S_inode_size = int32(binary.Size(Inode{}))
	newSuperblock.S_block_size = int32(binary.Size(Fileblock{}))

	newSuperblock.S_bm_inode_start = TempMBR.Particion[index].Start + int32(binary.Size(Superblock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + n
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*n
	newSuperblock.S_block_start = newSuperblock.S_inode_start + n*newSuperblock.S_inode_size

	if fs_ == "2fs" {
		output += create_ext2(n, TempMBR.Particion[index], newSuperblock, "26/10/2024", file)
	} else if fs_ == "3fs" {
		output += create_ext3(n, TempMBR.Particion[index], newSuperblock, "26/10/2024", file)
	}

	defer file.Close()
	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("mkfs finalizado :)\n")
	output += fmt.Sprintf("-----------------------------\n")
	return output
}

func create_ext2(n int32, partition Particion, newSuperblock Superblock, date string, file *os.File) string {
	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Creando EXT2...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += "INODOS: " + fmt.Sprint(n) + "\n"

	fmt.Println("----------Creando EXT2----------")
	fmt.Println("INODOS:", n)

	PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	for i := int32(0); i < n; i++ {
		if err := WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			output += "Error: " + fmt.Sprint(err) + "\n"
			return output
		}
	}

	for i := int32(0); i < 3*n; i++ {
		if err := WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			output += "Error: " + fmt.Sprint(err) + "\n"
			return output
		}
	}

	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}

	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}

	if err := WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}

	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}

	for i := int32(0); i < 1; i++ {
		var fileblock Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Folderblock{})) + i*int32(binary.Size(Fileblock{})))
		if err := ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			output += "Error al leer Fileblock: " + fmt.Sprint(err) + "\n"
			return output
		}

		PrintFileblock(fileblock)
	}

	fmt.Println("----------EXT2 Creado----------")

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("EXT2 Creado :)\n")
	output += fmt.Sprintf("-----------------------------\n")
	return output
}

func create_ext3(n int32, partition Particion, newSuperblock Superblock, date string, file *os.File) string {

	var output string

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("Creando EXT3...\n")
	output += fmt.Sprintf("-----------------------------\n")
	output += "INODOS: " + fmt.Sprint(n) + "\n"

	fmt.Println("------Creando EXT3------")
	fmt.Println("INODOS:", n)

	PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	if err := initJournaling(newSuperblock, file); err != nil {
		fmt.Println("Error al inicializar el Journaling: ", err)
		output += "Error al inicializar el Journaling: " + fmt.Sprint(err) + "\n"
		return output
	}
	fmt.Println("Journaling inicializado correctamente.")
	output += "Journaling inicializado correctamente.\n"

	for i := int32(0); i < n; i++ {
		if err := WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			output += "Error: " + fmt.Sprint(err) + "\n"
			return output
		}
	}
	fmt.Println("Bitmap de inodos escrito correctamente.")
	output += "Bitmap de inodos escrito correctamente.\n"

	for i := int32(0); i < 3*n; i++ {
		if err := WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			output += "Error: " + fmt.Sprint(err) + "\n"
			return output
		}
	}
	fmt.Println("Bitmap de bloques escrito correctamente.")
	output += "Bitmap de bloques escrito correctamente.\n"

	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}
	fmt.Println("Inodos y bloques inicializados correctamente.")
	output += "Inodos y bloques inicializados correctamente.\n"

	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}
	fmt.Println("Carpeta raíz y archivo users.txt creados correctamente.")
	output += "Carpeta raíz y archivo users.txt creados correctamente.\n"

	if err := WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}
	fmt.Println("Superbloque escrito correctamente.")
	output += "Superbloque escrito correctamente.\n"

	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		output += "Error: " + fmt.Sprint(err) + "\n"
		return output
	}
	fmt.Println("Inodos y bloques iniciales marcados como usados correctamente.")
	output += "Inodos y bloques iniciales marcados como usados correctamente.\n"

	for i := int32(0); i < 1; i++ {
		var fileblock Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(Folderblock{})) + i*int32(binary.Size(Fileblock{})))
		if err := ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			output += "Error al leer Fileblock: " + fmt.Sprint(err) + "\n"
			return output
		}
		PrintFileblock(fileblock)
	}

	fmt.Println("Fileblocks impresos correctamente.")
	fmt.Println("------EXT3 Creado------")

	output += fmt.Sprintf("   \n")
	output += fmt.Sprintf("-----------------------------\n")
	output += fmt.Sprintf("EXT3 Creado :)\n")
	output += fmt.Sprintf("-----------------------------\n")
	return output
}

func initJournaling(newSuperblock Superblock, file *os.File) error {
	var journaling Journaling
	journaling.Size = 50
	journaling.Ultimo = 0

	journalingStart := newSuperblock.S_inode_start - int32(binary.Size(Journaling{}))*journaling.Size

	for i := 0; i < 50; i++ {
		if err := WriteObject(file, journaling, int64(journalingStart+int32(i*binary.Size(journaling)))); err != nil {
			return fmt.Errorf("error al inicializar el journaling: %v", err)
		}
	}

	fmt.Println("Journaling inicializado correctamente.")
	return nil
}

func initInodesAndBlocks(n int32, newSuperblock Superblock, file *os.File) error {
	var newInode Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

func createRootAndUsersFile(newSuperblock Superblock, date string, file *os.File) error {
	var Inode0, Inode1 Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize

	var Fileblock1 Fileblock
	copy(Fileblock1.B_content[:actualSize], data)

	fmt.Println("Contenido del bloque antes de escribir (byte por byte):")
	for i := 0; i < len(Fileblock1.B_content); i++ {
		fmt.Printf("Byte %d: %v (%c)\n", i, Fileblock1.B_content[i], Fileblock1.B_content[i])
	}

	var nullBytes int
	for i := 0; i < len(Fileblock1.B_content); i++ {
		if Fileblock1.B_content[i] == 0 {
			nullBytes++
		}
	}
	fmt.Printf("Cantidad de bytes nulos en el bloque: %d\n", nullBytes)

	var Folderblock0 Folderblock
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	if err := WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(Inode{})))); err != nil {
		return err
	}
	if err := WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(Folderblock{})))); err != nil {
		return err
	}

	return nil
}

func initInode(inode *Inode, date string) {
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	copy(inode.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_block[i] = -1
	}
}

func markUsedInodesAndBlocks(newSuperblock Superblock, file *os.File) error {
	if err := WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}
