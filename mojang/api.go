package mojang

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	image2 "image"
	"io"
	"io/ioutil"
	"net/http"
)

var client = http.Client{}

type UUIDResponse struct {
	UUID	 string `json:"id"`
	Name string `json:"name"`
}

type ProfileProperty struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type ProfileResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Properties []ProfileProperty `json:"properties"`
}

type TextureInformation struct {
	Timestamp int64 `json:"timestamp"`
	ProfileID string `json:"profileId"`
	ProfileName string `json:"profileName"`
	SignatureRequired bool `json:"signatureRequired"`
	Textures Textures `json:"textures"`
}

type Texture struct {
	Url string `json:"url"`
}

type Textures struct {
	Skin *Texture `json:"SKIN"`
	Cape *Texture `json:"CAPE"`
}

func GetUUIDFromUsername(username string) (*UUIDResponse, error) {
	if len(username) < 1 {
		return nil, fmt.Errorf("invalid username")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username), nil)
	request.Header.Set("User-Agent", "MyMCUUID-API")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(body))
		var uuid UUIDResponse
		err = json.Unmarshal(body, &uuid)
		if err != nil {
			return nil, err
		}
		return &uuid, nil
	}
	return nil, fmt.Errorf("no such player")
}

func GetProfileFromUUID(uuid string) (*ProfileResponse, error) {
	if len(uuid) < 1 {
		return nil, fmt.Errorf("invalid uuid")
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/profile/%s", uuid), nil)
	request.Header.Set("User-Agent", "MyMCUUID-API")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var profile ProfileResponse
		err = json.Unmarshal(body, &profile)
		if err != nil {
			return nil, err
		}
		return &profile, nil
	}
	return nil, fmt.Errorf("no such player")
}

func GetSkinFromProfile(profile ProfileResponse) (*image2.Image, error) {
	var texture *TextureInformation
	propertiesAsString, err := json.Marshal(profile)
	if err == nil {
		fmt.Println("Debug Data:")
		fmt.Println(string(propertiesAsString))
	}
	for _, val := range profile.Properties {
		if val.Name == "textures" {
			textures, err := base64.StdEncoding.DecodeString(profile.Properties[0].Value)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(textures, &texture)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if texture != nil {
		var image io.ReadCloser
		if texture.Textures.Skin == nil {
			image, err = GetImage("https://static.wikia.nocookie.net/minecraft_gamepedia/images/3/37/Steve_skin.png/revision/latest?cb=20191231170209")
		} else {
			image, err = GetImage(texture.Textures.Skin.Url)
		}
		if err != nil {
			return nil, err
		}
		aImage, err := imaging.Decode(image)
		if err != nil {
			return nil, err
		}
		return &aImage, nil
	}
	return nil, fmt.Errorf("something went wrong")
}

func GetHeadFromProfile(profile ProfileResponse) (*image2.Image, error) {
	skin, err := GetSkinFromProfile(profile)
	if err != nil {return nil, err}
	aImage := *skin
	headImage := imaging.Crop(aImage, image2.Rect(8, 8, 16, 16))
	helmImage := imaging.Crop(aImage, image2.Rect(40, 8, 40+8, 8+8))
	overlayedImage := imaging.Overlay(headImage, helmImage, image2.Pt(0, 0), 1)
	aImage = imaging.Resize(overlayedImage, 200, 200, imaging.NearestNeighbor)
	return &aImage, nil
}

func GetHeadFromSkin(skin *image2.Image) (*image2.Image, error){
	aImage := *skin
	headImage := imaging.Crop(aImage, image2.Rect(8, 8, 16, 16))
	helmImage := imaging.Crop(aImage, image2.Rect(40, 8, 40+8, 8+8))
	overlayedImage := imaging.Overlay(headImage, helmImage, image2.Pt(0, 0), 1)
	aImage = imaging.Resize(overlayedImage, 200, 200, imaging.NearestNeighbor)
	return &aImage, nil
}

func GetHeadFromUUID(uuid string) (*image2.Image, error) {
	profile, err := GetProfileFromUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("could not find player with that username")
	}
	image, err := GetHeadFromProfile(*profile)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func GetSkinFromUUID(uuid string) (*image2.Image, error){
	profile, err := GetProfileFromUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("could not find player with that username")
	}
	image, err := GetSkinFromProfile(*profile)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func GetImage(imageURL string) (io.ReadCloser, error) {
	request, err := http.NewRequest("GET", imageURL, nil)
	request.Header.Set("User-Agent", "MyMCUUID-API")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
