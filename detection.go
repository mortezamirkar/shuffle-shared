package shuffle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func HandleTenzirHealthUpdate(resp http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		request.Method = "POST"
	}

	type HealthUpdate struct {
		Status string `json:"status"`
	}

	var healthUpdate HealthUpdate
	err := json.NewDecoder(request.Body).Decode(&healthUpdate)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(resp, "Failed to decode JSON: %v", err)
		return
	}
	ctx := context.Background()
	status := healthUpdate.Status

	result, err := GetDisabledRules(ctx)
	if (err != nil && err.Error() == "rules doesn't exist") || err == nil {
		result.IsTenzirActive = status
		result.LastActive = time.Now().Unix()

		err = StoreDisabledRules(ctx, *result)
		if err != nil {
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(`{"success": true}`)))
		return
	}
	resp.WriteHeader(500)
	resp.Write([]byte(`{"success": false}`))
	return
}

func HandleGetDetectionRules(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		orgId, err := fileAuthentication(request)
		if err != nil {
			log.Printf("[WARNING] Bad file authentication in get sigma rules %s: %s", "sigma", err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		user.ActiveOrg.Id = orgId
		user.Username = "Execution File API"
	}

	// Extract detection_type
	location := strings.Split(request.URL.String(), "/")
	if len(location) < 5 {
		log.Printf("[WARNING] Path too short: %d", len(location))
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	detectionType := location[4]
	log.Printf("[AUDIT] User '%s' (%s) is trying to get files from namespace %#v", user.Username, user.Id, detectionType)

	ctx := GetContext(request)
	files, err := GetAllFiles(ctx, user.ActiveOrg.Id, detectionType)
	if err != nil && len(files) == 0 {
		log.Printf("[ERROR] Failed to get files: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Error getting files."}`))
		return
	}

	disabledRules, err := GetDisabledRules(ctx)
	if err != nil && err.Error() != "rules doesn't exist" {
		log.Printf("[ERROR] Failed to get disabled rules: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Error getting disabled rules."}`))
		return
	}

	sort.Slice(files[:], func(i, j int) bool {
		return files[i].UpdatedAt > files[j].UpdatedAt
	})

	type DetectionResponse struct {
		DetectionInfo  []DetectionFileInfo `json:"detection_info"`
		FolderDisabled bool            `json:"folder_disabled"`
		IsConnectorActive bool            `json:"is_connector_active"`
	}

	var sigmaFileInfo []DetectionFileInfo

	for _, file := range files {
		if file.OrgId != user.ActiveOrg.Id {
			continue
		}

		var fileContent []byte

		if file.Encrypted {
			if project.Environment == "cloud" || file.StorageArea == "google_storage" {
				log.Printf("[ERROR] No namespace handler for cloud decryption!")
				//continue
			} else {
				Openfile, err := os.Open(file.DownloadPath)
				if err != nil {
					log.Printf("[ERROR] Failed to open file %s: %s", file.Filename, err)
					continue
				}
				defer Openfile.Close()

				allText := []byte{}
				buf := make([]byte, 1024)
				for {
					n, err := Openfile.Read(buf)
					if err == io.EOF {
						break
					}

					if err != nil {
						log.Printf("[ERROR] Failed to read file %s: %s", file.Filename, err)
						continue
					}

					if n > 0 {
						allText = append(allText, buf[:n]...)
					}
				}

				passphrase := fmt.Sprintf("%s_%s", user.ActiveOrg.Id, file.Id)
				if len(file.ReferenceFileId) > 0 {
					passphrase = fmt.Sprintf("%s_%s", user.ActiveOrg.Id, file.ReferenceFileId)
				}

				decryptedData, err := HandleKeyDecryption(allText, passphrase)
				if err != nil {
					log.Printf("[ERROR] Failed decrypting file %s: %s", file.Filename, err)
					continue
				}

				fileContent = []byte(decryptedData)
			}
		} else {
			fileContent, err = ioutil.ReadFile(file.DownloadPath)
			if err != nil {
				log.Printf("[ERROR] Failed to read file %s: %s", file.Filename, err)
				continue
			}
		}

		var rule DetectionFileInfo
		err = yaml.Unmarshal(fileContent, &rule)
		if err != nil {
			log.Printf("[ERROR] Failed to parse YAML file %s: %s", file.Filename, err)
			continue
		}

		isDisabled := disabledRules.DisabledFolder
		found := false
		if isDisabled {
			rule.IsEnabled = false
		} else {
			for _, disabledFile := range disabledRules.Files {
				if disabledFile.Id == file.Id {
					found = true
					break
				}
			}
			if found {
				rule.IsEnabled = false
			} else {
				rule.IsEnabled = true
			}
		}

		rule.FileId = file.Id
		rule.FileName = file.Filename
		sigmaFileInfo = append(sigmaFileInfo, rule)
	}

	var isTenzirAlive bool
	if time.Now().Unix() > disabledRules.LastActive+10 {
		isTenzirAlive = false
	} else {
		isTenzirAlive = true
	}

	response := DetectionResponse{
		DetectionInfo:      sigmaFileInfo,
		FolderDisabled: disabledRules.DisabledFolder,
		IsConnectorActive: isTenzirAlive,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal response data: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false, "reason": "Error processing rules."}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(responseData)
}

