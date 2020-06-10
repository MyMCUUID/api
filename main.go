package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"image/png"
	"mymcuu.id/api/mojang"
	"net/http"
)

type JsonError struct {
	Error string `json:"error"`
}

type UUIDResponse struct {
	UUID string `json:"uuid"`
	Username string `json:"username"`
	Avatar string `json:"avatar"`
}
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "https://mymcuu.id")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}


func ErrorJson(error string) string {
	bytes, _ := json.Marshal(JsonError{Error: error})
	return string(bytes)
}

func GetUUIDFromUsername(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	username := vars["username"]
	if ok, _ := HasData(r.Context(), username); ok {
		resp, err := GetData(r.Context(), username)
		if err != nil {
			fmt.Fprintf(w, ErrorJson(err.Error()))
			return
		}
		if resp != nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, *resp)
			return
		}
	}
	resp, err := mojang.GetUUIDFromUsername(username)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	headImage, err := mojang.GetHeadFromUUID(resp.UUID)
	buf := new(bytes.Buffer)
	png.Encode(buf, *headImage)
	withDashes, err := uuid.Parse(resp.UUID)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	bytes, err := json.Marshal(UUIDResponse{
		UUID:     withDashes.String(),
		Username: resp.Name,
		Avatar: fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes())),
	})
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	jsonResponse := string(bytes)
	err = StoreData(r.Context(), username, jsonResponse)
	if err != nil {
		fmt.Printf("failed to save %s to cache", username)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, jsonResponse)
}

func GetHeadFromUUID(w http.ResponseWriter, r *http.Request){
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	resp, err := mojang.GetHeadFromUUID(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
	}
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, *resp)
}

func main(){
	err := godotenv.Load()
	if err != nil {
		//log.Fatal("Error loading .env file")
	}
	SetupRedis()
	r := mux.NewRouter()
	r.HandleFunc("/username/{username}", GetUUIDFromUsername)
	r.HandleFunc("/head/{uuid}", GetHeadFromUUID)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
