package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	uuid2 "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"image"
	"image/png"
	"mymcuu.id/api/mojang"
	"net/http"
	"os"
)

var CORS = "https://mymcuu.id"

type JsonError struct {
	Error string `json:"error"`
}

type UUIDResponse struct {
	UUID string `json:"uuid"`
	Username string `json:"username"`
	Avatar string `json:"avatar"`
	Skin string `json:"skin"`
}
func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Cache-Control", "public, max-age=86400")
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
	if ok, _ := HasDataFromUsername(r.Context(), username); ok {
		resp, err := GetDataFromUsername(r.Context(), username)
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
	skinImage, err := mojang.GetSkinFromUUID(resp.UUID)
	skinBuf := new(bytes.Buffer)
	png.Encode(skinBuf, *skinImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	headImage, err := mojang.GetHeadFromSkin(skinImage)
	headBuf := new(bytes.Buffer)
	png.Encode(headBuf, *headImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	withDashes, err := uuid2.Parse(resp.UUID)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	bytes, err := json.Marshal(UUIDResponse{
		UUID:     withDashes.String(),
		Username: resp.Name,
		Avatar: fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(headBuf.Bytes())),
		Skin: fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(skinBuf.Bytes())),
	})
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	jsonResponse := string(bytes)
	err = StoreData(r.Context(), username, withDashes.String(), jsonResponse)
	if err != nil {
		fmt.Printf("failed to save %s to cache", username)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, jsonResponse)
}