func HandleToggleRule(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	var fileId string
	location := strings.Split(request.URL.String(), "/")
	if location[1] == "api" {
		if len(location) <= 4 {
			log.Printf("Path too short: %d", len(location))
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		fileId = location[5]
	}
	ctx := GetContext(request)

	if len(fileId) != 36 && !strings.HasPrefix(fileId, "file_") {
		log.Printf("[WARNING] Bad format for fileId %s", fileId)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Badly formatted fileId"}`))
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		orgId, err := fileAuthentication(request)
		if err != nil {
			log.Printf("[WARNING] Bad user & file authentication in get for ID %s: %s", fileId, err)
			resp.WriteHeader(401)
			resp.Write([]byte(`{"success": false}`))
			return
		}

		user.ActiveOrg.Id = orgId
		user.Username = "Execution File API"
	}

	file, err := GetFile(ctx, fileId)
	if err != nil {
		log.Printf("[ERROR] File %s not found: %s", fileId, err)
		resp.WriteHeader(400)
		resp.Write([]byte(`{"success": false, "reason": "File not found"}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to delete files: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	var action string
	switch location[6] {
	case "disable_rule":
		action = "disable"
	case "enable_rule":
		action = "enable"
	default:
		log.Printf("[WARNING] path not found: %s", location[6])
		resp.WriteHeader(404)
		resp.Write([]byte(`{"success": false, "message": "The URL doesn't exist or is not allowed."}`))
		return
	}

	if action == "disable" {
		err := disableRule(*file)
		if err != nil {
			log.Printf("[ERROR] Failed to %s file", action)
			resp.WriteHeader(500)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	} else if action == "enable" {
		err := enableRule(*file)
		if err != nil {
			if err.Error() != "rules doesn't exist" {
				log.Printf("[ERROR] Failed to %s file, reason: %s", action, err)
				resp.WriteHeader(404)
				resp.Write([]byte(`{"success": false}`))
				return
			} else {
				log.Printf("[ERROR] Failed to %s file, reason: %s", action, err)
				resp.WriteHeader(500)
				resp.Write([]byte(`{"success": false}`))
				return
			}
		}
	}

	var execType string

	if action == "disable" {
		execType = "DISABLE_SIGMA_FILE"
	} else if action == "enable" {
		execType = "ENABLE_SIGMA_FILE"
	}

	err = SetExecRequest(ctx, execType, file.Filename)
	if err != nil {
		log.Printf("[ERROR] Failed setting workflow queue for env: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte((`{"success": true}`)))
}

func HandleFolderToggle(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	location := strings.Split(request.URL.String(), "/")
	if location[1] != "api" || len(location) < 6 {
		log.Printf("Path too short or incorrect: %s", request.URL.String())
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	ctx := GetContext(request)
	action := location[5]

	rules, err := GetDisabledRules(ctx)
	if err != nil {
		log.Printf("[WARNING] Cannot get the rules, reason %s", err)
		resp.WriteHeader(404)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if action == "disable_folder" {
		rules.DisabledFolder = true
	} else if action == "enable_folder" {
		rules.DisabledFolder = false
	} else {
		log.Printf("[WARNING] path not found: %s", action)
		resp.WriteHeader(404)
		resp.Write([]byte(`{"success": false, "message": "The URL doesn't exist or is not allowed."}`))
		return
	}

	err = StoreDisabledRules(ctx, *rules)
	if err != nil {
		log.Printf("[ERROR] Failed to store disabled rules: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var execType string
	if action == "disable_folder" {
		execType = "DISABLE_SIGMA_FOLDER"
	} else {
		execType = "CATEGORY_UPDATE"
	}

	err = SetExecRequest(ctx, execType, "")
	if err != nil {
		log.Printf("[ERROR] Failed setting workflow queue for env: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write([]byte(`{"success": true}`))
}

func disableRule(file File) error {
	ctx := context.Background()
	resp, err := GetDisabledRules(ctx)
	if err != nil {
		if err.Error() == "rules doesn't exist" {
			// FIX ME :- code duplication : (
			disabRules := &DisabledRules{}
			disabRules.Files = append(disabRules.Files, file)
			err = StoreDisabledRules(ctx, *disabRules)
			if err != nil {
				return err
			}

			log.Printf("[INFO] file with ID %s is disabled successfully", file.Id)
			return nil
		} else {
			return err
		}
	}

	resp.Files = append(resp.Files, file)
	err = StoreDisabledRules(ctx, *resp)
	if err != nil {
		return err
	}

	log.Printf("[INFO] file with ID %s is disabled successfully", file.Id)
	return nil
}

func enableRule(file File) error {
	ctx := context.Background()
	resp, err := GetDisabledRules(ctx)
	if err != nil {
		return err
	}

	// Check if resp.Files is empty
	if len(resp.Files) == 0 {
		log.Printf("[INFO] No disabled rules found.")
		return nil
	}

	found := false
	for i, innerFile := range resp.Files {
		if innerFile.Id == file.Id {
			resp.Files = append(resp.Files[:i], resp.Files[i+1:]...)
			found = true
			break 
		}
	}

	if !found {
		log.Printf("[INFO] File with ID %s not found in disabled rules", file.Id)
		return nil
	}

	err = StoreDisabledRules(ctx, *resp)
	if err != nil {
		return err
	}

	log.Printf("[INFO] File with ID %s is enabled successfully", file.Id)
	return nil
}

func HandleGetSelectedRules(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}
	_, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in get env stats executions: %s", err)
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	var triggerId string
	location := strings.Split(request.URL.String(), "/")
	if len(location) < 5 || location[1] != "api" {
		log.Printf("[INFO] Path too short or incorrect: %d", len(location))
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	triggerId = location[4]

	selectedRules, err := GetSelectedRules(request.Context(), triggerId)
	if err != nil {
		if err.Error() != "rules doesnt exists" {
			log.Printf("[ERROR] Error getting selected rules for %s: %s", triggerId, err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(`{"success": false}`))
			return
		}
	}

	responseData, err := json.Marshal(selectedRules)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal response data: %s", err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success": false"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(responseData)
}

func HandleSaveSelectedRules(resp http.ResponseWriter, request *http.Request) {
	cors := HandleCors(resp, request)
	if cors {
		return
	}

	user, err := HandleApiAuthentication(resp, request)
	if err != nil {
		log.Printf("[WARNING] Api authentication failed in save selected rules: %s", err)
		resp.WriteHeader(http.StatusUnauthorized)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	if user.Role == "org-reader" {
		log.Printf("[WARNING] Org-reader doesn't have access to save rules: %s (%s)", user.Username, user.Id)
		resp.WriteHeader(http.StatusForbidden)
		resp.Write([]byte(`{"success": false, "reason": "Read only user"}`))
		return
	}

	location := strings.Split(request.URL.String(), "/")
	if len(location) < 5 || location[1] != "api" {
		log.Printf("[INFO] Path too short or incorrect: %d", len(location))
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	triggerId := location[4]

	selectedRules := SelectedDetectionRules{}
	
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&selectedRules)
	if err != nil {
		log.Printf("[ERROR] Failed to decode request body: %s", err)
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(`{"success": false, "reason": "Invalid request body"}`))
		return
	}

	err = StoreSelectedRules(request.Context(), triggerId, selectedRules)
	if err != nil {
			log.Printf("[ERROR] Error storing selected rules for %s: %s", triggerId, err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte(`{"success": false}`))
			return
	}

	responseData, err := json.Marshal(selectedRules)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal response data: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(`{"success": false}`))
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(responseData)
}