package main

import (
	"backend/funciones"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type CommandRequest struct {
	Commands string `json:"commands"`
}

type CommandResponse struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func CommandExecute(comando string) string {
	lineas := strings.Split(comando, "\n")
	resultado := ""

	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)

		switch {
		case strings.HasPrefix(linea, "#"):
			resultado += "Comentario: " + linea + "\n"

		case linea == "":
			resultado += "\n"
			continue

		default:
			output := funciones.Analyze(linea)
			resultado += "Comando procesado: " + output + "\n"
		}
	}

	return resultado
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	var req CommandRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Error al leer los datos", http.StatusBadRequest)
		return
	}

	resultado := CommandExecute(req.Commands)

	response := CommandResponse{
		Message: "Comandos ejecutados correctamente",
		Result:  resultado,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/execute", handleExecute).Methods("POST")

	handler := cors.Default().Handler(r)

	fmt.Println("Servidor escuchando en http://localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", handler))
}