func GetSkinFromUUID(w http.ResponseWriter, r *http.Request){
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	parsedUUID, err := uuid2.Parse(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson("invalid uuid"))
		return
	}
	if ok, _ := HasDataFromUUID(r.Context(), parsedUUID.String()); ok {
		resp, err := GetDataFromUUID(r.Context(), parsedUUID.String())
		if err != nil {
			fmt.Fprintf(w, ErrorJson(err.Error()))
			return
		}
		if resp != nil {
			var response *UUIDResponse
			json.Unmarshal([]byte(*resp), &response)
			if response != nil {
				base := response.Skin[22:]
				unbased, err := base64.StdEncoding.DecodeString(base)
				if err != nil {
					fmt.Fprintf(w, ErrorJson(err.Error()))
					return
				}
				r := bytes.NewReader(unbased)
				im, err := png.Decode(r)
				if err != nil {
					fmt.Fprintf(w, ErrorJson(err.Error()))
					return
				}
				w.WriteHeader(http.StatusOK)
				png.Encode(w, im)
				return
			}
		}
	}
	profile, err := mojang.GetProfileFromUUID(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	skinImage, err := mojang.GetSkinFromProfile(*profile)
	skinBuf := new(bytes.Buffer)
	png.Encode(skinBuf, *skinImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	headImage, err := mojang.GetHeadFromSkin(skinImage)
	headBuf := new(bytes.Buffer)
	png.Encode(headBuf, *headImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	bytes, err := json.Marshal(UUIDResponse{
		UUID:     parsedUUID.String(),
		Username: profile.Name,
		Avatar:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(headBuf.Bytes())),
		Skin:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(skinBuf.Bytes())),
	})
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	jsonResponse := string(bytes)
	err = StoreData(r.Context(), profile.Name, parsedUUID.String(), jsonResponse)
	if err != nil {
		fmt.Printf("failed to save %s to cache", profile.Name)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	w.Write(skinBuf.Bytes())
}

func GetHeadFromUUID(w http.ResponseWriter, r *http.Request){
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	parsedUUID, err := uuid2.Parse(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson("invalid uuid"))
		return
	}
	if ok, _ := HasDataFromUUID(r.Context(), parsedUUID.String()); ok {
		resp, err := GetDataFromUUID(r.Context(), parsedUUID.String())
		if err != nil {
			fmt.Fprintf(w, ErrorJson(err.Error()))
			return
		}
		if resp != nil {
			var response *UUIDResponse
			json.Unmarshal([]byte(*resp), &response)
			if response != nil {
				base := response.Avatar[22:]
				unbased, err := base64.StdEncoding.DecodeString(base)
				if err != nil {
					fmt.Fprintf(w, ErrorJson(err.Error()))
					return
				}
				r := bytes.NewReader(unbased)
				im, err := png.Decode(r)
				if err != nil {
					fmt.Fprintf(w, ErrorJson(err.Error()))
					return
				}
				w.WriteHeader(http.StatusOK)
				png.Encode(w, im)
				return
			}
		}
	}
	profile, err := mojang.GetProfileFromUUID(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	skinImage, err := mojang.GetSkinFromProfile(*profile)
	skinBuf := new(bytes.Buffer)
	png.Encode(skinBuf, *skinImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	headImage, err := mojang.GetHeadFromSkin(skinImage)
	headBuf := new(bytes.Buffer)
	png.Encode(headBuf, *headImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	bytes, err := json.Marshal(UUIDResponse{
		UUID:     parsedUUID.String(),
		Username: profile.Name,
		Avatar:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(headBuf.Bytes())),
		Skin:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(skinBuf.Bytes())),
	})
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	jsonResponse := string(bytes)
	err = StoreData(r.Context(), profile.Name, parsedUUID.String(), jsonResponse)
	if err != nil {
		fmt.Printf("failed to save %s to cache", profile.Name)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, *headImage)
}

func getSteveHead() (*image.Image, error){
	resp, err := mojang.GetUUIDFromUsername("MHF_Steve")
	if err != nil {
		return nil, err
	}
	headImage, err := mojang.GetHeadFromUUID(resp.UUID)
	if err != nil {
		return nil, err
	}
	return headImage, nil
}

func Status(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func GetUsernameFromUUID(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	parsedUUID, err := uuid2.Parse(uuid)
	if err != nil {
		fmt.Fprintf(w, ErrorJson("invalid uuid"))
		return
	}
	if ok, _ := HasDataFromUUID(r.Context(), parsedUUID.String()); ok {
		resp, err := GetDataFromUUID(r.Context(), parsedUUID.String())
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
	profile, err := mojang.GetProfileFromUUID(uuid)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	skinImage, err := mojang.GetSkinFromProfile(*profile)
	skinBuf := new(bytes.Buffer)
	png.Encode(skinBuf, *skinImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	headImage, err := mojang.GetHeadFromSkin(skinImage)
	headBuf := new(bytes.Buffer)
	png.Encode(headBuf, *headImage)
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	bytes, err := json.Marshal(UUIDResponse{
		UUID:     parsedUUID.String(),
		Username: profile.Name,
		Avatar:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(headBuf.Bytes())),
		Skin:   fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(skinBuf.Bytes())),
	})
	if err != nil {
		fmt.Fprintf(w, ErrorJson(err.Error()))
		return
	}
	jsonResponse := string(bytes)
	err = StoreData(r.Context(), profile.Name, parsedUUID.String(), jsonResponse)
	if err != nil {
		fmt.Printf("failed to save %s to cache", profile.Name)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, jsonResponse)
}

func main(){
	err := godotenv.Load()
	if err != nil {
		//log.Fatal("Error loading .env file")
	}
	if len(os.Getenv("CORS")) > 0 {
		CORS = os.Getenv("CORS")
		fmt.Printf("Set cors to %v", CORS)
	}
	SetupRedis()
	r := mux.NewRouter()
	r.HandleFunc("/username/{username}", GetUUIDFromUsername)
	r.HandleFunc("/uuid/{uuid}", GetUsernameFromUUID)
	r.HandleFunc("/head/{uuid}", GetHeadFromUUID)
	r.HandleFunc("/skin/{uuid}", GetSkinFromUUID)
	r.HandleFunc("/status", Status)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
